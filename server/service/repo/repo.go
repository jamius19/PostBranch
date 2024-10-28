package repo

import (
	"context"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/repo/zfs"
	"github.com/jamius19/postbranch/web/responseerror"
)

var log = logger.Logger

func InitializeRepo(ctx context.Context, repoinit *dto.RepoInit) (*dto.RepoResponse, error) {
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
			Name:     repoinit.Name,
			PoolID:   pool.ID,
			RepoType: repoinit.RepoType,
			Size:     repoinit.Size,
			SizeUnit: repoinit.SizeUnit,
		}

		createdRepo, err := data.Fetcher.CreateRepo(ctx, repoCreateDto)
		if err != nil {
			// TODO: Cleanup Pool and Dataset
			log.Infof("Failed to insert repo. Name: %s Data: %v Error: %s", repoinit.Name, repoCreateDto, err)
			return nil, responseerror.Clarify("Failed to create repository")
		}

		repoResponse := dto.RepoResponse{
			ID:        createdRepo.ID,
			Name:      createdRepo.Name,
			Path:      pool.Path,
			RepoType:  createdRepo.RepoType,
			Size:      createdRepo.Size,
			SizeUnit:  createdRepo.SizeUnit,
			PgID:      nil,
			PoolID:    pool.ID,
			CreatedAt: createdRepo.CreatedAt,
			UpdatedAt: createdRepo.UpdatedAt,
		}

		return &repoResponse, nil
	}

	return nil, nil
}
