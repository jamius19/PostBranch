package zfs

import (
	"fmt"
	"github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/runner"
	"github.com/jamius19/postbranch/internal/util"
	"os"
	"path/filepath"
)

func EmptyDataset(pool model.ZfsPool, branchName string) error {
	log.Infof("ZFS Dataset init %v", pool)
	datasetName := fmt.Sprintf("%s/%s", pool.Name, branchName)
	datasetPath := filepath.Join(pool.MountPath, branchName)

	if _, err := os.Stat(datasetPath); err == nil {
		log.Errorf("Dataset path already exists: %s, removing", datasetPath)

		datasetFileRemovePattern := filepath.Join(pool.MountPath, branchName, "*")
		err := util.RemoveGlob(datasetFileRemovePattern)
		if err != nil {
			log.Errorf("Failed to remove existing data from dataset path: %s", err)
			return err
		}

		return nil
	}

	_, err := runner.Single("create-dataset", false, false, "zfs", "create", datasetName)
	if err != nil {
		log.Errorf("Failed to create dataset: %s", err)
		return err
	}

	log.Infof("Created dataset. Dataset: %s Pool: %s", datasetName, pool.Name)
	return nil
}
