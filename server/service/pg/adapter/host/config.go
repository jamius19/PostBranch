package host

import (
	"fmt"
	"github.com/jamius19/postbranch/service/credential"
	"github.com/jamius19/postbranch/service/pg"
)

func getHbaFileConfig(auth pg.AuthInfo) ([]pg.HbaConfig, error) {
	var results []pg.HbaConfig

	_, rows, cleanup, err := pg.RunQuery(auth, pg.HbaConfigQuery)

	if err != nil {
		return nil, err
	}
	defer cleanup()

	for rows.Next() {
		var result pg.HbaConfig
		err := rows.Scan(&result.Type, &result.Database, &result.Username, &result.Address, &result.Netmask, &result.AuthMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to scan postgres. error: %v", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func addSuperuser(previousSuperUser, previousPassword string, newPort int32) (string, error) {
	password := credential.GeneratePassword()
	query := fmt.Sprintf(pg.CreatePostbranchUserQuery, pg.PostBranchDbUser, password)

	authInfo := pg.NewAuthInfo(
		"localhost",
		newPort,
		previousSuperUser,
		previousPassword,
		"disable",
	)

	_, err := pg.Single(&authInfo, query)
	if err != nil {
		return "", err
	}

	return password, nil
}
