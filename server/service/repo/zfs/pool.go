package zfs

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"os"
)

var log = logger.Logger

func VirtualPool(ctx context.Context, repoinit *repo.InitDto) (*dao.ZfsPool, error) {
	log.Infof("ZFS Pool init %v", *repoinit)

	if err := CreateSparseFile(repoinit.Path, repoinit.SizeInMb); err != nil {
		log.Errorf("Failed to create sparse file. Error: %s", err)
		return nil, responseerror.Clarify("Failed to create sparse file")
	}

	loopNo, err := SetupLoopDevice(repoinit.Path)
	if err != nil {
		log.Errorf("Failed to setup loopback device. Error: %s", err)
		return nil, responseerror.Clarify("Failed to setup loopback device")
	}

	pool, err := createPool(ctx, repoinit, loopNo)
	if err != nil {
		log.Errorf("Failed to create createPool: %s", err)
		return nil, err
	}

	return pool, nil
}

func createPool(ctx context.Context, repoinit *repo.InitDto, loopNo int) (*dao.ZfsPool, error) {
	devicePath := fmt.Sprintf("/dev/loop%d", loopNo)
	mountPath := fmt.Sprintf("/mnt/pb-%s", repoinit.Name)

	_, err := cmd.Single(
		"create-zpool",
		false,
		false,
		"zpool",
		"create", "-m",
		mountPath,
		repoinit.Name,
		devicePath,
	)

	if err != nil {
		log.Errorf("Failed to create createPool: %s", err)
		return nil, err
	}

	// Create a new Pool
	poolData := dao.CreatePoolParams{
		Name:      repoinit.Name,
		Path:      repoinit.Path,
		SizeInMb:  repoinit.SizeInMb,
		MountPath: mountPath,
		PoolType:  repoinit.RepoType,
	}

	pool, err := data.Fetcher.CreatePool(ctx, poolData)
	if err != nil {
		// TODO: Cleanup Pool
		log.Errorf("Failed to insert createPool. Repo:%v Path: %s Error:%s", poolData, devicePath, err)
		return nil, responseerror.Clarify("Failed to create createPool")
	}

	log.Infof("Pool insertion successful %s", pool.Name)

	return &pool, nil
}

func MountAll() error {
	pools, err := data.Fetcher.ListPool(context.Background())
	if err != nil {
		log.Errorf("Failed to list pools: %s", err)
		return err
	}

	if len(pools) == 0 {
		log.Info("No pools to mount")
		return nil
	}

	log.Infof("Mounting all pools")

	// failedPools will contain the list of the pool(s) for which loopback device(s) failed to mount
	var failedPools []dao.ZfsPool

	for _, pool := range pools {
		if pool.PoolType == "virtual" {
			if err := setupLoopback(&pool); err != nil {
				failedPools = append(failedPools, pool)
				log.Errorf("Failed to setup loopback for pool %v: %s", pool, err)
			}

		} else {
			log.Infof("Pool is not virtual, skipping loopback setup. pool %v", pool)
		}
	}

	if len(failedPools) > 0 {
		log.Errorf("Failed to setup loopback for the following pools: %v", failedPools)
	}

	log.Infof("**** This is a time consuming operation. Please wait. ****")

	output, err := cmd.Single("import-zpools", false, false, "zpool", "import", "-a")
	if err != nil {
		log.Errorf("Failed to import zpools: %s, output: %s", err, util.SafeStringVal(output))
		return err
	}

	log.Infof("**** Done! Thank you for your patience! :) ****")
	log.Infof("%d pool(s) are mounted.", len(pools))

	// TODO: Start postgres

	return nil
}

func setupLoopback(pool *dao.ZfsPool) error {
	err := cleanDanglingLoopbackDevices(pool)
	if err != nil {
		return err
	}

	log.Infof("Setting up loopbacks for pool %v", pool)
	if _, err := SetupLoopDevice(pool.Path); err != nil {
		return err
	}

	return nil
}

func cleanDanglingLoopbackDevices(pool *dao.ZfsPool) error {
	log.Infof("Unmounting in case it's already mounted. pool %v", pool)
	_, err := cmd.Single("zpool-export", true, false, "zpool", "export", pool.Name)

	if err != nil {
		log.Infof("Pool is not mounted. Continuing. pool: %v", pool)
	} else {
		log.Warnf("Pool is already mounted. Unmounting it. pool: %v", pool)
	}

	devices, err := FindLoopDeviceFromSys(pool.Path)
	if err != nil {
		log.Errorf("Failed to find loopback for pool %v, error: %s", pool, err)
		return err
	}

	if len(devices) > 0 {
		log.Warnf("Dangling loopback devices found for pool %v, devices: %v", pool, devices)
		log.Warnf("Releasing dangling loopback devices for pool %v", pool)
	}

	for _, device := range devices {
		if err := ReleaseLoopDevice(device); err != nil {
			return fmt.Errorf("failed to release loopback device: %s", err)
		}

		if err := os.Remove(device); err != nil {
			log.Errorf("Failed to remove loopback device %s: %s", device, err)
			return err
		}
	}
	return nil
}

func UnmountAll() error {
	log.Infof("Unmounting all pools")
	pools, err := data.Fetcher.ListPool(context.Background())
	if err != nil {
		log.Errorf("Failed to list pools: %s", err)
		return err
	}

	if len(pools) == 0 {
		log.Info("No pools to unmount")
		return nil
	}

	for _, pool := range pools {
		if err := Unmount(&pool); err != nil {
			log.Errorf("Failed to unmount pool: %v, error: %s", pool, err)
			return err
		}
	}

	return nil
}

func Unmount(pool *dao.ZfsPool) error {
	log.Infof("Unmounting pool %v", pool)

	// TODO: Stop running postgres

	loopbackPath, err := FindDevicePath(pool)
	if err != nil {
		return err
	}

	_, err = cmd.Single(
		"zpool-export",
		false,
		false,
		"zpool", "export", pool.Name,
	)

	if err != nil {
		return err
	}

	if pool.PoolType == "virtual" {
		if err := ReleaseLoopDevice(loopbackPath); err != nil {
			return fmt.Errorf("failed to release loopback device: %s", err)
		}

		if err := os.Remove(loopbackPath); err != nil {
			return fmt.Errorf("failed to remove loopback device: %w", err)
		}
	}

	if err := os.RemoveAll(pool.MountPath); err != nil {
		return fmt.Errorf("failed to remove mount path: %w", err)
	}

	return nil
}

func FindDevicePath(pool *dao.ZfsPool) (string, error) {
	devicePath, err := cmd.Single(
		"find-zpool-device",
		false,
		false,
		"su", "-c",
		fmt.Sprintf(FindLoopBackFromZpoolCmd, pool.Name),
	)

	if err != nil {
		log.Errorf("Failed to get path path for pool: %v, error: %s", pool, err)
		return "", err
	}

	return "/dev/" + util.TrimmedString(devicePath), nil
}
