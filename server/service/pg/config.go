package pg

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/logger"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type HbaConfig struct {
	Type       string
	Database   string
	Username   string
	Address    *string
	Netmask    *string
	AuthMethod string
}

func getHbaFileConfig(auth HostAuthInfo) ([]HbaConfig, error) {
	var results []HbaConfig

	_, rows, cleanup, err := RunQuery(auth, `
		SELECT type as Type, database as Database, user_name AS Username, address AS Address, netmask as Netmask, auth_method AS AuthMethod 
		FROM pg_hba_file_rules 
		WHERE auth_method IN ('trust', 'peer', 'md5', 'scram-sha-256');`,
	)

	if err != nil {
		return nil, err
	}
	defer cleanup()

	for rows.Next() {
		var result HbaConfig
		err := rows.Scan(&result.Type, &result.Database, &result.Username, &result.Address, &result.Netmask, &result.AuthMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to scan postgres. error: %v", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func GetPgPort(ctx context.Context) (int32, error) {
	log.Info("Getting postgres port")

	ports, err := db.GetBranchPorts(ctx)
	if err != nil {
		return -1, err
	}

	for i := int32(5450); i < 8542; i++ {
		if slices.Contains(ports, i) {
			continue
		}

		portAvailable, err := checkPort(i)
		if err != nil {
			continue
		}

		if portAvailable {
			return i, nil
		}
	}

	logger.Logger.Errorf("No port available")
	return -1, err
}

func WritePostgresConfig(port int32, repoName, logPath, datasetPath string) error {
	log.Info("Writing postgres config")

	file, err := os.Create(filepath.Join(datasetPath, "postgresql.conf"))
	if err != nil {
		return fmt.Errorf("failed to create postgres config file: %w", err)
	}
	defer file.Close()

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("port = %d\n", port))
	builder.WriteString("listen_addresses = '*'\n")
	builder.WriteString("unix_socket_directories = '/var/run/postbranch'\n")
	builder.WriteString(fmt.Sprintf("max_connections = %d\n", MaxConnection))

	// This is set because we'll be using ZFS filesystem for Postgres data
	builder.WriteString("full_page_writes = off\n")

	builder.WriteString(fmt.Sprintf("log_directory = '%s'\n", logPath))
	builder.WriteString(fmt.Sprintf("log_filename = '%s_%s'\n", repoName, "%Y-%m-%d_%H-%M-%S.log"))
	builder.WriteString("logging_collector = on\n")
	builder.WriteString("log_rotation_size = 10MB\n")
	builder.WriteString("log_file_mode = 0600\n")
	builder.WriteString("log_checkpoints = on\n")

	_, err = file.WriteString(builder.String())
	if err != nil {
		return fmt.Errorf("failed to write to postgres config file: %w", err)
	}

	return nil
}

func WritePgHbaConfig(auth HostAuthInfo, datasetPath string) error {
	log.Info("Writing pg hba file")
	hbaConfigs, err := getHbaFileConfig(auth)
	if err != nil {
		return err
	}

	var builder strings.Builder

	// Iterate over the configs and build each line
	for _, config := range hbaConfigs {
		// Sanitize Database and Username fields to remove `{}` if present
		database := strings.Trim(config.Database, "{}")
		username := strings.Trim(config.Username, "{}")

		line := config.Type + "\t" + database + "\t" + username

		// Append Address and Netmask if provided
		if config.Address != nil {
			line += "\t" + *config.Address
			if config.Netmask != nil {
				line += "\t" + *config.Netmask
			}
		} else {
			line += "\t"
		}

		// Append the authentication method
		line += "\t" + config.AuthMethod + "\n"
		builder.WriteString(line)
	}

	// Write to file
	file, err := os.Create(filepath.Join(datasetPath, "pg_hba.conf"))
	if err != nil {
		return fmt.Errorf("failed to create hba file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(builder.String())
	if err != nil {
		return fmt.Errorf("failed to write to hba file: %w", err)
	}

	return nil
}

func checkPort(port int32) (bool, error) {
	host := ":" + strconv.Itoa(int(port))

	server, err := net.Listen("tcp", host)

	if err != nil {
		return false, nil
	}

	// close the server
	err = server.Close()
	if err != nil {
		return false, err
	}

	// we successfully used and closed the port
	// so it's now available to be used again
	return true, nil

}
