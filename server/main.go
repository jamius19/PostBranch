package main

import (
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/opts"
	"github.com/jamius19/postbranch/service/repo/zfs"
	"github.com/jamius19/postbranch/web"
	"os"
)

var log = logger.Logger

func main() {
	if os.Geteuid() != 0 {
		log.Fatal("PostBranch must be run with sudo privileges")
	}

	err := opts.Load()

	if err != nil {
		log.Fatal("Failed to load config")
	}
	log.Info("Config loaded")

	version, compatible := zfs.Version()

	if !compatible {
		log.Fatal("ZFS version is not compatible.")
	}

	log.Infof("Compatible ZFS version found (%s). Continuing...", *version)

	db := data.Initialize()
	defer db.Close()

	web.Initialize()
}
