package route

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/internal/db"
	"github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/dto"
	"github.com/jamius19/postbranch/internal/dto/repo"
	repoSvc "github.com/jamius19/postbranch/internal/service/repo"
	"github.com/jamius19/postbranch/internal/service/validation"
	"github.com/jamius19/postbranch/internal/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
)

func CreateBranch(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repoName")
	if repoName == "" {
		util.WriteError(
			w,
			r,
			responseerror.From("Repository Name is required"),
			http.StatusBadRequest,
		)

		return
	}

	var branchInit repo.BranchInit
	if err := json.NewDecoder(r.Body).Decode(&branchInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(branchInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	repoDetail, err := db.GetRepoByName(r.Context(), repoName)
	if err != nil {
		log.Error("Failed to load repo, Invalid Repository Name: %s", repoName)

		util.WriteError(
			w,
			r,
			responseerror.From("Invalid Repository Name"),
			http.StatusNotFound,
		)
		return
	}

	// TODO: Add validation for parent branch status

	branch, err := repoSvc.CreateBranch(r.Context(), repoDetail, branchInit)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Failed to create branch"),
			http.StatusBadRequest,
		)

		return
	}

	response := dto.Response[model.Branch]{
		Data:  &branch,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func CloseBranch(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repoName")
	if repoName == "" {
		util.WriteError(
			w,
			r,
			responseerror.From("Repository Name is required"),
			http.StatusBadRequest,
		)

		return
	}

	var branchClose repo.BranchClose
	if err := json.NewDecoder(r.Body).Decode(&branchClose); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(branchClose); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if branchClose.Name == "main" {
		util.WriteError(
			w,
			r,
			responseerror.From("Cannot close main branch"),
			http.StatusBadRequest,
		)

		return
	}

	repoDetail, err := db.GetRepoByName(r.Context(), repoName)
	if err != nil {
		log.Error("Failed to load repo, Invalid Repository Name: %s", repoName)

		util.WriteError(
			w,
			r,
			responseerror.From("Invalid Repository Name"),
			http.StatusNotFound,
		)
		return
	}

	err = repoSvc.CloseBranch(r.Context(), repoDetail, branchClose)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Failed to close branch"),
			http.StatusBadRequest,
		)

		return
	}

	return
}
