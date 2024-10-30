package repo

import (
	"github.com/go-playground/validator/v10"
	"strings"
)

type PgInitDto struct {
	PostgresPath     string `json:"postgresPath" validate:"required,min=1,excludesall= "`
	Version          int    `json:"version" validate:"required,min=15,max=17"`
	StopPostgres     bool   `json:"stopPostgres"`
	CustomConnection bool   `json:"customConnection"`
	PostgresUser     string `json:"postgresUser" validate:"required_if=CustomConnection false"`
	Host             string `json:"host" validate:"required_if=CustomConnection true,pgInitCon"`
	Port             int    `json:"port" validate:"required_if=CustomConnection true,pgInitCon"`
	Username         string `json:"username" validate:"required_if=CustomConnection true,pgInitCon"`
	Password         string `json:"password" validate:"required_if=CustomConnection true,pgInitCon"`
}

func PgInitValidation(fl validator.FieldLevel) bool {
	dto := fl.Parent().Interface().(PgInitDto)
	field := fl.FieldName()

	if !dto.CustomConnection {
		return true
	}

	switch field {
	case "PostgresUser":
		return dto.CustomConnection || (len(dto.PostgresUser) >= 1 || !strings.Contains(dto.PostgresUser, " "))
	case "Host":
		return !dto.CustomConnection || (len(dto.Host) >= 1 || !strings.Contains(dto.Host, " "))
	case "Port":
		return !dto.CustomConnection || (dto.Port > 0 && dto.Port <= 65535)
	case "Username":
		return !dto.CustomConnection || (len(dto.Username) >= 1 || !strings.Contains(dto.Host, " "))
	case "Password":
		return !dto.CustomConnection || (len(dto.Password) >= 1 || !strings.Contains(dto.Password, " "))
	}

	return true
}
