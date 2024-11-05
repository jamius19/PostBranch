package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/web/route"
)

func routes(r *chi.Mux) {
	r.Route("/api", func(r chi.Router) {
		r.Route("/repos", func(r chi.Router) {
			r.Get("/", route.ListRepos)
			r.Get("/{repoId}", route.GetRepo)

			r.Post("/", route.InitializeRepo)
			r.Post("/{repoId}/postgres", route.ImportPg)

			r.Delete("/{repoId}", route.DeleteRepo)
		})
	})
}
