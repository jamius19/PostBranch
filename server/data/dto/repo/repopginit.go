package repo

import (
	"fmt"
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

func (pgInit *PgInitDto) GetPostgresPath() string {
	return pgInit.PostgresPath
}

func (pgInit *PgInitDto) GetConnectionType() string {
	return pgInit.ConnectionType
}

func (pgInit *PgInitDto) GetPostgresOsUser() string {
	return pgInit.PostgresOsUser
}

func (pgInit *PgInitDto) GetHost() string {
	return pgInit.Host
}

func (pgInit *PgInitDto) GetPort() int {
	return pgInit.Port
}

func (pgInit *PgInitDto) GetDbUsername() string {
	return pgInit.DbUsername
}

func (pgInit *PgInitDto) GetPassword() string {
	return pgInit.Password
}

func (pgInit *PgInitDto) GetSslMode() string {
	return pgInit.SslMode
}

func (pgInit *PgInitDto) IsHostConnection() bool {
	return pgInit.ConnectionType == "host"
}

func (pgInit *PgInitDto) GetPgUser() string {
	if pgInit.IsHostConnection() {
		return pgInit.DbUsername
	} else {
		return pgInit.PostgresOsUser
	}
}

func (pgInit *PgInitDto) String() string {
	if pgInit.IsHostConnection() {
		return fmt.Sprintf(
			"{%s %d %t %s %s %s %d %s %s *****}",
			pgInit.PostgresPath,
			pgInit.Version,
			pgInit.StopPostgres,
			pgInit.ConnectionType,
			pgInit.PostgresOsUser,
			pgInit.Host,
			pgInit.Port,
			pgInit.SslMode,
			pgInit.DbUsername,
		)
	}

	return fmt.Sprintf(
		"{%s %d %t %s %s}",
		pgInit.PostgresPath,
		pgInit.Version,
		pgInit.StopPostgres,
		pgInit.ConnectionType,
		pgInit.PostgresOsUser,
	)
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
