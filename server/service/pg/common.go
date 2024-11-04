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
}

func Query(auth AuthInfo, cmdKey string, senstive bool, pgPath string, query string) (*string, error) {
	psqlPath := pgPath + "/bin/psql"

	if auth.GetConnectionType() == "host" {
		err := CreatePgPassFile(auth)
		if err != nil {
			return nil, err
		}

		defer RemovePgPassFile()

		return cmd.Single(
			cmdKey+"-host",
			false,
			senstive,
			psqlPath,
			"-t", "-w", "-P", "format=unaligned",
			"-d", "postgres",
			"-U", auth.GetDbUsername(),
			"-h", auth.GetHost(),
			"-p", fmt.Sprintf("%d", auth.GetPort()),
			"-c", query,
		)
	}

	return cmd.Single(
		cmdKey+"-local",
		false,
		senstive,
		"sudo",
		"-u", auth.GetPostgresOsUser(),
		psqlPath,
		"-t",
		"-w",
		"-P", "format=unaligned",
		"-w",
		"-c", query,
	)
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
		return responseerror.Clarify("Failed to do pre-connect housekeeping, please check logs")
	}

	return nil
}

func RemovePgPassFile() error {
	output, err := cmd.Single("remove-pg-pass-file", false, false, "su", "-c", "rm ~/.pgpass")
	if err != nil {
		log.Errorf("Failed to remove pgpass file. output: %s data: %v", util.SafeStringVal(output), err)
		return responseerror.Clarify("Failed to remove temporary pgpass file, please delete it manually")
	}

	return nil
}
