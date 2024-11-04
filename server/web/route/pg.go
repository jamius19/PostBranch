package route

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/service/repo/pg"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/util/validation"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
	"strconv"
)

func Import(w http.ResponseWriter, r *http.Request) {
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

	if repoPgInit.IsHostConnection() && repoPgInit.Host != "localhost" {
		util.WriteError(w, r, responseerror.Clarify("Only localhost is supported for now"), http.StatusBadRequest)
		return
	}

	repoData, err := data.Fetcher.GetRepo(r.Context(), repoId)
	if err != nil {
		log.Error("Failed to load repo %v", repoId)

		util.WriteError(
			w,
			r,
			responseerror.Clarify("Invalid Repository ID"),
			http.StatusInternalServerError,
		)
		return
	}

	// Check if postgres is already imported for this repo
	if repoData.Repo.PgID.Valid {
		pgInfo, err := data.Fetcher.GetPg(r.Context(), repoData.Repo.PgID.Int64)
		if err != nil {
			log.Error("Failed to load pg info for repo %v", repoData.Repo)
			util.WriteError(
				w,
				r,
				responseerror.Clarify("Failed to load repositories"),
				http.StatusInternalServerError,
			)
			return
		}

		if pgInfo.Status == dao.PgCompleted {
			log.Errorf("Postgres is already imported for repository %v", repoData.Repo)

			util.WriteError(
				w,
				r,
				responseerror.Clarify("Postgres is already imported for this repository"),
				http.StatusBadRequest,
			)
			return
		} else if pgInfo.Status == dao.PgStarted {
			log.Error("Postgres import in progress for repo %v", repoData.Repo)

			util.WriteError(
				w,
				r,
				responseerror.Clarify("Postgres import in progress for this repository"),
				http.StatusBadRequest,
			)
			return
		}
	}

	updatedRepo, err := pg.Import(r.Context(), &repoPgInit, &repoData.Repo)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	util.WriteResponse(w, r, updatedRepo, http.StatusOK)
}
