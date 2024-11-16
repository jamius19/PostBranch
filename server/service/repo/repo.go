package repo

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/db/gen/model"
	repoDto "github.com/jamius19/postbranch/dto/repo"
	"github.com/jamius19/postbranch/logger"
	pgSvc "github.com/jamius19/postbranch/service/pg"
	"github.com/jamius19/postbranch/service/zfs"
	"github.com/jamius19/postbranch/web/responseerror"
	"os"
)

var log = logger.Logger

func InitializeRepo(ctx context.Context, repoInit repoDto.Info) (model.Repo, model.ZfsPool, error) {
	if repoInit.GetRepoType() == "virtual" {
		log.Infof("Initializing virtual repo")

		pool, err := zfs.VirtualPool(ctx, repoInit)
		if err != nil {
			return model.Repo{}, model.ZfsPool{}, err
		}

		log.Infof("Initialized virtual pool. PoolInfo: %v", pool)

		_, err = zfs.EmptyDataset(ctx, pool, "main")
		if err != nil {
			return model.Repo{}, model.ZfsPool{}, err
		}

		repoInfo := model.Repo{
			Name:   repoInit.GetName(),
			PoolID: *pool.ID,
		}

		createdRepo, err := db.CreateRepo(ctx, repoInfo)
		if err != nil {
			// TODO: Cleanup Pool and Dataset
			log.Infof("Failed to insert repo. Name: %s Data: %v Error: %s", repoInit.GetName(), repoInfo, err)
			return model.Repo{}, model.ZfsPool{}, responseerror.From("Failed to create repository")
		}

		return createdRepo, pool, nil
	}

	return model.Repo{}, model.ZfsPool{}, fmt.Errorf("not implemented yet")
}

func DeleteRepo(ctx context.Context, repo model.Repo, pool model.ZfsPool, pg model.Pg) error {
	log.Infof("Deleting repo: %v, pool: %v", repo, pool)

	// TODO: Stop postgres
	datasets, err := db.ListDatasetByNameAndPoolId(ctx, *pool.ID)
	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		if err := pgSvc.StopPg(pg.PgPath, pool.MountPath, dataset.Name, false); err != nil {
			return err
		}
	}
	//pgSvc.StopPg(pg.PgPath, pool.MountPath)

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

	err = db.DeletePool(ctx, *pool.ID)
	if err != nil {
		return err
	}

	return nil
}
