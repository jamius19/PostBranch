package pg

import (
	"fmt"
	"github.com/jamius19/postbranch/logger"
	"os"
)

var log = logger.Logger

type AuthInfo interface {
	GetPostgresPath() string
	GetConnectionType() string
	GetPostgresOsUser() string
	GetHost() string
	GetPort() int
	GetDbUsername() string
	GetPassword() string
	GetSslMode() string
}

func CreatePgPassFile(auth AuthInfo) error {
	pgPassContent := fmt.Sprintf(
		`%s:%d:*:%s:%s`,
		auth.GetHost(),
		auth.GetPort(),
		auth.GetDbUsername(),
		auth.GetPassword(),
	)

	err := os.WriteFile(os.ExpandEnv("$HOME/.pgpass"), []byte(pgPassContent), 0600)

	if err != nil {
		return fmt.Errorf("failed to create pgpass file. error: %v", err)
	}

	return nil
}

func RemovePgPassFile() error {
	err := os.Remove(os.ExpandEnv("$HOME/.pgpass"))

	if err != nil {
		return fmt.Errorf("failed to remove pgpass file. error: %v", err)
	}

	return nil
}
