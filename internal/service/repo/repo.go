package repo

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/internal/db"
	"github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/dto/pg"
	repoDto "github.com/jamius19/postbranch/internal/dto/repo"
	"github.com/jamius19/postbranch/internal/logger"
	"github.com/jamius19/postbranch/internal/runner"
	pgSvc "github.com/jamius19/postbranch/internal/service/pg"
	"github.com/jamius19/postbranch/internal/service/zfs"
	"github.com/jamius19/postbranch/web/responseerror"
	"os"
)

var log = logger.Logger

func InitializeRepo(ctx context.Context, repoInit repoDto.Info, pgInfo pg.Info) (model.Repo, model.ZfsPool, error) {
	if repoInit.GetRepoType() == "virtual" {
		log.Infof("Initializing virtual repo")

		pool, err := zfs.VirtualPool(ctx, repoInit)
		if err != nil {
			return model.Repo{}, model.ZfsPool{}, err
		}

		log.Infof("Initialized virtual pool. PoolInfo: %v", pool)

		repoInfo := model.Repo{
			Name:    repoInit.GetName(),
			PoolID:  *pool.ID,
			PgPath:  pgInfo.GetPgPath(),
			Version: pgInfo.GetVersion(),
			Status:  string(db.RepoStarted),
			Adapter: string(pgInfo.GetAdapter()),
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

func DeleteRepo(ctx context.Context, repoDetail db.RepoDetail) error {
	log.Infof("Deleting repo: %s, pool: %s", repoDetail.Repo.Name, repoDetail.Pool.Path)
	pool := repoDetail.Pool

	if _, err := os.Stat(pool.Path); os.IsNotExist(err) {
		log.Warnf("Pool file does not exist for pool %v: %s", pool, err)

		log.Infof("Trying to stop pg instances, expect failure")
		for _, branch := range repoDetail.Branches {
			_ = pgSvc.StopPg(
				repoDetail.Repo.PgPath,
				pool.MountPath,
				branch.Name,
				false,
			)
		}

		log.Infof("Trying to destroy pool %s anyway, expect failure", pool.Name)
		_, err = runner.Single(
			"zpool-destroy", false, false, "zpool", "destroy", "-f", pool.Name,
		)

		if err := os.RemoveAll(pool.MountPath); err != nil {
			return fmt.Errorf("failed to remove mount path: %w", err)
		}

		err = db.DeletePool(ctx, *pool.ID)
		if err != nil {
			return err
		}

		return nil
	}

	for _, branch := range repoDetail.Branches {
		err := pgSvc.StopPg(
			repoDetail.Repo.PgPath,
			pool.MountPath,
			branch.Name,
			false,
		)

		if err != nil {
			return err
		}
	}

	loopbackPath, err := zfs.FindDevicePath(pool.Name)
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
