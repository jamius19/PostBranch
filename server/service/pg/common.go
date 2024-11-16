package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

const (
	PostBranchUser = "postbranch"
	MaxConnection  = 20

	ClusterSizeQuery    = "SELECT CEIL(SUM(pg_database_size(datname)) / (1024 * 1024)) AS total_db_size_mb FROM pg_database;"
	VersionQuery        = "SELECT split_part(current_setting('server_version'), '.', 1) AS major_version;"
	SuperUserCheckQuery = `SELECT usesuper FROM pg_user WHERE usename = CURRENT_USER;`

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

	if rows.Next() {
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

	if err != nil || output == cmd.EmptyOutput {
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

func GetPsqlCommand(pgOsUser, pgPath, query string) (string, error) {
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

// StartPgAndUpdateBranch is potentially expensive. It SHOULD always be called as/inside a goroutine.
func StartPgAndUpdateBranch(
	ctx context.Context,
	pgPath,
	mountPath,
	datasetName string,
	datasetId int32,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	status, err := StartPg(pgPath, mountPath, datasetName)
	if err != nil || status == db.BranchPgStopped {
		log.Errorf("Failed to start postgres for datasetName: %s, error: %v", datasetName, err)
		return
	}

	log.Infof("Started Postgres for datasetName: %s", datasetName)

	if err := db.UpdateBranchPgStatus(ctx, datasetId, status); err != nil {
		log.Errorf("Failed to update branch postgres info for datasetName: %s, error: %v", datasetName, err)
		return
	}
}

// StopPgAndUpdateBranch is potentially expensive. It SHOULD always be called as/inside a goroutine.
func StopPgAndUpdateBranch(
	ctx context.Context,
	pgPath,
	mountPath,
	datasetName string,
	datasetId int32,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	err := StopPg(pgPath, mountPath, datasetName, false)
	if err != nil {
		log.Errorf("Failed to start postgres for datasetName: %s, error: %v", datasetName, err)
		return
	}

	log.Infof("Stopped Postgres for datasetName: %s", datasetName)

	if err := db.UpdateBranchPgStatus(ctx, datasetId, db.BranchPgStopped); err != nil {
		log.Errorf("Failed to update branch postgres info for datasetName: %s, error: %v", datasetName, err)
		return
	}
}

func StopDangingPg(pgPath, mountPath, datasetName string, wg *sync.WaitGroup) {
	defer wg.Done()

	err := StopPg(pgPath, mountPath, datasetName, true)
	if err != nil {
		log.Infof("No dangling postgres found for dataset: %v", datasetName)
	}
}

// StartPg is potentially expensive. It SHOULD always be called as/inside a goroutine.
func StartPg(pgPath, mountPath, datasetName string) (db.BranchPgStatus, error) {
	log.Infof("Starting Postgres for dataset: %v with postgres path: %v and mount path: %v", datasetName, pgPath, mountPath)

	mainDatasetPath := filepath.Join(mountPath, datasetName, "data")

	if _, err := os.Stat(filepath.Join(mainDatasetPath, "postmaster.pid")); err == nil {
		log.Warnf("postmaster.pid file exists in the db cluster. deleting it")

		if err := cleanPidFile(mainDatasetPath); err != nil {
			return "", err
		}
	}

	logPath := filepath.Join(mountPath, datasetName, "logs", "postgres_start.log")
	pgCtlPath := filepath.Join(pgPath, "bin", "pg_ctl")

	output, err := cmd.Single(
		"starting-postgres",
		false,
		false,
		"sudo",
		"-u", PostBranchUser,
		pgCtlPath,
		"start",
		"-l", logPath,
		"-D", mainDatasetPath,
	)

	outputString := strings.Replace(output, "\n", "\\\\", -1)

	if err != nil {
		log.Errorf("Failed to start postgres. output: %s data: %v", outputString, err)
		return "", err
	}

	log.Infof("Started postgres. output: %s", outputString)

	status, err := getPgStatus(pgPath, mainDatasetPath, false)
	if err != nil {
		return "", err
	}

	// As we just started the postgres, we expect the status to be running
	if status == db.BranchPgStopped {
		return db.BranchPgFailed, nil
	}

	log.Infof("Postgres is ready. status: %s", status)
	return status, nil
}

// StopPg is potentially expensive. It SHOULD always be called as/inside a goroutine.
func StopPg(pgPath, mountPath, datasetName string, skipLog bool) error {
	if !skipLog {
		log.Infof("Stopping Postgres for dataset: %v with postgres path: %v and mount path: %v", datasetName, pgPath, mountPath)
	}

	mainDatasetPath := filepath.Join(mountPath, datasetName, "data")

	status, err := getPgStatus(pgPath, mainDatasetPath, skipLog)
	if err != nil {
		return err
	}

	if status == db.BranchPgStopped {
		log.Infof("Postgres is already stopped for dataset: %v", datasetName)
		return nil
	}

	//logPath := filepath.Join(mountPath, datasetName, "logs", "postgres_start.log")
	pgCtlPath := filepath.Join(pgPath, "bin", "pg_ctl")

	output, err := cmd.Single(
		"stop-postgres",
		skipLog,
		false,
		"sudo",
		"-u", PostBranchUser,
		pgCtlPath,
		"stop",
		"-D", mainDatasetPath,
	)

	if err != nil {
		if !skipLog {
			log.Errorf("Failed to stop postgres. output: %s data: %v", output, err)
		}

		return err
	}

	if skipLog {
		log.Infof("Stopped postgres. output: %s", output)
	}

	return nil
}

func getPgStatus(pgPath, mainDatasetPath string, skipLog bool) (db.BranchPgStatus, error) {
	pgCtlPath := filepath.Join(pgPath, "bin", "pg_ctl")

	pgCtlCmd := exec.Command(
		"sudo",
		"-u", PostBranchUser,
		pgCtlPath,
		"status", "-D",
		mainDatasetPath,
	)

	pgCtlCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	output, err := pgCtlCmd.CombinedOutput()
	outputStr := strings.Replace(string(output), "\n", "\\\\", -1)

	if err == nil {
		return db.BranchPgRunning, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// Exit code 3 means that postgres is not running
		// See https://www.postgresql.org/docs/current/app-pg-ctl.html
		if exitErr.ExitCode() == 3 {
			return db.BranchPgStopped, nil
		}
	}

	// All other cases, it's an error
	if !skipLog {
		log.Errorf("Failed to run pg_ctl status. output: %s, error: %v", outputStr, err)
	}

	return "", err
}

func ValidatePgPath(pgPath string) error {
	pgBaseBackupPath := filepath.Join(pgPath, "bin", "pg_basebackup")
	postgresPath := filepath.Join(pgPath, "bin", "postgres")
	pgCtlPath := filepath.Join(pgPath, "bin", "pg_ctl")

	if _, err := os.Stat(pgBaseBackupPath); errors.Is(err, os.ErrNotExist) {
		return responseerror.From("Invalid Postgres path, pg_basebackup not found")
	}

	if _, err := os.Stat(postgresPath); errors.Is(err, os.ErrNotExist) {
		return responseerror.From("Invalid Postgres path, postgres not found")
	}

	if _, err := os.Stat(pgCtlPath); errors.Is(err, os.ErrNotExist) {
		return responseerror.From("Invalid Postgres path, pg_ctl not found")
	}

	return nil
}

func CleanupConfig(mainDatasetPath string) error {
	if err := util.RemoveFile(filepath.Join(mainDatasetPath, "postgresql.conf")); err != nil {
		log.Errorf("Failed to remove postgresql.conf")
		return err
	}

	if err := util.RemoveFile(filepath.Join(mainDatasetPath, "pg_hba.conf")); err != nil {
		log.Errorf("Failed to remove pg_hba.conf")
		return err
	}

	if err := util.RemoveFile(filepath.Join(mainDatasetPath, "pg_ident.conf")); err != nil {
		log.Errorf("Failed to remove pg_ident.conf")
		return err
	}

	if err := cleanPidFile(mainDatasetPath); err != nil {
		log.Errorf("Failed to remove postmaster.pid")
		return err
	}

	return nil
}

func cleanPidFile(mainDatasetPath string) error {
	if err := util.RemoveFile(filepath.Join(mainDatasetPath, "postmaster.pid")); err != nil {
		log.Errorf("Failed to remove postmaster.pid")
		return err
	}

	return nil
}
