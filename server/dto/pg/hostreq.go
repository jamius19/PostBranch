package pg

import (
	"fmt"
)

// HostImportReqDto This is used to perform some validations for postgres import
type HostImportReqDto struct {
	PostgresPath string `json:"postgresPath" validate:"required,min=1,excludesall= "`
	Version      int32  `json:"version" validate:"required,min=15,max=17"`
	Host         string `json:"host,omitempty" validate:"required,min=1,excludesall= "`
	Port         int32  `json:"port,omitempty" validate:"required,min=1,max=65535"`
	SslMode      string `json:"sslMode,omitempty" validate:"required,oneof=disable require verify-ca verify-full,excludesall= "`
	DbUsername   string `json:"dbUsername,omitempty" validate:"required,min=1,excludesall= "`
	Password     string `json:"password,omitempty" validate:"required,min=1,excludesall= "`
}

func (pgInit *HostImportReqDto) GetPostgresPath() string {
	return pgInit.PostgresPath
}

func (pgInit *HostImportReqDto) GetHost() string {
	return pgInit.Host
}

func (pgInit *HostImportReqDto) GetPort() int32 {
	return pgInit.Port
}

func (pgInit *HostImportReqDto) GetDbUsername() string {
	return pgInit.DbUsername
}

func (pgInit *HostImportReqDto) GetPassword() string {
	return pgInit.Password
}

func (pgInit *HostImportReqDto) GetSslMode() string {
	return pgInit.SslMode
}

func (pgInit *HostImportReqDto) IsHostConnection() bool {
	return true
}

func (pgInit *HostImportReqDto) String() string {
	return fmt.Sprintf(
		"{%s %d %s %d %s %s *****}",
		pgInit.PostgresPath,
		pgInit.Version,
		pgInit.Host,
		pgInit.Port,
		pgInit.SslMode,
		pgInit.DbUsername,
	)
}
