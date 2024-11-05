package route

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dao/conversion"
	"github.com/jamius19/postbranch/data/dto"
	repoDto "github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/repo"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/util/validation"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
	"strconv"
)

var log = logger.Logger

func InitializeRepo(w http.ResponseWriter, r *http.Request) {
	var repoInit repoDto.InitDto

	if err := json.NewDecoder(r.Body).Decode(&repoInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(repoInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	sameRepoCount, err := data.Fetcher.CountRepoByNameOrPath(r.Context(), dao.CountRepoByNameOrPathParams{
		Name: repoInit.Name,
		Path: repoInit.Path,
	})

	if err != nil {
		log.Errorf("Error fetching similar repository. RepoInitDto: %v", &repoInit)
		util.WriteError(
			w,
			r,
			responseerror.Clarify("Error fetching similar repository"),
			http.StatusInternalServerError,
		)

		return
	} else if sameRepoCount > 0 {
		log.Errorf("Repo exists with same name and/or path. RepoInitDto: %v", &repoInit)
		util.WriteError(
			w,
			r,
			responseerror.Clarify("Repository exists with same name and/or path"),
			http.StatusBadRequest,
		)

		return
	}

	repoResponse, err := repo.InitializeRepo(r.Context(), &repoInit)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	response := dto.Response[repoDto.Response]{
		Data:  repoResponse,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func ListRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := data.Fetcher.ListRepo(r.Context())

	if err != nil {
		log.Error(err)
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	reposResponse := []repoDto.Response{}
	for i := range repos {
		var pgInfo *repoDto.Pg = nil
		branchesInfo := []repoDto.Branch{}

		repoData, pool, pg := conversion.SplitRepoRow((*dao.GetRepoRow)(&repos[i]))

		if pg != nil {
			pg, err := data.Fetcher.GetPg(r.Context(), pg.ID)
			if err != nil {
				log.Errorf("Failed to load pg info for repo %v", repoData)
				util.WriteError(
					w,
					r,
					responseerror.Clarify("Failed to load repositories"),
					http.StatusInternalServerError,
				)
				return
			}

			pgInfo = &repoDto.Pg{
				PgID:    pg.ID,
				Version: pg.Version,
				Status:  pg.Status,
				Output:  util.GetNullableString(&pg.Output),
			}

			branches, err := data.Fetcher.ListBranchesByRepoId(r.Context(), repoData.ID)

			if err != nil {
				log.Error("Failed to load branches for repo %v", repoData)
				util.WriteError(
					w,
					r,
					responseerror.Clarify("Failed to load repositories"),
					http.StatusInternalServerError,
				)
				return
			}

			for _, branch := range branches {
				branchesInfo = append(branchesInfo, repoDto.Branch{
					Id:       branch.ID,
					Name:     branch.Name,
					ParentId: util.GetNullableInt64(&branch.ParentID),
				})
			}
		}

		repoResponse := repoDto.Response{
			ID:        repoData.ID,
			Name:      repoData.Name,
			Path:      pool.Path,
			SizeInMb:  pool.SizeInMb,
			Pg:        pgInfo,
			Branches:  branchesInfo,
			PoolID:    pool.ID,
			CreatedAt: repoData.CreatedAt,
			UpdatedAt: repoData.UpdatedAt,
		}

		reposResponse = append(reposResponse, repoResponse)
	}

	response := dto.Response[[]repoDto.Response]{
		Data:   &reposResponse,
		Error:  nil,
		IsList: true,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func GetRepo(w http.ResponseWriter, r *http.Request) {
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

	repoRow, err := data.Fetcher.GetRepo(r.Context(), repoId)
	if err != nil {
		log.Error("Failed to load repo, Invalid Repository ID: %d", repoId)

		util.WriteError(
			w,
			r,
			responseerror.Clarify("Invalid Repository ID"),
			http.StatusNotFound,
		)
		return
	}

	var pgInfo *repoDto.Pg = nil
	branchesInfo := []repoDto.Branch{}

	repoData, pool, pg := conversion.SplitRepoRow(&repoRow)
	if pg != nil {
		pg, err := data.Fetcher.GetPg(r.Context(), pg.ID)
		if err != nil {
			log.Errorf("Failed to load pg info for repo %v", repoData)
			util.WriteError(
				w,
				r,
				responseerror.Clarify("Failed to load repositories"),
				http.StatusInternalServerError,
			)
			return
		}

		pgInfo = &repoDto.Pg{
			PgID:    pg.ID,
			Version: pg.Version,
			Status:  pg.Status,
			Output:  util.GetNullableString(&pg.Output),
		}

		branches, err := data.Fetcher.ListBranchesByRepoId(r.Context(), repoData.ID)

		if err != nil {
			log.Error("Failed to load branches for repo %v", repoData)
			util.WriteError(
				w,
				r,
				responseerror.Clarify("Failed to load repositories"),
				http.StatusInternalServerError,
			)
			return
		}

		for _, branch := range branches {
			branchesInfo = append(branchesInfo, repoDto.Branch{
				Id:       branch.ID,
				Name:     branch.Name,
				ParentId: util.GetNullableInt64(&branch.ParentID),
			})
		}
	}

	repoResponse := repoDto.Response{
		ID:        repoData.ID,
		Name:      repoData.Name,
		Path:      pool.Path,
		SizeInMb:  pool.SizeInMb,
		Pg:        pgInfo,
		Branches:  branchesInfo,
		PoolID:    pool.ID,
		CreatedAt: repoData.CreatedAt,
		UpdatedAt: repoData.UpdatedAt,
	}

	response := dto.Response[repoDto.Response]{
		Data:  &repoResponse,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func DeleteRepo(w http.ResponseWriter, r *http.Request) {
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

	repoRow, err := data.Fetcher.GetRepo(r.Context(), repoId)
	if err != nil {
		log.Error("Failed to load repo, Invalid Repository ID: %d", repoId)

		util.WriteError(
			w,
			r,
			responseerror.Clarify("Invalid Repository ID"),
			http.StatusNotFound,
		)
		return
	}

	repoData, pool, _ := conversion.SplitRepoRow(&repoRow)
	err = repo.DeleteRepo(r.Context(), repoData, pool)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.Clarify("Failed to delete repository"),
			http.StatusInternalServerError,
		)

		return
	}

	response := dto.Response[int64]{
		Data:  &repoData.ID,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}
