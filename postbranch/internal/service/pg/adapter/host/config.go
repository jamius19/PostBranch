package host

import (
	"fmt"
	pg2 "github.com/jamius19/postbranch/internal/service/pg"
)

func getHbaFileConfig(auth pg2.AuthInfo) ([]pg2.HbaConfig, error) {
	var results []pg2.HbaConfig

	_, rows, cleanup, err := pg2.RunQuery(auth, pg2.HbaConfigQuery)

	if err != nil {
		return nil, err
	}
	defer cleanup()

	for rows.Next() {
		var result pg2.HbaConfig
		err := rows.Scan(&result.Type, &result.Database, &result.Username, &result.Address, &result.Netmask, &result.AuthMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to scan postgres. error: %v", err)
		}
		results = append(results, result)
	}

	return results, nil
}
