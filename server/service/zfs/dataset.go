package zfs

import (
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/db/gen/model"
)

func EmptyDataset(pool model.ZfsPool, name string) error {
	log.Infof("ZFS Dataset init %v", pool)

	datasetName := fmt.Sprintf("%s/%s", pool.Name, name)
	_, err := cmd.Single("create-dataset", false, false, "zfs", "create", datasetName)
	if err != nil {
		log.Errorf("Failed to create dataset: %s", err)
		return err
	}

	log.Infof("Created dataset. Dataset: %s Pool: %s", datasetName, pool.Name)
	return nil
}
