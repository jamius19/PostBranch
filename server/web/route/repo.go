package route

import (
	"encoding/json"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto"
	repoDto "github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/repo"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/util/validation"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
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

		if repos[i].Repo.PgID.Valid {
			pg, err := data.Fetcher.GetPg(r.Context(), repos[i].Repo.PgID.Int64)
			if err != nil {
				log.Errorf("Failed to load pg info for repo %v", repos[i].Repo)
				util.WriteError(
					w,
					r,
					responseerror.Clarify("Failed to load repositories"),
					http.StatusInternalServerError,
				)
				return
			}

			pgInfo = &repoDto.Pg{
				PgID:    repos[i].Repo.PgID.Int64,
				Version: pg.Version,
				Status:  pg.Status,
				Output:  util.GetNullableString(&pg.Output),
			}

			branches, err := data.Fetcher.ListBranchesByRepoId(r.Context(), repos[i].Repo.ID)

			if err != nil {
				log.Error("Failed to load branches for repo %v", repos[i].Repo)
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
			ID:        repos[i].Repo.ID,
			Name:      repos[i].Repo.Name,
			Path:      repos[i].ZfsPool.Path,
			RepoType:  repos[i].Repo.RepoType,
			SizeInMb:  repos[i].ZfsPool.SizeInMb,
			Pg:        pgInfo,
			Branches:  branchesInfo,
			PoolID:    repos[i].Repo.PoolID,
			CreatedAt: repos[i].Repo.CreatedAt,
			UpdatedAt: repos[i].Repo.UpdatedAt,
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

//func ListBlockStorage(w http.ResponseWriter, r *http.Request) {
//	output, err := cmd.Single("lsblk", "-ndo", "NAME,MOUNTPOINT")
//	if err != nil {
//		log.Error(err)
//		util.WriteError(w, r, err, http.StatusInternalServerError)
//		return
//	}
//
//	devices := make(map[string]string)
//	scanner := bufio.NewScanner(strings.NewReader(*output))
//
//	for scanner.Scan() {
//		line := strings.TrimSpace(scanner.Text())
//		if line == "" {
//			continue
//		}
//
//		fields := strings.Fields(line)
//		name := fields[0]
//		mountpoint := ""
//		if len(fields) > 1 {
//			mountpoint = fields[1]
//		}
//
//		devices[name] = mountpoint
//	}
//
//	response := dto2.Response[map[string]string]{
//		Data:  &devices,
//		Error: nil,
//	}
//
//	util.WriteResponse(w, r, response, http.StatusOK)
//}
