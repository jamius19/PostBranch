package route

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	db2 "github.com/jamius19/postbranch/internal/db"
	"github.com/jamius19/postbranch/internal/dto"
	"github.com/jamius19/postbranch/internal/dto/pg"
	repo2 "github.com/jamius19/postbranch/internal/dto/repo"
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

	var repoInit repo2.InitDto[pg.HostImportReqDto]

	if err := json.NewDecoder(r.Body).Decode(&repoInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(repoInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	sameRepoCount, err := db2.CountRepoByNameOrPath(r.Context(), repoInit.RepoConfig.Name, repoInit.RepoConfig.Path)
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

	requiredSize := max(clusterSize+repo2.MinSizeInMb, 500)

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

	createdRepo, pool, err := repo.InitializeRepo(r.Context(), &repoInit)

	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	createdPg, err := host.Import(r.Context(), repoInit.PgConfig, createdRepo, pool, nil)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	pgResponse := repo2.Pg{
		ID:      createdPg.ID,
		Version: createdPg.Version,
		Status:  db2.PgStatus(createdPg.Status),
		Output:  createdPg.Output,
	}

	poolResponse := repo2.Pool{
		ID:       pool.ID,
		Type:     pool.PoolType,
		SizeInMb: pool.SizeInMb,
		Path:     pool.Path,
	}

	repoResponse := repo2.Response{
		ID:        createdRepo.ID,
		Name:      createdRepo.Name,
		Pool:      poolResponse,
		Pg:        pgResponse,
		CreatedAt: createdRepo.CreatedAt,
		UpdatedAt: createdRepo.UpdatedAt,
	}

	response := dto.Response[repo2.Response]{
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

	repoDetail, err := db2.GetRepoByName(r.Context(), repoName)
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

	requiredSize := max(clusterSize+repo2.MinSizeInMb, 500)

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

	updatedPg, err := host.Import(r.Context(), pgConfig, repoDetail.Repo, repoDetail.Pool, &repoDetail.Pg)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	pgResponse := repo2.Pg{
		ID:      updatedPg.ID,
		Version: updatedPg.Version,
		Status:  db2.PgStatus(updatedPg.Status),
		Output:  updatedPg.Output,
	}

	poolResponse := repo2.Pool{
		ID:       repoDetail.Pool.ID,
		Type:     repoDetail.Pool.PoolType,
		SizeInMb: repoDetail.Pool.SizeInMb,
		Path:     repoDetail.Pool.Path,
	}

	repoResponse := repo2.Response{
		ID:        repoDetail.Repo.ID,
		Name:      repoDetail.Repo.Name,
		Pool:      poolResponse,
		Pg:        pgResponse,
		CreatedAt: repoDetail.Repo.CreatedAt,
		UpdatedAt: repoDetail.Repo.UpdatedAt,
	}

	response := dto.Response[repo2.Response]{
		Data:  &repoResponse,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func ListRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := db2.ListRepo(r.Context())

	if err != nil {
		log.Error(err)
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	repoResponseList := []repo2.Response{}

	for _, repoDetail := range repos {
		repoResponse := getRepoResponse(repoDetail)
		repoResponseList = append(repoResponseList, repoResponse)
	}

	response := dto.Response[[]repo2.Response]{
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

	repoDetail, err := db2.GetRepoByName(r.Context(), repoName)
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

	response := dto.Response[repo2.Response]{
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

	repoDetail, err := db2.GetRepoByName(r.Context(), repoName)
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

	err = db2.DeleteRepo(r.Context(), int64(*repoDetail.Repo.ID))
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

func getRepoResponse(repoDetail db2.RepoDetail) repo2.Response {
	var pgInfo repo2.Pg
	branchesInfo := []repo2.Branch{}

	pgInfo = repo2.Pg{
		ID:      repoDetail.Pg.ID,
		Version: repoDetail.Pg.Version,
		Status:  db2.PgStatus(repoDetail.Pg.Status),
		Output:  repoDetail.Pg.Output,
	}

	poolInfo := repo2.Pool{
		ID:        repoDetail.Pool.ID,
		Type:      repoDetail.Pool.PoolType,
		Path:      repoDetail.Pool.Path,
		MountPath: repoDetail.Pool.MountPath,
		SizeInMb:  repoDetail.Pool.SizeInMb,
	}

	for _, branch := range repoDetail.Branches {
		branchesInfo = append(branchesInfo, repo2.Branch{
			ID:        branch.ID,
			Name:      branch.Name,
			Status:    db2.BranchStatus(branch.Status),
			PgStatus:  db2.BranchPgStatus(branch.PgStatus),
			Port:      branch.PgPort,
			ParentID:  branch.ParentID,
			CreatedAt: branch.CreatedAt,
			UpdatedAt: branch.UpdatedAt,
		})
	}

	repoResponse := repo2.Response{
		ID:        repoDetail.Repo.ID,
		Name:      repoDetail.Repo.Name,
		Pg:        pgInfo,
		Branches:  branchesInfo,
		Pool:      poolInfo,
		CreatedAt: repoDetail.Repo.CreatedAt,
		UpdatedAt: repoDetail.Repo.UpdatedAt,
	}
	return repoResponse
}
