package repo

import (
	"github.com/jamius19/postbranch/dto"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/repo/zfs"
)

var log = logger.Logger

func InitializeRepo(repoinit *dto.RepoInit) error {
	if repoinit.RepoType == "virtual" {
		log.Infof("Initializing virtual repo")
		
		err := initializeVirtual(repoinit)
		if err != nil {
			return err
		}
	}

	return nil
}

func initializeVirtual(repoinit *dto.RepoInit) error {
	err := zfs.InitializeVirtual(repoinit)
	if err != nil {
		return err
	}

	return nil
}
