package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jamius19/postbranch/internal/db"
	"github.com/jamius19/postbranch/internal/logger"
	"github.com/jamius19/postbranch/internal/runner"
	"github.com/jamius19/postbranch/internal/util"
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

	// ReplicationCheckQuery TODO: Fix potential sql injection
	ReplicationCheckQuery = `SELECT CASE 
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

	HbaConfigQuery = `
		SELECT type as Type, database as Database, user_name AS Username, 
				address AS Address, netmask as Netmask, auth_method AS AuthMethod 
		FROM pg_hba_file_rules 
		WHERE auth_method IN ('trust', 'peer', 'md5', 'scram-sha-256');`

	CreatePostbranchUserQuery = "CREATE USER %s WITH SUPERUSER PASSWORD '%s';"
)

var log = logger.Logger

func GetConnString(pg AuthInfo) string {
	return fmt.Sprintf(
		"user=%s host=%s port=%d password=%s dbname=postgres sslmode=%s",
		pg.GetDbUsername(),
		pg.GetHost(),
		pg.GetPort(),
		pg.GetPassword(),
		pg.GetSslMode(),
	)
}

func Single(auth AuthInfo, query string) (string, error) {
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

func RunQuery(pgInit AuthInfo, query string) (*sql.DB, *sql.Rows, func(), error) {
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

func CreatePgPassFile(auth AuthInfo) error {
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

func GetPsqlCommand(pgOsUser, pgPath, query string, port int32) (string, error) {
	return runner.Single(
		"pg-version-check",
		false,
		false,
		"sudo",
		"-u", pgOsUser,
		filepath.Join(pgPath, "bin", "psql"),
		"-t",
		"-w",
		"-p", fmt.Sprintf("%d", port),
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
	branchName string,
	branchId int32,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	status, err := StartPg(pgPath, mountPath, branchName)
	if err != nil || status == db.BranchPgStopped {
		log.Errorf("Failed to start postgres for branch: %s, error: %v", branchName, err)
		return
	}

	log.Infof("Started Postgres for branch: %s", branchName)

	if err := db.UpdateBranchPgStatus(ctx, branchId, status); err != nil {
		log.Errorf("Failed to update postgres info for branch: %s, error: %v", branchName, err)
		return
	}
}

// StopPgAndUpdateBranch is potentially expensive. It SHOULD always be called as/inside a goroutine.
func StopPgAndUpdateBranch(
	ctx context.Context,
	pgPath,
	mountPath,
	branchName string,
	branchId int32,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	err := StopPg(pgPath, mountPath, branchName, false)
	if err != nil {
		log.Errorf("Failed to start postgres for branch: %s, error: %v", branchName, err)
		return
	}

	log.Infof("Stopped Postgres for branch: %s", branchName)

	if err := db.UpdateBranchPgStatus(ctx, branchId, db.BranchPgStopped); err != nil {
		log.Errorf("Failed to update branch postgres info for branch: %s, error: %v", branchName, err)
		return
	}
}

func StopDangingPg(pgPath, mountPath, branchName string, wg *sync.WaitGroup) {
	defer wg.Done()

	_ = StopPg(pgPath, mountPath, branchName, true)
}

// StartPg is potentially expensive. It SHOULD always be called as/inside a goroutine.
func StartPg(pgPath, mountPath, branchName string) (db.BranchPgStatus, error) {
	log.Infof("Starting Postgres for dataset: %v with postgres path: %v and mount path: %v", branchName, pgPath, mountPath)

	datasetPath := filepath.Join(mountPath, branchName, "data")

	if _, err := os.Stat(filepath.Join(datasetPath, "postmaster.pid")); err == nil {
		log.Warnf("postmaster.pid file exists in the db cluster. deleting it")

		if err := CleanupConfig(datasetPath); err != nil {
			return "", err
		}
	}

	logPath := filepath.Join(mountPath, branchName, "logs", "postgres_start.log")
	pgCtlPath := filepath.Join(pgPath, "bin", "pg_ctl")

	output, err := runner.Single(
		"starting-postgres",
		false,
		false,
		"sudo",
		"-u", PostBranchUser,
		pgCtlPath,
		"start",
		"-l", logPath,
		"-D", datasetPath,
	)

	outputString := strings.Replace(output, "\n", "\\\\", -1)

	if err != nil {
		log.Errorf("Failed to start postgres. output: %s data: %v", outputString, err)
		return db.BranchPgFailed, nil
	}

	log.Infof("Started postgres. output: %s", outputString)

	return db.BranchPgRunning, nil
}

// StopPg is potentially expensive. It SHOULD always be called as/inside a goroutine.
func StopPg(pgPath, mountPath, branchName string, skipLog bool) error {
	if !skipLog {
		log.Infof("Stopping Postgres for branch: %v with postgres path: %v and mount path: %v", branchName, pgPath, mountPath)
	}

	datasetPath := filepath.Join(mountPath, branchName, "data")

	status, err := getPgStatus(pgPath, datasetPath, skipLog)
	if err != nil {
		return err
	}

	if status == db.BranchPgStopped {
		log.Infof("Postgres is already stopped for branch: %v", branchName)
		return nil
	}

	//logPath := filepath.Join(mountPath, branchName, "logs", "postgres_start.log")
	pgCtlPath := filepath.Join(pgPath, "bin", "pg_ctl")

	output, err := runner.Single(
		"stop-postgres",
		skipLog,
		false,
		"sudo",
		"-u", PostBranchUser,
		pgCtlPath,
		"stop",
		"-D", datasetPath,
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

func getPgStatus(pgPath, datasetPath string, skipLog bool) (db.BranchPgStatus, error) {
	pgCtlPath := filepath.Join(pgPath, "bin", "pg_ctl")

	pgCtlCmd := exec.Command(
		"sudo",
		"-u", PostBranchUser,
		pgCtlPath,
		"status", "-D",
		datasetPath,
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

func CleanupConfig(datasetPath string) error {
	if err := util.RemoveFile(filepath.Join(datasetPath, "postgresql.conf")); err != nil {
		log.Errorf("Failed to remove postgresql.conf")
		return err
	}

	if err := util.RemoveFile(filepath.Join(datasetPath, "pg_hba.conf")); err != nil {
		log.Errorf("Failed to remove pg_hba.conf")
		return err
	}

	if err := util.RemoveFile(filepath.Join(datasetPath, "pg_ident.conf")); err != nil {
		log.Errorf("Failed to remove pg_ident.conf")
		return err
	}

	if err := CleanPidFile(datasetPath); err != nil {
		log.Errorf("Failed to remove postmaster.pid")
		return err
	}

	return nil
}

func CleanPidFile(datasetPath string) error {
	if err := util.RemoveFile(filepath.Join(datasetPath, "postmaster.pid")); err != nil {
		log.Errorf("Failed to remove postmaster.pid")
		return err
	}

	return nil
}
