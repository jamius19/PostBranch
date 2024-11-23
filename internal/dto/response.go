package dto

import (
	"encoding/json"
	"reflect"
)

type Response[T any] struct {
	Data   *T
	Error  *[]string
	IsList bool
}

func (r Response[T]) MarshalJSON() ([]byte, error) {
	var data any

	if r.IsList {
		if r.Data != nil {
			v := reflect.ValueOf(*r.Data)
			if v.Kind() == reflect.Slice {
				if v.Len() > 0 {
					data = r.Data
				} else {
					data = []T{}
				}
			} else {
				data = r.Data
			}
		} else {
			data = []T{}
		}
	} else {
		data = r.Data
	}

	return json.Marshal(&struct {
		Data  any       `json:"data"`
		Error *[]string `json:"errors"`
	}{
		Data:  data,
		Error: r.Error,
	})
}
