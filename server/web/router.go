package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/web/route"
)

func routes(r *chi.Mux) {
	r.Route("/api/settings", func(r chi.Router) {
		r.Post("/", route.AddSettings)
		r.Get("/{key}", route.GetSettings)
	})

	r.Route("/api/repo", func(r chi.Router) {
		r.Post("/", route.InitializeRepo)
	})
}
