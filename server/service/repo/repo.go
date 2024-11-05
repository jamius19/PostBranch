package repo

import (
	"context"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/repo/zfs"
	"github.com/jamius19/postbranch/web/responseerror"
)

var log = logger.Logger

func InitializeRepo(ctx context.Context, repoinit *repo.InitDto) (*repo.Response, error) {
	if repoinit.RepoType == "virtual" {
		log.Infof("Initializing virtual repo")

		pool, err := zfs.VirtualPool(ctx, repoinit)
		if err != nil {
			return nil, err
		}

		log.Infof("Initialized virtual pool. PoolInfo: %v", pool)

		_, err = zfs.EmptyDataset(ctx, pool, "main")
		if err != nil {
			return nil, err
		}

		repoCreateDto := dao.CreateRepoParams{
			Name:   repoinit.Name,
			PoolID: pool.ID,
		}

		createdRepo, err := data.Fetcher.CreateRepo(ctx, repoCreateDto)
		if err != nil {
			// TODO: Cleanup Pool and Dataset
			log.Infof("Failed to insert repo. Name: %s Data: %v Error: %s", repoinit.Name, repoCreateDto, err)
			return nil, responseerror.Clarify("Failed to create repository")
		}

		repoResponse := repo.Response{
			ID:        createdRepo.ID,
			Name:      createdRepo.Name,
			Path:      pool.Path,
			RepoType:  pool.PoolType,
			SizeInMb:  pool.SizeInMb,
			Pg:        nil,
			PoolID:    pool.ID,
			CreatedAt: createdRepo.CreatedAt,
			UpdatedAt: createdRepo.UpdatedAt,
		}

		return &repoResponse, nil
	}

	return nil, nil
}

func DeleteRepo(ctx context.Context, repo *dao.Repo, pool *dao.ZfsPool) error {
	log.Infof("Deleting repo: %v, pool: %v", repo, pool)

	// TODO: Stop postgres

	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()

	cmds.Set(
		"zpool-destroy",
		cmd.Get("zpool", "destroy", "-f", pool.Name),
	)

	if pool.PoolType == "virtual" {
		loopbackPath, err := zfs.FindDevicePath(pool)
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
		log.Errorf("Failed to delete repo: %v, pool: %v, error: %s", repo, pool, err)
		return err
	}

	err = data.Fetcher.DeletePool(ctx, pool.ID)
	if err != nil {
		return err
	}

	return nil
}
