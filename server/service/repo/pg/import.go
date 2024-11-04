package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/pg"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"io/fs"
	"os"
	"strings"
)

var log = logger.Logger

func Import(ctx context.Context, pgInit *repo.PgInitDto, repo *dao.Repo) (*dao.Repo, error) {
	pgBaseBackupPath := pgInit.PostgresPath + "/bin/pg_basebackup"
	if _, err := os.Stat(pgBaseBackupPath); errors.Is(err, fs.ErrNotExist) {
		return nil, responseerror.Clarify("Invalid Postgres path, please check the path")
	}

	err := checkOsUser(pgInit.PostgresOsUser)
	if err != nil {
		return nil, err
	}

	err = getPgVersion(pgInit)
	if err != nil {
		return nil, err
	}

	err = checkPgSuperuser(pgInit)
	if err != nil {
		return nil, err
	}

	err = checkPgReplication(pgInit)
	if err != nil {
		return nil, err
	}

	pool, err := data.Fetcher.GetPool(ctx, repo.PoolID)
	if err != nil {
		log.Errorf("Pool not found for repo: %v", repo)
		return nil, responseerror.Clarify("Associated Data Pool not found")
	}

	// Get the main dataset for importing Postgres data
	dataset, err := data.Fetcher.GetDatasetByName(ctx, pool.Name+"/main")
	if err != nil {
		log.Errorf("Dataset not found for repo: %v and pool: %v", repo, pool)
		return nil, responseerror.Clarify("Associated Dataset not found")
	}

	createdPg, err := createPgEntry(ctx, pgInit, err)
	if err != nil {
		return nil, err
	}

	updatedRepo, err := linkRepoWithPg(ctx, repo, createdPg, err)
	if err != nil {
		return nil, err
	}

	go copyPostgresData(pgInit, repo, &pool, &dataset, &createdPg)

	return updatedRepo, nil
}

func checkOsUser(username string) error {
	_, err := cmd.Single("os-username-check", false, false, "id", "-u", username)
	if err != nil {
		log.Errorf("User %s not found", username)
		return responseerror.Clarify("Invalid Postgres OS user")
	}

	return nil
}

func getPgVersion(pgInit *repo.PgInitDto) error {
	versionQuery := "SELECT split_part(current_setting('server_version'), '.', 1) AS major_version;"
	pgPath := pgInit.PostgresPath

	pgVersion, err := pg.Query(pgInit, "pg-version-check", false, pgPath, versionQuery)

	if err != nil || pgVersion == nil {
		errMsg := fmt.Sprintf(
			"Can't connect to PostgreSQL. Is it running and accessible via %s user on %s connection?",
			pgInit.GetPgUser(),
			pgInit.ConnectionType,
		)

		log.Error(errMsg)
		return responseerror.Clarify(errMsg)
	}

	if !strings.Contains(*pgVersion, util.StringVal(pgInit.Version)) {
		log.Error("Postgres version mismatch")
		return responseerror.Clarify("Postgres version mismatch")
	}

	return nil
}

func checkPgSuperuser(pgInit *repo.PgInitDto) error {
	pgPath := pgInit.PostgresPath
	superuserQuery := fmt.Sprintf(dao.PgSuperUserCheckQuery, pgInit.GetPgUser())

	superuserQueryOutput, err := pg.Query(
		pgInit,
		fmt.Sprintf("pg-superuser-check-%s", pgInit.ConnectionType),
		true,
		pgPath,
		superuserQuery,
	)

	if err != nil || superuserQueryOutput == nil {
		errMsg := fmt.Sprintf(
			"Can't connect to PostgreSQL. Is it running and accessible via %s user?",
			pgInit.GetPgUser(),
		)

		log.Error(errMsg)
		return responseerror.Clarify(errMsg)
	}

	if !strings.Contains(*superuserQueryOutput, "Superuser") {
		errMsg := fmt.Sprintf(
			"%s is not a superuser. Please connect using a superuser credentials.",
			pgInit.GetPgUser(),
		)

		log.Error(errMsg)
		return responseerror.Clarify(errMsg)
	}

	return nil
}

func checkPgReplication(pgInit *repo.PgInitDto) error {
	pgPath := pgInit.PostgresPath
	var replicationQuery string

	if pgInit.ConnectionType == "host" {
		replicationQuery = fmt.Sprintf(dao.PgHostReplicationCheckQuery, pgInit.GetDbUsername())
	} else {
		replicationQuery = fmt.Sprintf(dao.PgLocalReplicationCheckQuery, pgInit.GetPostgresOsUser())
	}

	replicationQueryOutput, err := pg.Query(
		pgInit,
		fmt.Sprintf("pg-replication-check-%s", pgInit.ConnectionType),
		true,
		pgPath,
		replicationQuery,
	)

	if err != nil || replicationQueryOutput == nil {
		errMsg := fmt.Sprintf(
			"Can't connec to PostgreSQL using user %s on %s connection.",
			pgInit.GetPgUser(),
			pgInit.ConnectionType,
		)

		log.Error(errMsg)
		return responseerror.Clarify(errMsg)
	}

	if "REPLICATION_ALLOWED" != strings.TrimSpace(*replicationQueryOutput) {
		errMsg := fmt.Sprintf(
			"Replication is not enabled for user %s on %s connection.",
			pgInit.GetPgUser(),
			pgInit.ConnectionType,
		)

		log.Error(errMsg)
		return responseerror.Clarify(errMsg)
	}

	return nil
}

func getPgCreateParams(pgInit *repo.PgInitDto) dao.CreatePgParams {
	return dao.CreatePgParams{
		PgPath:           pgInit.PostgresPath,
		Version:          int64(pgInit.Version),
		StopPg:           pgInit.StopPostgres,
		PgUser:           pgInit.PostgresOsUser,
		CustomConnection: pgInit.IsHostConnection(),
		Host:             sql.NullString{String: pgInit.Host, Valid: pgInit.IsHostConnection()},
		Port:             sql.NullInt64{Int64: int64(pgInit.Port), Valid: pgInit.IsHostConnection()},
		Username:         sql.NullString{String: pgInit.DbUsername, Valid: pgInit.IsHostConnection()},
		Password:         sql.NullString{String: pgInit.Password, Valid: pgInit.IsHostConnection()},
		Status:           dao.PgStarted,
	}
}

func createPgEntry(ctx context.Context, pgInit *repo.PgInitDto, err error) (dao.Pg, error) {
	pgParams := getPgCreateParams(pgInit)
	log.Infof("Creating postgres entry: %v", pgParams)

	createdPg, err := data.Fetcher.CreatePg(ctx, pgParams)
	if err != nil {
		log.Errorf("Cannot add Postgres data: %v", err)
		return dao.Pg{}, responseerror.Clarify("Cannot save Postgres info, please check logs")
	}

	log.Infof("Created postgres entry: %v", createdPg)
	return createdPg, nil
}

func linkRepoWithPg(ctx context.Context, repo *dao.Repo, createdPg dao.Pg, err error) (*dao.Repo, error) {
	updateRepoPgParams := dao.UpdateRepoPgParams{
		ID:   repo.ID,
		PgID: sql.NullInt64{Int64: createdPg.ID, Valid: true},
	}

	log.Infof("Linking repo with pg: %v", updateRepoPgParams)
	updatedRepo, err := data.Fetcher.UpdateRepoPg(ctx, updateRepoPgParams)
	if err != nil {
		log.Errorf("Cannot update repo pgParams: %v", err)
		return nil, responseerror.Clarify("Cannot save Postgres info, please check logs")
	}

	log.Infof("Linked repo with pg: %v", updatedRepo)
	return &updatedRepo, nil
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

	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()
	cmds.Set(
		"clean-data-directory",
		cmd.Get("rm", "-rf", mainDatasetPath),
	)

	cmds.Set(
		"create-data-directory",
		cmd.Get("mkdir", mainDatasetPath),
	)

	cmds.Set(
		"change-data-directory-permissions",
		cmd.Get(
			"chown", "-R",
			fmt.Sprintf("%s:%s", pgInit.PostgresOsUser, pgInit.PostgresOsUser),
			mainDatasetPath,
		),
	)

	var backupCmd cmd.Command
	if pgInit.IsHostConnection() {
		backupCmd = cmd.Get(
			pgBaseBackupPath,
			"-w",
			"-U", pgInit.GetDbUsername(),
			"-h", pgInit.GetHost(),
			"-p", fmt.Sprintf("%d", pgInit.GetPort()),
			"-D", mainDatasetPath,
		)
	} else {
		backupCmd = cmd.Get(
			pgBaseBackupPath,
			"-w",
			"-U", pgInit.GetPostgresOsUser(),
			"-D", mainDatasetPath,
		)
	}

	cmds.Set("pgInstance-basebackup", backupCmd)

	err := pg.CreatePgPassFile(pgInit)
	if err != nil {
		return
	}

	output, err := cmd.Multi(cmds)
	_ = pg.RemovePgPassFile()

	if err != nil {
		errStr := cmd.GetError(output)
		log.Errorf("Failed to copy pgInstance. output: %s data: %v", errStr, err)

		updatePgParams := dao.UpdatePgParams{
			Status: dao.PgFailed,
			Output: sql.NullString{String: errStr, Valid: true},
			ID:     pgInstance.ID,
		}

		updatedPg, err := data.Fetcher.UpdatePg(ctx, updatePgParams)
		if err != nil {
			log.Errorf("Failed to update import status of pgInstance: %v", err)
		}

		log.Infof("Updated import status of pgInstance: %v", updatedPg)

		return
	}

	updatePgParams := dao.UpdatePgParams{
		Status: dao.PgCompleted,
		Output: sql.NullString{String: cmd.GetError(output), Valid: true},
		ID:     pgInstance.ID,
	}

	updatedPg, err := data.Fetcher.UpdatePg(ctx, updatePgParams)
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
