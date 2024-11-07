package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/pg"
	"github.com/jamius19/postbranch/service/repo/zfs"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	_ "github.com/lib/pq"
	"io/fs"
	"os"
	"os/user"
	"strconv"
	"strings"
)

var log = logger.Logger

func Import(ctx context.Context, pgInit *repo.PgInitDto, repo *dao.Repo, pgInfo *dao.Pg, pool *dao.ZfsPool) error {
	pgBaseBackupPath := pgInit.PostgresPath + "/bin/pg_basebackup"
	if _, err := os.Stat(pgBaseBackupPath); errors.Is(err, fs.ErrNotExist) {
		return responseerror.From("Invalid Postgres path, please check the path")
	}

	err := checkOsUser(pgInit.PostgresOsUser)
	if err != nil {
		return err
	}

	err = checkPgVersion(pgInit)
	if err != nil {
		return err
	}

	err = checkPgSuperuser(pgInit)
	if err != nil {
		return err
	}

	err = checkPgReplication(pgInit)
	if err != nil {
		return err
	}

	// Get the main dataset for importing Postgres data
	dataset, err := data.Fetcher.GetDatasetByName(ctx, pool.Name+"/main")
	if err != nil {
		log.Errorf("Dataset not found for repo: %v and pool: %v", repo, pool)
		return responseerror.From("Associated Dataset not found")
	}

	createdPg, err := insertPgEntry(ctx, pgInit, repo, pgInfo)
	if err != nil {
		return err
	}

	go copyPostgresData(pgInit, repo, pool, &dataset, &createdPg)

	return nil
}

func checkOsUser(username string) error {
	_, err := user.Lookup(username)
	if err != nil {
		log.Errorf("User %s not found", username)
		return responseerror.From("Invalid Postgres OS user")
	}

	return nil
}

func checkPgVersion(pgInit *repo.PgInitDto) error {
	versionQuery := "SELECT split_part(current_setting('server_version'), '.', 1) AS major_version;"
	var version string

	if pgInit.IsHostConnection() {
		_, rows, cleanup, err := dao.RunQuery(pgInit, versionQuery)
		if err != nil {
			return err
		}
		defer cleanup()

		for rows.Next() {
			err := rows.Scan(&version)
			if err != nil {
				log.Errorf("Failed to scan: %v", err)
				return responseerror.From("Failed to query Postgres version")
			}
		}
	} else {
		output, err := cmd.Single(
			"pg-version-check",
			false,
			false,
			"sudo",
			"-u", pgInit.GetPostgresOsUser(),
			pgInit.PostgresPath+"/bin/psql",
			"-t",
			"-w",
			"-P", "format=unaligned",
			"-w",
			"-c", versionQuery,
		)
		version = util.TrimmedString(output)

		if err != nil || version == cmd.EmptyOutput {
			log.Errorf("Failed to query Postgres version, output: %v error: %v", output, err)
			return responseerror.From("Can't connect to PostgreSQL. Is it running and the configuration is correct?")
		}
	}

	if !strings.Contains(version, util.StringVal(pgInit.Version)) {
		log.Error("Postgres version mismatch")
		return responseerror.From("Postgres version mismatch")
	}

	return nil
}

func checkPgSuperuser(pgInit *repo.PgInitDto) error {
	superuserQuery := dao.PgSuperUserCheckQuery

	var queryResult string

	if pgInit.IsHostConnection() {
		_, rows, cleanup, err := dao.RunQuery(pgInit, superuserQuery)
		if err != nil {
			return err
		}
		defer cleanup()

		for rows.Next() {
			err := rows.Scan(&queryResult)
			if err != nil {
				log.Errorf("Failed to scan: %v", err)
				return responseerror.From("Failed to query Postgres Superuser permission")
			}
		}
	} else {
		output, err := cmd.Single(
			"pg-superuser-check",
			false,
			true,
			"sudo",
			"-u", pgInit.GetPostgresOsUser(),
			pgInit.PostgresPath+"/bin/psql",
			"-t",
			"-w",
			"-P", "format=unaligned",
			"-w",
			"-c", superuserQuery,
		)

		queryResult = util.TrimmedString(output)

		if err != nil {
			log.Errorf("Failed to query Postgres version, output: %v error: %v", output, err)
			return responseerror.From("Failed to query Postgres Superuser permission")
		}
	}

	if queryResult == cmd.EmptyOutput {
		errMsg := "Can't connect to PostgreSQL. Is it running and the configuration is correct?"

		log.Error(errMsg)
		return responseerror.From(errMsg)
	}

	if !strings.Contains(queryResult, "t") {
		errMsg := fmt.Sprintf(
			"%s is not a superuser. Please connect using a superuser credentials.",
			pgInit.GetPgUser(),
		)

		log.Error(errMsg)
		return responseerror.From(errMsg)
	}

	return nil
}

func checkPgReplication(pgInit *repo.PgInitDto) error {
	var queryResult string

	if pgInit.IsHostConnection() {
		_, rows, cleanup, err := dao.RunQuery(pgInit, fmt.Sprintf(dao.PgHostReplicationCheckQuery, pgInit.GetDbUsername()))
		if err != nil {
			return err
		}
		defer cleanup()

		for rows.Next() {
			err := rows.Scan(&queryResult)
			if err != nil {
				log.Errorf("Failed to scan: %v", err)
				return responseerror.From("Failed to query Postgres replication permission")
			}
		}
	} else {
		output, err := cmd.Single(
			"pg-replication-check",
			false,
			true,
			"sudo",
			"-u", pgInit.GetPostgresOsUser(),
			pgInit.PostgresPath+"/bin/psql",
			"-t",
			"-w",
			"-P", "format=unaligned",
			"-w",
			"-c", fmt.Sprintf(dao.PgLocalReplicationCheckQuery, pgInit.GetPostgresOsUser()),
		)

		queryResult = util.TrimmedString(output)

		if err != nil {
			log.Errorf("Failed to query Postgres version, output: %v error: %v", output, err)
			return responseerror.From("Can't connect to PostgreSQL. Is it running and the configuration is correct?")
		}
	}

	if queryResult == cmd.EmptyOutput {
		errMsg := "Can't connect to PostgreSQL. Is it running and the configuration is correct?"

		log.Error(errMsg)
		return responseerror.From(errMsg)
	}

	if "REPLICATION_ALLOWED" != strings.TrimSpace(queryResult) {
		errMsg := fmt.Sprintf(
			"Replication is not enabled for user %s on %s connection.",
			pgInit.GetPgUser(),
			pgInit.ConnectionType,
		)

		log.Error(errMsg)
		return responseerror.From(errMsg)
	}

	return nil
}

func insertPgEntry(ctx context.Context, pgInit *repo.PgInitDto, repo *dao.Repo, pgInfo *dao.Pg) (dao.Pg, error) {
	var createdPg dao.Pg
	var err error

	if pgInfo != nil {
		log.Infof("Updating existing Postgres entry %v", pgInfo)
		pgUpdateParams := dao.UpdatePgParams{
			PgPath:         pgInit.PostgresPath,
			Version:        int64(pgInit.Version),
			StopPg:         pgInit.StopPostgres,
			PgUser:         "postgres",
			ConnectionType: "local",
			Host:           sql.NullString{String: pgInit.Host, Valid: pgInit.IsHostConnection()},
			Port:           sql.NullInt64{Int64: int64(pgInit.Port), Valid: pgInit.IsHostConnection()},
			SslMode:        sql.NullString{String: pgInit.SslMode, Valid: pgInit.IsHostConnection()},
			Username:       sql.NullString{String: pgInit.DbUsername, Valid: pgInit.IsHostConnection()},
			Password:       sql.NullString{String: pgInit.Password, Valid: pgInit.IsHostConnection()},
			Status:         dao.PgStarted,
			ID:             pgInfo.ID,
		}

		createdPg, err = data.Fetcher.UpdatePg(ctx, pgUpdateParams)
	} else {
		log.Infof("Creating new Postgres entry")

		pgParams := dao.CreatePgParams{
			PgPath:         pgInit.PostgresPath,
			Version:        int64(pgInit.Version),
			StopPg:         pgInit.StopPostgres,
			PgUser:         pgInit.PostgresOsUser,
			ConnectionType: pgInit.ConnectionType,
			Host:           sql.NullString{String: pgInit.Host, Valid: pgInit.IsHostConnection()},
			Port:           sql.NullInt64{Int64: int64(pgInit.Port), Valid: pgInit.IsHostConnection()},
			SslMode:        sql.NullString{String: pgInit.SslMode, Valid: pgInit.IsHostConnection()},
			Username:       sql.NullString{String: pgInit.DbUsername, Valid: pgInit.IsHostConnection()},
			Password:       sql.NullString{String: pgInit.Password, Valid: pgInit.IsHostConnection()},
			Status:         dao.PgStarted,
			RepoID:         repo.ID,
		}

		createdPg, err = data.Fetcher.CreatePg(ctx, pgParams)
	}

	if err != nil {
		log.Errorf("Cannot add Postgres data: %v", err)
		return dao.Pg{}, responseerror.From("Cannot save Postgres info, please check logs")
	}

	log.Infof("Created postgres entry: %v", createdPg)
	return createdPg, nil
}

func copyPostgresData(
	pgInit *repo.PgInitDto,
	repo *dao.Repo,
	pool *dao.ZfsPool,
	dataset *dao.ZfsDataset,
	pgInstance *dao.Pg,
) {

	log.Info("Started copying Postgres data to ZFS Dataset")
	log.Infof("Repo: %v", repo)
	log.Infof("Pool: %v", pool)
	log.Infof("Dataset: %v", dataset)
	log.Infof("Pg: %v", pgInstance)

	ctx := context.Background()
	pgBaseBackupPath := pgInit.PostgresPath + "/bin/pg_basebackup"
	mainDatasetPath := pool.MountPath + "/main/data"

	if err := os.RemoveAll(mainDatasetPath); err != nil {
		log.Errorf("Failed to cleanup main dataset directory: %v", err)
		return
	}

	if err := zfs.CreateDirectories(mainDatasetPath, 0700); err != nil {
		log.Errorf("Failed to create main dataset directory: %v", err)
		return
	}

	osUser, err := user.Lookup(pgInit.PostgresOsUser)
	if err != nil {
		log.Errorf("Failed to lookup postgres user: %s, error: %v", pgInit.PostgresOsUser, err)
		return
	}

	uid, _ := strconv.Atoi(osUser.Uid)
	gid, _ := strconv.Atoi(osUser.Gid)

	if err := os.Chown(mainDatasetPath, uid, gid); err != nil {
		log.Errorf("Failed to change ownership of main dataset directory: %v", err)
		return
	}

	var output *string
	var cmderr error

	if err := pg.CreatePgPassFile(pgInit); err != nil {
		return
	}

	if pgInit.IsHostConnection() {
		output, cmderr = cmd.Single(
			"pg-base-backup-host",
			false,
			false,
			pgBaseBackupPath,
			"-w",
			"-U", pgInit.GetDbUsername(),
			"-h", pgInit.GetHost(),
			"-p", fmt.Sprintf("%d", pgInit.GetPort()),
			"-D", mainDatasetPath,
		)
	} else {
		output, cmderr = cmd.Single(
			"pg-base-backup-local",
			false,
			false,
			"sudo",
			"-u", pgInit.GetPostgresOsUser(),
			pgBaseBackupPath,
			"-w",
			"-D", mainDatasetPath,
		)
	}

	_ = pg.RemovePgPassFile()

	if cmderr != nil {
		log.Errorf("Failed to copy pg instance. output: %s data: %v", util.SafeStringVal(output), cmderr)

		updatePgParams := dao.UpdatePgStatusParams{
			Status: dao.PgFailed,
			Output: sql.NullString{String: util.SafeStringVal(output), Valid: true},
			ID:     pgInstance.ID,
		}

		updatedPg, err := data.Fetcher.UpdatePgStatus(ctx, updatePgParams)
		if err != nil {
			log.Errorf("Failed to update import status of pgInstance: %v", err)
		}

		log.Infof("Updated import status of pgInstance: %v", updatedPg)

		return
	}

	updatePgParams := dao.UpdatePgStatusParams{
		Status: dao.PgCompleted,
		Output: sql.NullString{String: util.SafeStringVal(output), Valid: true},
		ID:     pgInstance.ID,
	}

	updatedPg, err := data.Fetcher.UpdatePgStatus(ctx, updatePgParams)
	if err != nil {
		log.Errorf("Failed to update import status of pgInstance: %v", err)
	}
	log.Infof("Updated pgInstance: %v", updatedPg)

	branchParams := dao.CreateBranchParams{
		RepoID:    repo.ID,
		Name:      "main",
		DatasetID: dataset.ID,
	}

	_, err = data.Fetcher.CreateBranch(ctx, branchParams)
	if err != nil {
		log.Errorf("Failed to create main branch: %v", err)
		return
	}

	log.Infof("Postgres backup successful for repo: %v", repo)
}
