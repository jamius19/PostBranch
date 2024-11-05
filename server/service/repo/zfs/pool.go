package zfs

import (
	"context"
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"strings"
)

var log = logger.Logger

func VirtualPool(ctx context.Context, repoinit *repo.InitDto) (*dao.ZfsPool, error) {
	log.Infof("ZFS Pool init %v", *repoinit)

	_, path := util.SplitPath(repoinit.Path)
	log.Infof("Creating virtual disk file at %s", repoinit.Path)

	//
	// Set the Commands
	//
	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()
	cmds.Set("create-folder", cmd.Get("mkdir", "-p", path))

	filesize := fmt.Sprintf("%dM", repoinit.SizeInMb)
	cmds.Set("create-file", cmd.Get("fallocate", "-l", filesize, repoinit.Path))

	loopNo, err := findFreeLoopDevice()
	if err != nil {
		log.Errorf("Failed to find free loop device: %s", err)
		return nil, responseerror.Clarify("Failed to find free loop device")
	}

	loopDevice := fmt.Sprintf("/dev/loop%d", loopNo)

	cmds.Set("create-loop-node", cmd.Get("mknod", loopDevice, "b", "7", fmt.Sprintf("%d", loopNo)))
	cmds.Set("setup-loopback", cmd.Get("losetup", loopDevice, repoinit.Path))

	//
	// Run the Commands
	//
	_, err = cmd.Multi(cmds)
	if err != nil {
		log.Errorf("Failed to create virtual disk: %s", err)
		return nil, responseerror.Clarify("Error creating virtual disk")
	}

	pool, err := createPool(ctx, repoinit, loopDevice)
	if err != nil {
		log.Errorf("Failed to create createPool: %s", err)
		return nil, err
	}

	return pool, nil
}

func createPool(ctx context.Context, repoinit *repo.InitDto, devicePath string) (*dao.ZfsPool, error) {
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

	// failedPoolIds will contain the IDs of the pools that failed to mount
	var failedPoolIds []int64

	for _, pool := range pools {
		if pool.PoolType == "virtual" {
			err := setupLoopback(&pool)
			if err != nil {
				failedPoolIds = append(failedPoolIds, pool.ID)
				log.Errorf("Failed to setup loopback for pool %v: %s", pool, err)
			}
		}
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
	if pool.PoolType != "virtual" {
		log.Infof("Pool is not virtual, skipping loopback setup. pool %v", pool)
		return nil
	}

	log.Infof("Unmounting in case it's already mounted. pool %v", pool)
	_, err := cmd.Single("zpool-export", false, false, "zpool", "export", pool.Name)

	if err != nil {
		log.Infof("[IGNORE] Failed to export pool: %s", err)
	}

	log.Infof("Detaching any dangling loop devices. pool %v", pool)
	loopbackOutput, err := cmd.Single(
		"find-loopback-"+pool.Name,
		false,
		false,
		"su", "-c",
		fmt.Sprintf(FindLoopBackCmd, pool.Path),
	)

	if output := util.TrimmedString(loopbackOutput); err == nil && output != cmd.EmptyOutput {
		log.Infof("Loopback found for pool %v, output: %s", pool, output)
		devices := strings.Split(output, "\n")

		for _, device := range devices {
			log.Infof("Detaching loopback device %s", device)

			cmds := orderedmap.NewOrderedMap[string, cmd.Command]()
			cmds.Set("detach-loopback", cmd.Get("losetup", "-d", device))
			cmds.Set("remove-loopback", cmd.Get("rm", device))
			_, err = cmd.Multi(cmds)
			if err != nil {
				log.Errorf("Failed to remove loopback device %s: %s", device, err)
				return err
			}
		}
	} else {
		log.Infof("[IGNORE] Failed to find loopback for pool %v, output: %s, error: %s", pool, output, err)
	}

	log.Infof("Setting up loopbacks for pool %v", pool)
	loopNo, err := findFreeLoopDevice()
	if err != nil {
		log.Errorf("Failed to find free loop device: %s", err)
		return err
	}

	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()
	loopDevice := fmt.Sprintf("/dev/loop%d", loopNo)
	cmds.Set("create-loop-node", cmd.Get("mknod", loopDevice, "b", "7", fmt.Sprintf("%d", loopNo)))
	cmds.Set("setup-loopback", cmd.Get("losetup", loopDevice, pool.Path))

	_, err = cmd.Multi(cmds)
	if err != nil {
		log.Errorf("Failed to setup loopback for pool, error %s", err)
		return err
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
		err := Unmount(&pool)
		if err != nil {
			log.Errorf("Failed to unmount pool: %v, error: %s", pool, err)
			return err
		}
	}

	return nil
}

func Unmount(pool *dao.ZfsPool) error {
	log.Infof("Unmounting pool %v", pool)
	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()

	// TODO: Stop running postgres

	cmds.Set(
		"zpool-export",
		cmd.Get("zpool", "export", pool.Name),
	)

	if pool.PoolType == "virtual" {
		loopbackPath, err := FindDevicePath(pool)
		if err != nil {
			return err
		}

		loopbackPath = "/dev/" + loopbackPath

		cmds.Set("loopback-detach", cmd.Get("losetup", "-d", loopbackPath))
		cmds.Set("remove-device", cmd.Get("rm", "-rf", loopbackPath))
	}

	cmds.Set("remove-mount-path", cmd.Get("rm", "-rf", pool.MountPath))

	_, err := cmd.Multi(cmds)
	if err != nil {
		return err
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

	return util.TrimmedString(devicePath), nil
}
