package route

import (
	"encoding/json"
	"github.com/jamius19/postbranch/dto"
	pgDto "github.com/jamius19/postbranch/dto/pg"
	"github.com/jamius19/postbranch/service/pg/adapter/host"
	"github.com/jamius19/postbranch/service/validation"
	"github.com/jamius19/postbranch/util"
	"net/http"
)

func ValidateHostPg(w http.ResponseWriter, r *http.Request) {
	log.Info("Starting validation of host pg")

	var pgInit pgDto.HostImportReqDto
	if err := json.NewDecoder(r.Body).Decode(&pgInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(pgInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	err := host.Validate(pgInit)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	clusterSizeInMb, err := host.GetClusterSize(pgInit)
	if err != nil {
		util.WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	pgInitWithSize := pgDto.ValidationResponseDto[pgDto.HostImportReqDto]{
		PgConfig:        pgInit,
		ClusterSizeInMb: clusterSizeInMb,
	}

	response := dto.Response[pgDto.ValidationResponseDto[pgDto.HostImportReqDto]]{
		Data:  &pgInitWithSize,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}
