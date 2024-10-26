package web

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/opts"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var log = logger.Logger

func Initialize() {
	// Channel to listen for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)

	rootCtx, rootCancel := context.WithCancel(context.Background())

	r := chi.NewRouter()
	middlewares(r, rootCtx)
	routes(r)

	log.Infof("Starting server on port %d", opts.Config.Server.Port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", opts.Config.Server.Port),
		Handler: r,
	}

	go start(srv)

	// Wait for interrupt signal
	<-stop
	rootCancel()

	log.Info("Received interrupt/terminate signal, shutting down...")

	// Create shutdown context as child of root context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server shutdown error: %v", err)
	}
}

func start(server *http.Server) {
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Logger.Fatal(err)
	}
}
