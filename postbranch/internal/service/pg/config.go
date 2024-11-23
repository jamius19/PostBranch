package pg

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/internal/db"
	"github.com/jamius19/postbranch/internal/logger"
	"github.com/jamius19/postbranch/internal/util"
	"net"
	"os"
	"path/filepath"
	"slices"
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
			log.Infof("Found available port: %d", i)
			return i, nil
		}
	}

	logger.Logger.Errorf("No port available")
	return -1, err
}

func UpdatePostgresConfig(datasetPath, configName, configVal string) error {
	log.Info("Updating postgres config")
	configPath := filepath.Join(datasetPath, "postgresql.conf")

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read postgres config file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, configName) {
			lines[i] = fmt.Sprintf("%s = %s", configName, configVal)
			break
		}
	}

	err = os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to write postgres config file: %w", err)
	}

	return nil

}

func WritePostgresConfig(port int32, repoName, branchName, logPath, datasetPath string) error {
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
	builder.WriteString("password_encryption = 'scram-sha-256'\n")

	builder.WriteString(fmt.Sprintf("log_directory = '%s'\n", logPath))
	builder.WriteString(fmt.Sprintf("log_filename = '%s_%s__%s.log'\n", repoName, branchName, "%Y-%m-%d_%H-%M-%S"))
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

func WritePgHbaConfig(hbaConfigs []HbaConfig, datasetPath string) error {
	log.Info("Writing pg hba file")

	var builder strings.Builder

	// Iterate over the configs and build each line
	for _, config := range hbaConfigs {
		// Sanitize Database and Username fields to remove `{}` if present
		database := strings.Trim(config.Database, "{}")
		username := strings.Trim(config.Username, "{}")

		line := fmt.Sprintf("%-15s %-50s %-50s %-20s %-15s %-15s\n",
			config.Type,
			database,
			username,
			util.TrimmedString(config.Address),
			util.TrimmedString(config.Netmask),
			config.AuthMethod,
		)

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
	host := fmt.Sprintf(":%d", port)

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
