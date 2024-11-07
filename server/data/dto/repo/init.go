package repo

import (
	"github.com/go-playground/validator/v10"
)

type InitDto struct {
	Name     string `json:"name" validate:"required,min=1,excludesall= "`
	Path     string `json:"path" validate:"required,min=1,excludesall= "`
	RepoType string `json:"repoType" validate:"oneof=block virtual,excludesall= "`
	SizeInMb int64  `json:"sizeInMb" validate:"required_if=RepoType virtual,initCon"`
	PgInitResponseDto
}

func InitValidation(fl validator.FieldLevel) bool {
	dto := fl.Parent().Interface().(InitDto)
	field := fl.FieldName()
	if dto.RepoType == "block" {
		return true
	}

	switch field {
	case "SizeInMb":
		return dto.SizeInMb >= 256
	}

	return true
}
