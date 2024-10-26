package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/dto"
	"github.com/jamius19/postbranch/service"
	"github.com/jamius19/postbranch/util"
	"net/http"
)

func AddSettings(w http.ResponseWriter, r *http.Request) {
	//key := chi.URLParam(r, "key")
}

func GetSettings(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	setting, err := service.GetSetting(r.Context(), key)
	if err != nil {
		util.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	if setting == nil {
		response := dto.Response[dto.Setting]{
			Data:  nil,
			Error: dto.GetError("setting not found"),
		}

		util.WriteResponse(w, response, http.StatusOK)
		return
	}

	response := dto.Response[dto.Setting]{
		Data:  dto.GetSetting(setting),
		Error: nil,
	}

	util.WriteResponse(w, response, http.StatusOK)

}
