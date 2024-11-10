package zfs

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/db/gen/model"
)

func EmptyDataset(ctx context.Context, pool model.ZfsPool, name string) (model.ZfsDataset, error) {
	log.Infof("ZFS Dataset init %v", pool)

	datasetName := fmt.Sprintf("%s/%s", pool.Name, name)
	_, err := cmd.Single("create-dataset", false, false, "zfs", "create", datasetName)
	if err != nil {
		log.Errorf("Failed to create dataset: %s", err)
		return model.ZfsDataset{}, err
	}

	log.Infof("Created dataset. Dataset: %s Pool: %s", datasetName, pool.Name)

	// Save dataset in db
	dataset := model.ZfsDataset{
		Name:   datasetName,
		PoolID: *pool.ID,
	}

	createdDataset, err := db.CreateDataset(ctx, dataset)
	if err != nil {
		log.Errorf("Failed to insert dataset. Name: %s Pool: %v Error: %s", name, pool, err)
		return model.ZfsDataset{}, err
	}

	log.Infof("Dataset insertion successful %s", pool.Name)

	return createdDataset, nil
}
