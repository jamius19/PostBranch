package repo

import (
	"context"
	"fmt"
	db2 "github.com/jamius19/postbranch/internal/db"
	model2 "github.com/jamius19/postbranch/internal/db/gen/model"
	repoDto "github.com/jamius19/postbranch/internal/dto/repo"
	"github.com/jamius19/postbranch/internal/logger"
	"github.com/jamius19/postbranch/internal/runner"
	pgSvc "github.com/jamius19/postbranch/internal/service/pg"
	zfs2 "github.com/jamius19/postbranch/internal/service/zfs"
	"github.com/jamius19/postbranch/web/responseerror"
	"os"
)

var log = logger.Logger

func InitializeRepo(ctx context.Context, repoInit repoDto.Info) (model2.Repo, model2.ZfsPool, error) {
	if repoInit.GetRepoType() == "virtual" {
		log.Infof("Initializing virtual repo")

		pool, err := zfs2.VirtualPool(ctx, repoInit)
		if err != nil {
			return model2.Repo{}, model2.ZfsPool{}, err
		}

		log.Infof("Initialized virtual pool. PoolInfo: %v", pool)

		repoInfo := model2.Repo{
			Name:   repoInit.GetName(),
			PoolID: *pool.ID,
		}

		createdRepo, err := db2.CreateRepo(ctx, repoInfo)
		if err != nil {
			// TODO: Cleanup Pool and Dataset
			log.Infof("Failed to insert repo. Name: %s Data: %v Error: %s", repoInit.GetName(), repoInfo, err)
			return model2.Repo{}, model2.ZfsPool{}, responseerror.From("Failed to create repository")
		}

		return createdRepo, pool, nil
	}

	return model2.Repo{}, model2.ZfsPool{}, fmt.Errorf("not implemented yet")
}

func DeleteRepo(ctx context.Context, repoDetail db2.RepoDetail) error {
	log.Infof("Deleting repo: %s, pool: %s", repoDetail.Repo.Name, repoDetail.Pool.Path)
	pool := repoDetail.Pool

	for _, branch := range repoDetail.Branches {
		err := pgSvc.StopPg(
			repoDetail.Pg.PgPath,
			pool.MountPath,
			branch.Name,
			false,
		)

		if err != nil {
			return err
		}
	}

	loopbackPath, err := zfs2.FindDevicePath(pool.Name)
	if err != nil {
		return err
	}

	_, err = runner.Single(
		"zpool-destroy", false, false, "zpool", "destroy", "-f", pool.Name,
	)

	if err != nil {
		return fmt.Errorf("failed to destroy pool: %s", err)
	}

	if pool.PoolType == "virtual" {
		if err := zfs2.ReleaseLoopDevice(loopbackPath); err != nil {
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

	err = db2.DeletePool(ctx, *pool.ID)
	if err != nil {
		return err
	}

	return nil
}
