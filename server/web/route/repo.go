package route

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/dto"
	"github.com/jamius19/postbranch/dto/pg"
	repoDto "github.com/jamius19/postbranch/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/pg/adapter/host"
	"github.com/jamius19/postbranch/service/pg/adapter/local"
	"github.com/jamius19/postbranch/service/repo"
	"github.com/jamius19/postbranch/service/validation"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
	"strconv"
)

var log = logger.Logger

func InitializeLocalRepo(w http.ResponseWriter, r *http.Request) {
	log.Info("Initializing local repo")

	var repoInit repoDto.InitDto[pg.LocalImportReqDto]

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

	if err := local.Validate(repoInit.PgConfig); err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Postgres configuration is invalid, please start again"),
			http.StatusBadRequest,
		)

		return
	}

	clusterSize, err := local.GetClusterSize(repoInit.PgConfig)
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

	createdRepo, pool, err := repo.InitializeRepo(r.Context(), &repoInit)

	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	createdPg, err := local.Import(r.Context(), repoInit.PgConfig, createdRepo, pool, nil)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	pgResponse := repoDto.Pg{
		ID:      createdPg.ID,
		Version: createdPg.Version,
		Status:  db.PgStatus(createdPg.Status),
		Output:  createdPg.Output,
	}

	poolResponse := repoDto.Pool{
		ID:       pool.ID,
		Type:     pool.PoolType,
		SizeInMb: pool.SizeInMb,
		Path:     pool.Path,
	}

	repoResponse := repoDto.Response{
		ID:        createdRepo.ID,
		Name:      createdRepo.Name,
		Pool:      poolResponse,
		Pg:        pgResponse,
		CreatedAt: createdRepo.CreatedAt,
		UpdatedAt: createdRepo.UpdatedAt,
	}

	response := dto.Response[repoDto.Response]{
		Data:  &repoResponse,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

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

	pgResponse := repoDto.Pg{
		ID:      createdPg.ID,
		Version: createdPg.Version,
		Status:  db.PgStatus(createdPg.Status),
		Output:  createdPg.Output,
	}

	poolResponse := repoDto.Pool{
		ID:       pool.ID,
		Type:     pool.PoolType,
		SizeInMb: pool.SizeInMb,
		Path:     pool.Path,
	}

	repoResponse := repoDto.Response{
		ID:        createdRepo.ID,
		Name:      createdRepo.Name,
		Pool:      poolResponse,
		Pg:        pgResponse,
		CreatedAt: createdRepo.CreatedAt,
		UpdatedAt: createdRepo.UpdatedAt,
	}

	response := dto.Response[repoDto.Response]{
		Data:  &repoResponse,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func ReInitializeLocalPg(w http.ResponseWriter, r *http.Request) {
	log.Infof("Re-Initializing local repo with ID: %s", chi.URLParam(r, "repoId"))

	repoId, err := strconv.ParseInt(chi.URLParam(r, "repoId"), 10, 64)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Repo Id should be a number"),
			http.StatusBadRequest,
		)

		return
	}

	var pgConfig pg.LocalImportReqDto
	if err := json.NewDecoder(r.Body).Decode(&pgConfig); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(pgConfig); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	repoDetail, err := db.GetRepo(r.Context(), repoId)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Invalid repository id"),
			http.StatusInternalServerError,
		)

		return
	}

	if err := local.Validate(pgConfig); err != nil {
		util.WriteError(
			w,
			r,
			err,
			http.StatusBadRequest,
		)

		return
	}

	clusterSize, err := local.GetClusterSize(pgConfig)
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

	updatedPg, err := local.Import(r.Context(), pgConfig, repoDetail.Repo, repoDetail.Pool, &repoDetail.Pg)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	pgResponse := repoDto.Pg{
		ID:      updatedPg.ID,
		Version: updatedPg.Version,
		Status:  db.PgStatus(updatedPg.Status),
		Output:  updatedPg.Output,
	}

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
		Pg:        pgResponse,
		CreatedAt: repoDetail.Repo.CreatedAt,
		UpdatedAt: repoDetail.Repo.UpdatedAt,
	}

	response := dto.Response[repoDto.Response]{
		Data:  &repoResponse,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}

func ReInitializeHostPg(w http.ResponseWriter, r *http.Request) {
	log.Infof("Re-Initializing host repo with ID: %s", chi.URLParam(r, "repoId"))

	repoId, err := strconv.ParseInt(chi.URLParam(r, "repoId"), 10, 64)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Repo Id should be a number"),
			http.StatusBadRequest,
		)

		return
	}

	var pgConfig pg.HostImportReqDto
	if err := json.NewDecoder(r.Body).Decode(&pgConfig); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(pgConfig); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	repoDetail, err := db.GetRepo(r.Context(), repoId)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Invalid repository id"),
			http.StatusInternalServerError,
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

	updatedPg, err := host.Import(r.Context(), pgConfig, repoDetail.Repo, repoDetail.Pool, &repoDetail.Pg)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	pgResponse := repoDto.Pg{
		ID:      updatedPg.ID,
		Version: updatedPg.Version,
		Status:  db.PgStatus(updatedPg.Status),
		Output:  updatedPg.Output,
	}

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
		Pg:        pgResponse,
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
	repoId, err := strconv.ParseInt(chi.URLParam(r, "repoId"), 10, 64)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Repository ID should be a number"),
			http.StatusBadRequest,
		)

		return
	}

	repoDetail, err := db.GetRepo(r.Context(), repoId)
	if err != nil {
		log.Error("Failed to load repo, Invalid Repository ID: %d", repoId)

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
	repoId, err := strconv.ParseInt(chi.URLParam(r, "repoId"), 10, 64)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Repo Id should be a number"),
			http.StatusBadRequest,
		)

		return
	}

	repoDetail, err := db.GetRepo(r.Context(), repoId)
	if err != nil {
		log.Error("Failed to load repo, Invalid Repository Id: %d", repoId)

		util.WriteError(
			w,
			r,
			responseerror.From("Invalid Repository ID"),
			http.StatusNotFound,
		)
		return
	}

	err = repo.DeleteRepo(r.Context(), repoDetail.Repo, repoDetail.Pool, repoDetail.Pg)
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
	var pgInfo repoDto.Pg
	branchesInfo := []repoDto.Branch{}

	pgInfo = repoDto.Pg{
		ID:      repoDetail.Pg.ID,
		Version: repoDetail.Pg.Version,
		Status:  db.PgStatus(repoDetail.Pg.Status),
		Output:  repoDetail.Pg.Output,
	}

	poolInfo := repoDto.Pool{
		ID:       repoDetail.Pool.ID,
		Type:     repoDetail.Pool.PoolType,
		Path:     repoDetail.Pool.Path,
		SizeInMb: repoDetail.Pool.SizeInMb,
	}

	for _, branch := range repoDetail.Branches {
		branchesInfo = append(branchesInfo, repoDto.Branch{
			ID:       branch.ID,
			Name:     branch.Name,
			Status:   db.BranchStatus(branch.Status),
			PgStatus: db.BranchPgStatus(branch.PgStatus),
			Port:     branch.PgPort,
			ParentID: branch.ParentID,
		})
	}

	repoResponse := repoDto.Response{
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
