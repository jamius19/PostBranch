package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/web/route"
)

func routes(r *chi.Mux) {
	r.Route("/api", func(r chi.Router) {
		r.Route("/repos", func(r chi.Router) {
			r.Get("/", route.ListRepos)
			r.Get("/{repoName}", route.GetRepo)
			r.Post("/{repoName}/branch", route.CreateBranch)
			r.Post("/{repoName}/branch/close", route.CloseBranch)

			// Adapters for different pg sources
			r.Route("/postgres/validate", func(r chi.Router) {
				r.Post("/host", route.ValidateHostPg)
			})

			// Adapters for different pg sources
			r.Route("/import", func(r chi.Router) {
				r.Post("/host", route.InitializeHostRepo)
				r.Post("/{repoName}/host", route.ReInitializeHostPg)
			})

			r.Delete("/{repoName}", route.DeleteRepo)
		})

	})
}
