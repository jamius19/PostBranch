package route

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/jamius19/postbranch/dto"
	"github.com/jamius19/postbranch/logger"
	zfsService "github.com/jamius19/postbranch/service/repo/zfs"
	"github.com/jamius19/postbranch/util"
	"net/http"
)

var validate = validator.New()
var log = logger.Logger

func InitializeRepo(w http.ResponseWriter, r *http.Request) {
	var repo dto.RepoInit

	if err := json.NewDecoder(r.Body).Decode(&zfsVirtualInit); err != nil {
		util.WriteError(w, err, http.StatusBadRequest)
	}

	if err := validate.Struct(zfsVirtualInit); err != nil {
		util.WriteError(w, err, http.StatusBadRequest)
	}

	err := zfsService.InitializeVirtual(&zfsVirtualInit)

	if err != nil {
		util.WriteError(w, err, http.StatusInternalServerError)
		return
	}
}
