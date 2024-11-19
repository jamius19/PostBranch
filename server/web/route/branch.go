package route

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/dto"
	"github.com/jamius19/postbranch/dto/repo"
	repoSvc "github.com/jamius19/postbranch/service/repo"
	"github.com/jamius19/postbranch/service/validation"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
	"strconv"
)

func CreateBranch(w http.ResponseWriter, r *http.Request) {
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

	var branchInit repo.BranchInit
	if err := json.NewDecoder(r.Body).Decode(&branchInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := validation.Validate(branchInit); err != nil {
		util.WriteError(w, r, err, http.StatusBadRequest)
		return
	}

	repoDetail, err := db.GetRepo(r.Context(), repoId)

	// TODO: Add validation for parent branch status

	branch, err := repoSvc.CreateBranch(r.Context(), repoDetail, branchInit)
	if err != nil {
		util.WriteError(
			w,
			r,
			responseerror.From("Failed to create branch"),
			http.StatusBadRequest,
		)

		return
	}

	response := dto.Response[model.Branch]{
		Data:  &branch,
		Error: nil,
	}

	util.WriteResponse(w, r, response, http.StatusOK)
}
