package main

import (
	"context"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/opts"
	"github.com/jamius19/postbranch/service/zfs"
	"github.com/jamius19/postbranch/web"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var log = logger.Logger
var webWg = sync.WaitGroup{}

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

	// Channel to listen for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)

	rootCtx, rootCancel := context.WithCancel(context.Background())

	closeDb := db.Initialize()
	defer closeDb()

	go web.Initialize(rootCtx, &webWg)

	<-stop
	rootCancel()
	webWg.Wait()
}
