package web

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jamius19/postbranch/logger"
	"net/http"
)

func middlewares(r *chi.Mux, rootCtx context.Context) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger.Logger}))
	r.Use(shutdownContextMiddleware(rootCtx))
}

func shutdownContextMiddleware(rootCtx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create combined context
			ctx, cancel := context.WithCancel(r.Context())
			go func() {
				select {
				case <-rootCtx.Done():
					cancel()
				case <-ctx.Done():
				}
			}()

			// Add combined context to request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
