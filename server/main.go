package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/logger"
	"net/http"
)

func main() {
	db.Initialize()
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger.Logger}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	logger.Logger.Info("Starting server on port 9099")
	err := http.ListenAndServe(":9099", r)

	if err != nil {
		logger.Logger.Fatal(err)
	}
}
