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

			// Adapters for different pg sources
			r.Route("/postgres/validate", func(r chi.Router) {
				r.Post("/local", route.ValidateLocalPg)
				r.Post("/host", route.ValidateHostPg)
			})

			// Adapters for different pg sources
			r.Route("/import", func(r chi.Router) {
				r.Post("/local", route.InitializeLocalRepo)
				r.Post("/{repoId}/local", route.ReInitializeLocalPg)

				r.Post("/host", route.InitializeHostRepo)
				r.Post("/{repoId}/host", route.ReInitializeHostPg)
			})

			r.Delete("/{repoId}", route.DeleteRepo)
		})

	})
}
