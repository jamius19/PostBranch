package pg

import (
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
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
	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()

	pgPassContent := fmt.Sprintf(
		`"%s:%d:*:%s:%s"`,
		auth.GetHost(),
		auth.GetPort(),
		auth.GetDbUsername(),
		auth.GetPassword(),
	)

	// TODO: Remove this
	cmds.Set(
		"who-am-i",
		cmd.Get("whoami"),
	)

	cmds.Set(
		"touch-pg-pass-file",
		cmd.Get("su", "-c", "touch ~/.pgpass"),
	)

	cmds.Set(
		"chmod-pg-pass-file",
		cmd.Get("su", "-c", "chmod 600 ~/.pgpass"),
	)

	pgPassCommand := fmt.Sprintf("echo %s > $HOME/.pgpass", pgPassContent)
	cmds.Set(
		"create-pg-pass-file",
		cmd.GetSensitive(
			"su",
			"-c", pgPassCommand,
		),
	)

	output, err := cmd.Multi(cmds)

	if err != nil {
		errStr := cmd.GetError(output)
		log.Errorf("Failed to create pgpass file. output: %s data: %v", errStr, err)
		return responseerror.From("Failed to do pre-connect housekeeping, please check logs")
	}

	return nil
}

func RemovePgPassFile() error {
	output, err := cmd.Single("remove-pg-pass-file", false, false, "su", "-c", "rm ~/.pgpass")
	if err != nil {
		log.Errorf("Failed to remove pgpass file. output: %s data: %v", util.SafeStringVal(output), err)
		return responseerror.From("Failed to remove temporary pgpass file, please delete it manually")
	}

	return nil
}
