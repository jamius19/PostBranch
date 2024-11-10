package pg

import (
	"database/sql"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
	"os"
	"path/filepath"
)

const (
	Started   = "STARTED"
	Completed = "COMPLETED"
	Failed    = "FAILED"

	ClusterSizeQuery    = "SELECT CEIL(SUM(pg_database_size(datname)) / (1024 * 1024)) AS total_db_size_mb FROM pg_database;"
	VersionQuery        = "SELECT split_part(current_setting('server_version'), '.', 1) AS major_version;"
	SuperUserCheckQuery = `SELECT usesuper FROM pg_user WHERE usename = CURRENT_USER;`

	ConfigFilePathQuery = `SHOW config_file;`
	HbaFilePathQuery    = `SHOW hba_file;`
	IdentFilePathQuery  = `SHOW ident_file;`

	// LocalReplicationCheckQuery TODO: Fix potential sql injection
	LocalReplicationCheckQuery = `SELECT CASE 
           WHEN EXISTS (
               SELECT 1
               FROM pg_hba_file_rules
               WHERE type = 'local'
                 AND 'replication' = ANY(database)
                 AND auth_method IN ('trust', 'peer')
                 AND ('%s' = ANY(user_name) OR 'all' = ANY(user_name))
           ) 
           THEN 'REPLICATION_ALLOWED'
           ELSE 'REPLICATION_NOT_ALLOWED'
       END AS replication_status;`

	// HostReplicationCheckQuery TODO: Fix potential sql injection
	HostReplicationCheckQuery = `SELECT CASE 
           WHEN EXISTS (
               SELECT 1
               FROM pg_hba_file_rules
               WHERE type = 'host'
                 AND 'replication' = ANY(database)
                 AND auth_method IN ('md5', 'scram-sha-256')
                 AND ('%s' = ANY(user_name) OR 'all' = ANY(user_name))
           ) 
           THEN 'REPLICATION_ALLOWED'
           ELSE 'REPLICATION_NOT_ALLOWED'
       END AS replication_status;`
)

var log = logger.Logger

type HostAuthInfo interface {
	GetHost() string
	GetPort() int32
	GetDbUsername() string
	GetPassword() string
	GetSslMode() string
}

func GetConnString(pg HostAuthInfo) string {
	return fmt.Sprintf(
		"user=%s host=%s port=%d password=%s dbname=postgres sslmode=%s",
		pg.GetDbUsername(),
		pg.GetHost(),
		pg.GetPort(),
		pg.GetPassword(),
		pg.GetSslMode(),
	)
}

func Single(auth HostAuthInfo, query string) (string, error) {
	var result string

	_, rows, cleanup, err := RunQuery(auth, query)
	if err != nil {
		return "", err
	}
	defer cleanup()

	for rows.Next() {
		err := rows.Scan(&result)
		if err != nil {
			return "", fmt.Errorf("failed to scan postgres. error: %v", err)
		}
	}

	return result, nil
}

func SingleLocal(pgOsUser, pgPath, query string) (string, error) {
	var result string

	output, err := GetPsqlCommand(pgOsUser, pgPath, query)
	result = util.TrimmedString(output)

	if err != nil || result == cmd.EmptyOutput {
		return "", fmt.Errorf("failed to scan postgres. error: %v", err)
	}

	return result, nil
}

func RunQuery(pgInit HostAuthInfo, query string) (*sql.DB, *sql.Rows, func(), error) {
	cleanup := func() {}
	log.Debugf("Running query: %s", query)

	db, err := sql.Open("postgres", GetConnString(pgInit))
	if err != nil {
		log.Errorf("Failed to open db: %v", err)
		return nil, nil, cleanup, err
	}

	rows, err := db.Query(query)
	if err != nil {
		log.Errorf("Failed to run query: %s, error: %v", query, err)

		if db != nil {
			if err := db.Close(); err != nil {
				log.Errorf("Failed to close db: %v", err)
			}
		}

		return nil, nil, cleanup, err
	}

	cleanup = func() {
		if db != nil {
			if err := db.Close(); err != nil {
				log.Errorf("Failed to close db: %v", err)
			}
		}

		if rows != nil {
			if err := rows.Close(); err != nil {
				log.Errorf("Failed to close rows: %v", err)
			}
		}
	}
	return db, rows, cleanup, err
}

func CreatePgPassFile(auth HostAuthInfo) error {
	pgPassContent := fmt.Sprintf(
		`%s:%d:*:%s:%s`,
		auth.GetHost(),
		auth.GetPort(),
		auth.GetDbUsername(),
		auth.GetPassword(),
	)
	pgPassPath := filepath.Join(os.ExpandEnv("$HOME"), ".pgpass")
	err := os.WriteFile(pgPassPath, []byte(pgPassContent), 0600)

	if err != nil {
		return fmt.Errorf("failed to create pgpass file. error: %v", err)
	}

	return nil
}

func RemovePgPassFile() error {
	err := os.Remove(filepath.Join(os.ExpandEnv("$HOME"), ".pgpass"))

	if err != nil {
		return fmt.Errorf("failed to remove pgpass file. error: %v", err)
	}

	return nil
}

func GetPsqlCommand(pgOsUser, pgPath, query string) (*string, error) {
	return cmd.Single(
		"pg-version-check",
		false,
		false,
		"sudo",
		"-u", pgOsUser,
		filepath.Join(pgPath, "bin", "psql"),
		"-t",
		"-w",
		"-P", "format=unaligned",
		"-w",
		"-c", query,
	)
}
