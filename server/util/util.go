package util

import (
	"encoding/json"
	"github.com/jamius19/postbranch/data/dto"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
	"strings"
)

func WriteResponse(w http.ResponseWriter, r *http.Request, data interface{}, code int) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		WriteError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonData)
}

func WriteError(w http.ResponseWriter, r *http.Request, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := dto.Response[any]{
		Data:  nil,
		Error: responseerror.AddAndGetErrors(r.Context(), err.Error()),
	}

	responseJson, _ := json.Marshal(response)

	w.Write(responseJson)
}

func SplitPath(filepath string) (filename string, path string) {
	lastIndex := strings.LastIndex(filepath, "/")
	if lastIndex == -1 {
		return filepath, ""
	}

	filename = filepath[lastIndex+1:]
	path = filepath[:lastIndex+1]

	if len(path) > 1 {
		path = path[:len(path)-1]
	}

	return filename, path
}

func StringVal(str *string) string {
	var output = "<nil>"

	if str != nil && *str != "" {
		output = *str
	}

	return output
}
