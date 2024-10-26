package route

import (
	"bufio"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/fetch"
	"github.com/jamius19/postbranch/dto"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/repo"
	"github.com/jamius19/postbranch/util"
	"net/http"
	"strings"
)

var validate = validator.New()
var log = logger.Logger

func InitializeRepo(w http.ResponseWriter, r *http.Request) {
	var repoinit dto.RepoInit

	if err := json.NewDecoder(r.Body).Decode(&repoinit); err != nil {
		util.WriteError(w, err, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(repoinit); err != nil {
		util.WriteError(w, err, http.StatusBadRequest)
		return
	}

	err := repo.InitializeRepo(&repoinit)

	if err != nil {
		util.WriteError(w, err, http.StatusInternalServerError)
		return
	}
}

func ListRepos(w http.ResponseWriter, r *http.Request) {
	repo, err := data.Fetcher.ListRepo(r.Context())
	if err != nil {
		log.Error(err)
		util.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	response := dto.Response[[]fetch.Repo]{
		Data:  &repo,
		Error: nil,
	}

	util.WriteResponse(w, response, http.StatusOK)
}

func ListBlockStorage(w http.ResponseWriter, r *http.Request) {
	output, err := cmd.Single("lsblk", "-ndo", "NAME,MOUNTPOINT")
	if err != nil {
		log.Error(err)
		util.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	devices := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(*output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		name := fields[0]
		mountpoint := ""
		if len(fields) > 1 {
			mountpoint = fields[1]
		}

		devices[name] = mountpoint
	}

	response := dto.Response[map[string]string]{
		Data:  &devices,
		Error: nil,
	}

	util.WriteResponse(w, response, http.StatusOK)
}
