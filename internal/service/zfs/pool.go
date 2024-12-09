package zfs

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/internal/db"
	"github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/dto/repo"
	"github.com/jamius19/postbranch/internal/logger"
	"github.com/jamius19/postbranch/internal/runner"
	"github.com/jamius19/postbranch/internal/service/pg"
	"github.com/jamius19/postbranch/web/responseerror"
	"os"
	"slices"
	"strings"
	"sync"
)

var log = logger.Logger

func VirtualPool(ctx context.Context, repoinit repo.Info) (model.ZfsPool, error) {
	log.Infof("ZFS Pool init %v", repoinit)

	if err := CreateSparseFile(repoinit.GetPath(), repoinit.GetSizeInMb()); err != nil {
		log.Errorf("Failed to create sparse file. Error: %s", err)
		return model.ZfsPool{}, responseerror.From("Failed to create sparse file")
	}

	loopNo, err := SetupLoopDevice(repoinit.GetPath())
	if err != nil {
		log.Errorf("Failed to setup loopback device. Error: %s", err)
		return model.ZfsPool{}, responseerror.From("Failed to setup loopback device")
	}

	pool, err := createPool(ctx, repoinit, loopNo)
	if err != nil {
		log.Errorf("Failed to create createPool: %s", err)
		return model.ZfsPool{}, err
	}

	return pool, nil
}

func createPool(ctx context.Context, repoinit repo.Info, loopNo int) (model.ZfsPool, error) {
	devicePath := fmt.Sprintf("/dev/loop%d", loopNo)
	mountPath := fmt.Sprintf("/mnt/pb-%s", repoinit.GetName())

	_, err := runner.Single(
		"create-zpool",
		false,
		false,
		"zpool",
		"create", "-m",
		mountPath,
		repoinit.GetName(),
		devicePath,
	)

	if err != nil {
		log.Errorf("Failed to create createPool: %s", err)
		return model.ZfsPool{}, err
	}

	// Create a new Pool
	poolData := model.ZfsPool{
		Name:      repoinit.GetName(),
		Path:      repoinit.GetPath(),
		SizeInMb:  repoinit.GetSizeInMb(),
		MountPath: mountPath,
		PoolType:  repoinit.GetRepoType(),
	}

	pool, err := db.CreatePool(ctx, poolData)
	if err != nil {
		// TODO: Cleanup Pool
		log.Errorf("Failed to insert createPool. Repo:%v Path: %s Error:%s", poolData, devicePath, err)
		return model.ZfsPool{}, responseerror.From("Failed to create createPool")
	}

	log.Infof("Pool insertion successful %s", pool.Name)

	return pool, nil
}

func MountAll(ctx context.Context) error {
	repoDetails, err := db.ListRepo(ctx)
	if err != nil {
		log.Errorf("Failed to list pools: %s", err)
		return err
	}

	if len(repoDetails) == 0 {
		log.Info("No pools to mount")
		return nil
	}

	log.Infof("Mounting all pools")

	// failedPools will contain the list of the pool(s) for which loopback device(s) failed to mount
	var failedPools []string
	var poolWg sync.WaitGroup

	log.Infof("Stopping potential dangling postgres instances")
	for _, repoDetail := range repoDetails {
		for _, branch := range repoDetail.Branches {
			if branch.Status == string(db.BranchClosed) {
				continue
			}

			poolWg.Add(1)

			go pg.StopDangingPg(
				repoDetail.Repo.PgPath,
				repoDetail.Pool.MountPath,
				branch.Name,
				&poolWg,
			)
		}
	}

	poolWg.Wait()

	for _, repoDetail := range repoDetails {
		pool := repoDetail.Pool

		if _, err := os.Stat(pool.Path); os.IsNotExist(err) {
			failedPools = append(failedPools, pool.Name)

			_, err := db.UpdateRepoStatus(ctx, *repoDetail.Repo.ID, db.RepoFailed, fmt.Sprintf("Pool file %s does not exist", pool.Path))
			if err != nil {
				log.Errorf("Failed to update repo status: %s", err)
			}

			log.Errorf("Pool file does not exist for pool %v: %s", pool, err)
			continue
		}

		if pool.PoolType == "virtual" {
			if err := setupLoopback(&pool); err != nil {
				failedPools = append(failedPools, pool.Name)
				log.Errorf("Failed to setup loopback for pool %v: %s", pool, err)
			}
		} else {
			log.Infof("Pool is not virtual, skipping loopback setup. pool %v", repoDetail.Pool)
		}

		_, err := db.UpdateRepoStatus(ctx, *repoDetail.Repo.ID, db.RepoCompleted, "")
		if err != nil {
			log.Errorf("Failed to update repo status: %s", err)
		}
	}

	if len(failedPools) > 0 {
		log.Errorf("Failed to setup loopback for the following pools: %v", failedPools)
	}

	if len(failedPools) < len(repoDetails) {
		log.Infof("**** Importing all pools. This is a time consuming operation. Please wait. ****")

		output, err := runner.Single("import-zpools", false, false, "zpool", "import", "-a")
		if err != nil {
			log.Errorf("Failed to import zpools: %s, output: %s", err, output)
			return err
		}

		log.Infof("%d pool(s) are mounted.", len(repoDetails))
	}

	select {
	case <-ctx.Done():
		log.Infof("Root Context cancelled. Skipping database start")
		return nil
	default:
	}

	log.Infof("**** Importing all databases. Please wait. ****")

	for _, repoDetail := range repoDetails {
		if slices.Contains(failedPools, repoDetail.Repo.Name) {
			continue
		}

		for _, branch := range repoDetail.Branches {
			if branch.Status == string(db.BranchClosed) {
				continue
			}

			poolWg.Add(1)

			go pg.StartPgAndUpdateBranch(
				ctx,
				repoDetail.Repo.PgPath,
				repoDetail.Pool.MountPath,
				branch.Name,
				*branch.ID,
				&poolWg,
			)
		}
	}

	log.Infof("Waiting for all databases to start")
	poolWg.Wait()

	log.Infof("**** All Done! Thank you for your patience! :) ****")

	return nil
}

func setupLoopback(pool *model.ZfsPool) error {
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

func cleanDanglingLoopbackDevices(pool *model.ZfsPool) error {
	log.Infof("Unmounting in case it's already mounted. pool %v", pool)
	_, err := runner.Single("zpool-export", true, false, "zpool", "export", pool.Name)

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
	repoDetails, err := db.ListRepoWithStatus(context.Background(), db.RepoCompleted)

	if err != nil {
		log.Errorf("Failed to list repoDetails: %s", err)
		return err
	}

	if len(repoDetails) == 0 {
		log.Info("No pools to unmount")
		return nil
	}

	log.Infof("Stopping all databases")
	var poolWg sync.WaitGroup
	ctx := context.Background()

	for _, repoDetail := range repoDetails {
		for _, branch := range repoDetail.Branches {
			if branch.Status == string(db.BranchClosed) {
				continue
			}

			poolWg.Add(1)

			go pg.StopPgAndUpdateBranch(
				ctx,
				repoDetail.Repo.PgPath,
				repoDetail.Pool.MountPath,
				branch.Name,
				*branch.ID,
				&poolWg,
			)
		}
	}

	log.Infof("Waiting for all databases to stop")
	poolWg.Wait()
	log.Infof("All databases are stopped")

	for _, repoDetail := range repoDetails {
		if err := Unmount(repoDetail.Pool); err != nil {
			log.Errorf("Failed to unmount pool: %v, error: %s", repoDetail.Pool, err)
			return err
		}
	}

	return nil
}

func Unmount(pool model.ZfsPool) error {
	log.Infof("Unmounting pool %v", pool)

	loopbackPath, err := FindDevicePath(pool.Name)
	if err != nil {
		return err
	}

	_, err = runner.Single(
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

func FindDevicePath(poolName string) (string, error) {
	devicePath, err := runner.Single(
		"find-zpool-device",
		false,
		false,
		"su", "-c",
		fmt.Sprintf(FindLoopBackFromZpoolCmd, poolName),
	)

	if err != nil {
		log.Errorf("Failed to get path path for pool: %s, error: %s", poolName, err)
		return "", err
	}

	return "/dev/" + strings.TrimSpace(devicePath), nil
}
