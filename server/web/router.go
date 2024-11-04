package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/web/route"
)

func routes(r *chi.Mux) {
	r.Route("/api/repos", func(r chi.Router) {
		r.Get("/", route.ListRepos)
		r.Post("/", route.InitializeRepo)
		r.Post("/{repoId}/postgres", route.Import)

		//r.Get("/block-storages", route.ListBlockStorage)
	})
}
