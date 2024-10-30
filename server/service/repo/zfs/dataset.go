package zfs

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
)

func EmptyDataset(ctx context.Context, pool *dao.ZfsPool, name string) (*dao.ZfsDataset, error) {
	log.Infof("ZFS Dataset init %v", *pool)

	datasetName := fmt.Sprintf("%s/%s", pool.Name, name)
	_, err := cmd.Single("create-dataset", false, "zfs", "create", datasetName)
	if err != nil {
		log.Errorf("Failed to create dataset: %s", err)
		return nil, err
	}

	log.Infof("Created dataset. Dataset: %s Pool: %s", datasetName, pool.Name)

	// Save dataset in db
	dataset := dao.CreateDatasetParams{
		Name:   datasetName,
		PoolID: pool.ID,
	}

	createdDataset, err := data.Fetcher.CreateDataset(ctx, dataset)
	if err != nil {
		log.Errorf("Failed to insert dataset. Name: %s Pool: %v Error: %s", name, pool, err)
		return nil, err
	}

	log.Infof("Dataset insertion successful %s", pool.Name)

	return &createdDataset, nil
}
