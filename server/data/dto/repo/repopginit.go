package repo

import (
	"github.com/go-playground/validator/v10"
	"slices"
	"strings"
)

var allowedSslModes = []string{
	"disable",
	"require",
	"verify-ca",
	"verify-full",
}

// PgInitDto This is used to perform some validations for postgres import
type PgInitDto struct {
	PostgresPath   string `json:"postgresPath" validate:"required,min=1,excludesall= "`
	Version        int    `json:"version" validate:"required,min=15,max=17"`
	StopPostgres   bool   `json:"stopPostgres"`
	ConnectionType string `json:"ConnectionType" validate:"oneof=host local,excludesall= "`
	PostgresOsUser string `json:"postgresOsUser,omitempty" validate:"required_if=CustomConnection false"`
	Host           string `json:"host,omitempty" validate:"required_if=CustomConnection true,pgInitCon"`
	Port           int    `json:"port,omitempty" validate:"required_if=CustomConnection true,pgInitCon"`
	SslMode        string `json:"sslMode,omitempty" validate:"required_if=CustomConnection true,pgInitCon"`
	DbUsername     string `json:"dbUsername,omitempty" validate:"required_if=CustomConnection true,pgInitCon"`
	Password       string `json:"password,omitempty" validate:"required_if=CustomConnection true,pgInitCon"`
}

func (pgCheck *PgInitDto) GetPostgresPath() string {
	return pgCheck.PostgresPath
}

func (pgCheck *PgInitDto) GetConnectionType() string {
	return pgCheck.ConnectionType
}

func (pgCheck *PgInitDto) GetPostgresOsUser() string {
	return pgCheck.PostgresOsUser
}

func (pgCheck *PgInitDto) GetHost() string {
	return pgCheck.Host
}

func (pgCheck *PgInitDto) GetPort() int {
	return pgCheck.Port
}

func (pgCheck *PgInitDto) GetDbUsername() string {
	return pgCheck.DbUsername
}

func (pgCheck *PgInitDto) GetPassword() string {
	return pgCheck.Password
}

func (pgCheck *PgInitDto) GetSslMode() string {
	return pgCheck.SslMode
}

func (pgCheck *PgInitDto) IsHostConnection() bool {
	return pgCheck.ConnectionType == "host"
}

func (pgCheck *PgInitDto) GetPgUser() string {
	if pgCheck.IsHostConnection() {
		return pgCheck.DbUsername
	} else {
		return pgCheck.PostgresOsUser
	}
}

func PgInitCheckValidation(fl validator.FieldLevel) bool {
	dto := fl.Parent().Interface().(PgInitDto)
	field := fl.FieldName()

	if !dto.IsHostConnection() {
		return true
	}

	switch field {
	case "PostgresOsUser":
		return len(dto.PostgresOsUser) >= 1 && !strings.Contains(dto.PostgresOsUser, " ")
	case "Host":
		return len(dto.Host) >= 1 && !strings.Contains(dto.Host, " ")
	case "Port":
		return dto.Port > 0 && dto.Port <= 65535
	case "DbUsername":
		return len(dto.DbUsername) >= 1 && !strings.Contains(dto.Host, " ")
	case "Password":
		return len(dto.Password) >= 1 && !strings.Contains(dto.Password, " ")
	case "SslMode":
		return len(dto.SslMode) >= 1 && slices.Contains(allowedSslModes, dto.SslMode)
	}

	return true
}
