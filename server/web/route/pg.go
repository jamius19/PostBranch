package route

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/service/repo/pg"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/util/validation"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
	"strconv"
)

func ImportPostgres(w http.ResponseWriter, r *http.Request) {
	repoId, err := strconv.ParseInt(chi.URLParam(r, "repoId"), 10, 64)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.Clarify("Repo ID should be a number"),
			http.StatusBadRequest,
		)

		return
	}

	var repoPgInit repo.PgInitDto
	if err := json.NewDecoder(r.Body).Decode(&repoPgInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(repoPgInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if repoPgInit.CustomConnection && repoPgInit.Host != "localhost" {
		util.WriteError(w, r, responseerror.Clarify("Only localhost is supported for now"), http.StatusBadRequest)
		return
	}

	repo, err := data.Fetcher.GetRepo(r.Context(), repoId)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.Clarify("Invalid Repo ID"),
			http.StatusInternalServerError,
		)
		return
	}

	if err := pg.Import(r.Context(), repoPgInit, &repo); err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}
}
