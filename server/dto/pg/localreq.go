package pg

import (
	"fmt"
)

// LocalImportReqDto This is used to perform some validations for postgres import
type LocalImportReqDto struct {
	PostgresPath   string `json:"postgresPath" validate:"required,min=1,excludesall= "`
	Version        int32  `json:"version" validate:"required,min=15,max=17"`
	PostgresOsUser string `json:"postgresOsUser,omitempty" validate:"required"`
}

func (pgInit *LocalImportReqDto) GetPostgresPath() string {
	return pgInit.PostgresPath
}

func (pgInit *LocalImportReqDto) GetPostgresOsUser() string {
	return pgInit.PostgresOsUser
}

func (pgInit *LocalImportReqDto) IsHostConnection() bool {
	return false
}

func (pgInit *LocalImportReqDto) String() string {
	return fmt.Sprintf(
		"{%s %d %s}",
		pgInit.PostgresPath,
		pgInit.Version,
		pgInit.PostgresOsUser,
	)
}
