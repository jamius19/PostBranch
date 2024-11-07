package route

import (
	"encoding/json"
	"github.com/jamius19/postbranch/data/dto"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/service/repo/pg"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/util/validation"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
)

func ValidatePg(w http.ResponseWriter, r *http.Request) {
	//repoId, err := strconv.ParseInt(chi.URLParam(r, "repoId"), 10, 64)
	//if err != nil {
	//	util.WriteError(
	//		w,
	//		r,
	//		responseerror.From("Repo Id should be a number"),
	//		http.StatusBadRequest,
	//	)
	//
	//	return
	//}

	var pgInit repo.PgInitDto
	if err := json.NewDecoder(r.Body).Decode(&pgInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(pgInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if pgInit.IsHostConnection() && pgInit.Host != "localhost" {
		util.WriteError(w, r, responseerror.From("Only localhost is supported for now"), http.StatusBadRequest)
		return
	}

	log.Infof("Validating Postgres, pgInit: %s", pgInit.String())

	//repoDataQuery, err := data.Fetcher.GetRepo(r.Context(), repoId)
	//if err != nil {
	//	log.Error("Failed to load repo %v", repoId)
	//
	//	util.WriteError(
	//		w,
	//		r,
	//		responseerror.From("Invalid Repository Id"),
	//		http.StatusInternalServerError,
	//	)
	//	return
	//}
	//
	//repoData, pool, pgInfo := conversion.SplitRepoRow(&repoDataQuery)

	//// Check if postgres is already imported for this repo
	//if pgInfo != nil {
	//	log.Infof("Existing postgres for repo: %v, pg: %v", repoData, pgInfo)
	//
	//	pgInfo, err := data.Fetcher.GetPg(r.Context(), pgInfo.Id)
	//	if err != nil {
	//		log.Error("Failed to load pg info for repo %v", repoData)
	//		util.WriteError(
	//			w,
	//			r,
	//			responseerror.From("Failed to load repositories"),
	//			http.StatusInternalServerError,
	//		)
	//		return
	//	}
	//
	//	if pgInfo.Status == dao.PgCompleted {
	//		log.Errorf("Postgres is already imported for repository %v", repoData)
	//
	//		util.WriteError(
	//			w,
	//			r,
	//			responseerror.From("Postgres is already imported for this repository"),
	//			http.StatusBadRequest,
	//		)
	//		return
	//	} else if pgInfo.Status == dao.PgStarted {
	//		log.Error("Postgres import in progress for repo %v", repoData)
	//
	//		util.WriteError(
	//			w,
	//			r,
	//			responseerror.From("Postgres import in progress for this repository"),
	//			http.StatusBadRequest,
	//		)
	//		return
	//	}
	//}

	err := pg.Validate(&pgInit)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	clusterSizeInMb, err := pg.GetClusterSize(&pgInit)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	pgInitWithSize := repo.PgInitResponseDto{
		PgInitDto:       pgInit,
		ClusterSizeInMb: clusterSizeInMb,
	}

	response := dto.Response[repo.PgInitResponseDto]{
		Data:  &pgInitWithSize,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}
