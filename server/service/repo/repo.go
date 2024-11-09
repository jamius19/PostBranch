package repo

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/repo/zfs"
	"github.com/jamius19/postbranch/web/responseerror"
	"os"
)

var log = logger.Logger

func InitializeRepo(ctx context.Context, repoinit *repo.InitDto) (*dao.Repo, *dao.ZfsPool, error) {
	if repoinit.RepoType == "virtual" {
		log.Infof("Initializing virtual repo")

		pool, err := zfs.VirtualPool(ctx, repoinit)
		if err != nil {
			return nil, nil, err
		}

		log.Infof("Initialized virtual pool. PoolInfo: %v", pool)

		_, err = zfs.EmptyDataset(ctx, pool, "main")
		if err != nil {
			return nil, nil, err
		}

		repoCreateDto := dao.CreateRepoParams{
			Name:   repoinit.Name,
			PoolID: pool.ID,
		}

		createdRepo, err := data.Db.CreateRepo(ctx, repoCreateDto)
		if err != nil {
			// TODO: Cleanup Pool and Dataset
			log.Infof("Failed to insert repo. Name: %s Data: %v Error: %s", repoinit.Name, repoCreateDto, err)
			return nil, nil, responseerror.From("Failed to create repository")
		}

		return &createdRepo, pool, nil
	}

	return nil, nil, fmt.Errorf("not implemented yet")
}

func DeleteRepo(ctx context.Context, repo *dao.Repo, pool *dao.ZfsPool) error {
	log.Infof("Deleting repo: %v, pool: %v", repo, pool)

	// TODO: Stop postgres

	loopbackPath, err := zfs.FindDevicePath(pool)
	if err != nil {
		return err
	}

	_, err = cmd.Single(
		"zpool-destroy", false, false, "zpool", "destroy", "-f", pool.Name,
	)

	if err != nil {
		return fmt.Errorf("failed to destroy pool: %s", err)
	}

	if pool.PoolType == "virtual" {
		if err := zfs.ReleaseLoopDevice(loopbackPath); err != nil {
			return fmt.Errorf("failed to release loopback device: %s", err)
		}

		if err := os.Remove(loopbackPath); err != nil {
			return fmt.Errorf("failed to remove loopback device: %w", err)
		}

		if err := os.Remove(pool.Path); err != nil {
			return fmt.Errorf("failed to remove pool image: %w", err)
		}
	}

	if err := os.RemoveAll(pool.MountPath); err != nil {
		return fmt.Errorf("failed to remove mount path: %w", err)
	}

	err = data.Db.DeletePool(ctx, pool.ID)
	if err != nil {
		return err
	}

	return nil
}
