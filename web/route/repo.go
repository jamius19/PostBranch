package route

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/internal/db"
	"github.com/jamius19/postbranch/internal/dto"
	"github.com/jamius19/postbranch/internal/dto/pg"
	repoDto "github.com/jamius19/postbranch/internal/dto/repo"
	"github.com/jamius19/postbranch/internal/logger"
	"github.com/jamius19/postbranch/internal/service/pg/adapter/host"
	"github.com/jamius19/postbranch/internal/service/repo"
	"github.com/jamius19/postbranch/internal/service/validation"
	"github.com/jamius19/postbranch/internal/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
)

var log = logger.Logger

func InitializeHostRepo(w http.ResponseWriter, r *http.Request) {
	log.Info("Initializing host repo")

	var repoInit repoDto.InitDto[pg.HostImportReqDto]

	if err := json.NewDecoder(r.Body).Decode(&repoInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(repoInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	sameRepoCount, err := db.CountRepoByNameOrPath(r.Context(), repoInit.RepoConfig.Name, repoInit.RepoConfig.Path)
	if err != nil {
		log.Errorf("Error fetching similar repository. RepoInitDto: %v", &repoInit)
		util.WriteError(
			w,
			r,
			responseerror.From("Error fetching similar repository"),
			http.StatusInternalServerError,
		)

		return
	} else if sameRepoCount > 0 {
		log.Errorf("Repo exists with same name and/or path. RepoInitDto: %v", &repoInit)
		util.WriteError(
			w,
			r,
			responseerror.From("Repository exists with same name and/or path"),
			http.StatusBadRequest,
		)

		return
	}

	if err := host.Validate(repoInit.PgConfig); err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Postgres configuration is invalid, please start again"),
			http.StatusBadRequest,
		)

		return
	}

	clusterSize, err := host.GetClusterSize(repoInit.PgConfig)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Can't connect to PostgreSQL. Is it running and is the provided configuration correct?"),
			http.StatusInternalServerError,
		)

		return
	}

	requiredSize := max(clusterSize+repoDto.MinSizeInMb, 500)

	if repoInit.RepoConfig.SizeInMb < requiredSize {
		log.Errorf("Requested size of %d MB is too small. Cluster size should be at least %d MB",
			repoInit.RepoConfig.SizeInMb, requiredSize)

		util.WriteError(
			w,
			r,
			responseerror.From(
				fmt.Sprintf("Requested size of %d MB is too small. Cluster size should be at least %d MB",
					repoInit.RepoConfig.SizeInMb, requiredSize),
			),
			http.StatusBadRequest,
		)

		return
	}

	repoInfo, pool, err := repo.InitializeRepo(r.Context(), &repoInit, &repoInit.PgConfig)

	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	host.Import(repoInit.PgConfig, repoInfo, pool)

	poolResponse := repoDto.Pool{
		ID:       pool.ID,
		Type:     pool.PoolType,
		SizeInMb: pool.SizeInMb,
		Path:     pool.Path,
	}

	repoResponse := repoDto.Response{
		ID:        repoInfo.ID,
		Name:      repoInfo.Name,
		Pool:      poolResponse,
		Output:    repoInfo.Output,
		Status:    db.RepoStatus(repoInfo.Status),
		CreatedAt: repoInfo.CreatedAt,
		UpdatedAt: repoInfo.UpdatedAt,
	}

	response := dto.Response[repoDto.Response]{
		Data:  &repoResponse,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func ReInitializeHostPg(w http.ResponseWriter, r *http.Request) {
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

	log.Infof("Re-Initializing host repo with name: %s", repoName)

	var pgConfig pg.HostImportReqDto
	if err := json.NewDecoder(r.Body).Decode(&pgConfig); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(pgConfig); err != nil {
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

	if err := host.Validate(pgConfig); err != nil {
		util.WriteError(
			w,
			r,
			err,
			http.StatusBadRequest,
		)

		return
	}

	clusterSize, err := host.GetClusterSize(pgConfig)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Can't connect to PostgreSQL. Is it running and is the provided configuration correct?"),
			http.StatusInternalServerError,
		)

		return
	}

	requiredSize := max(clusterSize+repoDto.MinSizeInMb, 500)

	if repoDetail.Pool.SizeInMb < requiredSize {
		log.Errorf("Database Cluster size is %d MB, but the repository size is %d MB. Please increase the repository size",
			requiredSize, repoDetail.Pool.SizeInMb)

		util.WriteError(
			w,
			r,
			responseerror.From(
				fmt.Sprintf("Database Cluster size is %d MB, but the repository size is %d MB. Please increase the repository size",
					requiredSize, repoDetail.Pool.SizeInMb),
			),
			http.StatusBadRequest,
		)

		return
	}

	host.Import(pgConfig, repoDetail.Repo, repoDetail.Pool)

	poolResponse := repoDto.Pool{
		ID:       repoDetail.Pool.ID,
		Type:     repoDetail.Pool.PoolType,
		SizeInMb: repoDetail.Pool.SizeInMb,
		Path:     repoDetail.Pool.Path,
	}

	repoResponse := repoDto.Response{
		ID:        repoDetail.Repo.ID,
		Name:      repoDetail.Repo.Name,
		Pool:      poolResponse,
		PgVersion: pgConfig.Version,
		Status:    db.RepoStatus(repoDetail.Repo.Status),
		Output:    repoDetail.Repo.Output,
		CreatedAt: repoDetail.Repo.CreatedAt,
		UpdatedAt: repoDetail.Repo.UpdatedAt,
	}

	response := dto.Response[repoDto.Response]{
		Data:  &repoResponse,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func ListRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := db.ListRepo(r.Context())

	if err != nil {
		log.Error(err)
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	repoResponseList := []repoDto.Response{}

	for _, repoDetail := range repos {
		repoResponse := getRepoResponse(repoDetail)
		repoResponseList = append(repoResponseList, repoResponse)
	}

	response := dto.Response[[]repoDto.Response]{
		Data:   &repoResponseList,
		Error:  nil,
		IsList: true,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func GetRepo(w http.ResponseWriter, r *http.Request) {
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

	repoDetail, err := db.GetRepoByName(r.Context(), repoName)
	if err != nil {
		log.Error("Failed to load repo, Invalid Repository Name: %s", repoName)

		util.WriteError(
			w,
			r,
			responseerror.From("Invalid Repository ID"),
			http.StatusNotFound,
		)
		return
	}

	repoResponse := getRepoResponse(repoDetail)

	response := dto.Response[repoDto.Response]{
		Data:  &repoResponse,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func DeleteRepo(w http.ResponseWriter, r *http.Request) {
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

	err = repo.DeleteRepo(r.Context(), repoDetail)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Failed to delete repository"),
			http.StatusInternalServerError,
		)

		return
	}

	err = db.DeleteRepo(r.Context(), *repoDetail.Repo.ID)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Failed to delete repository"),
			http.StatusInternalServerError,
		)

		return
	}

	response := dto.Response[int32]{
		Data:  repoDetail.Repo.ID,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func getRepoResponse(repoDetail db.RepoDetail) repoDto.Response {
	branchesInfo := []repoDto.Branch{}

	poolInfo := repoDto.Pool{
		ID:        repoDetail.Pool.ID,
		Type:      repoDetail.Pool.PoolType,
		Path:      repoDetail.Pool.Path,
		MountPath: repoDetail.Pool.MountPath,
		SizeInMb:  repoDetail.Pool.SizeInMb,
	}

	for _, branch := range repoDetail.Branches {
		branchesInfo = append(branchesInfo, repoDto.Branch{
			ID:        branch.ID,
			Name:      branch.Name,
			Status:    db.BranchStatus(branch.Status),
			PgStatus:  db.BranchPgStatus(branch.PgStatus),
			Port:      branch.PgPort,
			ParentID:  branch.ParentID,
			CreatedAt: branch.CreatedAt,
			UpdatedAt: branch.UpdatedAt,
		})
	}

	repoResponse := repoDto.Response{
		ID:        repoDetail.Repo.ID,
		Name:      repoDetail.Repo.Name,
		PgVersion: repoDetail.Repo.Version,
		Status:    db.RepoStatus(repoDetail.Repo.Status),
		Output:    repoDetail.Repo.Output,
		Branches:  branchesInfo,
		Pool:      poolInfo,
		CreatedAt: repoDetail.Repo.CreatedAt,
		UpdatedAt: repoDetail.Repo.UpdatedAt,
	}
	return repoResponse
}
