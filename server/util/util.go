package util

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jamius19/postbranch/data/dto"
	"github.com/jamius19/postbranch/web/responseerror"
	"net/http"
	"strings"
)

func WriteResponse(w http.ResponseWriter, r *http.Request, data any, code int) {
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

func SafeStringVal(str *string) string {
	var output = "<nil>"

	if str != nil && *str != "" {
		output = *str
	}

	return output
}

func GetNullableInt64(nullValue *sql.NullInt64) *int64 {
	if nullValue.Valid {
		return &nullValue.Int64
	} else {
		return nil
	}
}

func StringVal[T ~int | ~int8 | ~int16 | ~int32 |
~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 |
~uint64 | ~float32 | ~float64](num T) string {
	
	return fmt.Sprint(num)
}
