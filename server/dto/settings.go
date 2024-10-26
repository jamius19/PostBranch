package dto

import (
	"encoding/json"
	"github.com/jamius19/postbranch/data/fetch"
	"time"
)

type Setting struct {
	Key       string      `json:"key"`
	Value     string      `json:"value"`
	Json      interface{} `json:"json"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func GetSetting(setting *fetch.Setting) *Setting {
	var jsonData interface{}

	if setting.Json.Valid {
		json.Unmarshal([]byte(setting.Json.String), &jsonData)
	}

	return &Setting{
		Key:       setting.Key,
		Value:     setting.Value,
		Json:      jsonData,
		CreatedAt: setting.CreatedAt,
		UpdatedAt: setting.UpdatedAt,
	}
}
