package web

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/opts"
	"github.com/jamius19/postbranch/service/zfs"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/middleware"
	"net/http"
	"sync"
	"time"
)

var log = logger.Logger

func Initialize(rootCtx context.Context, webWg *sync.WaitGroup) {
	defer webWg.Done()

	err := zfs.MountAll(rootCtx)
	if err != nil {
		log.Fatal("Failed to mount ZFS pool(s). Error: %s", err)
	}

	select {
	case <-rootCtx.Done():
		log.Info("Root context cancelled. Unmounting pools")
		err := zfs.UnmountAll()
		if err != nil {
			log.Errorf("Failed to unmount ZFS pool(s). error: %s", err)
			return
		}
		return
	default:
	}

	r := chi.NewRouter()
	middleware.Middlewares(r, rootCtx)
	routes(r)

	log.Infof("Starting server on port %d", opts.Config.Server.Port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", opts.Config.Server.Port),
		Handler: r,
	}

	go start(srv)
	util.PrintReadyBanner()

	// Wait for interrupt signal
	select {
	case <-rootCtx.Done():
		log.Info("Received interrupt/terminate signal, shutting down...")
	}

	err = zfs.UnmountAll()
	if err != nil {
		log.Errorf("Failed to unmount ZFS pool(s). error: %s", err)
	}

	// Create shutdown context as child of root context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Failed to shutdown server. error: %v", err)
	}
}

func start(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Logger.Fatal(err)
	}
}
