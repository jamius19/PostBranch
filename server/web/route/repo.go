package route

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	dto2 "github.com/jamius19/postbranch/data/dto"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/repo"
	"github.com/jamius19/postbranch/util"
	"net/http"
)

var validate = validator.New()
var log = logger.Logger

func InitializeRepo(w http.ResponseWriter, r *http.Request) {
	var repoinit dto2.RepoInit

	if err := json.NewDecoder(r.Body).Decode(&repoinit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(repoinit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	repoResponse, err := repo.InitializeRepo(r.Context(), &repoinit)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	response := dto2.Response[dto2.RepoResponse]{
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

	response := dto2.Response[[]dao.Repo]{
		Data:  &repos,
		Error: nil,
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
