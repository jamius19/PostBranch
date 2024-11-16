package middleware

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jamius19/postbranch/logger"
)

func Middlewares(r *chi.Mux, rootCtx context.Context) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.StripSlashes)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger.Logger}))
	r.Use(shutdownContext(rootCtx))
	r.Use(requestError)
}
