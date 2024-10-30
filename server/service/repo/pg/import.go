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
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"strings"
)

var log = logger.Logger

func Import(ctx context.Context, repoPgInit repo.PgInitDto, repo *dao.Repo) error {
	pgBaseBackupPath := repoPgInit.PostgresPath + "/bin/pg_basebackup"
	if _, err := os.Stat(pgBaseBackupPath); errors.Is(err, fs.ErrNotExist) {
		return responseerror.Clarify("Invalid Postgres path, pg_basebackup not found")
	}

	//TODO: Copy Postgres data to Repo's ZFS Dataset
	pgVersion, err := cmd.Single(
		"pg-version-check",
		false,
		"sudo",
		"sudo",
		"-u",
		"postgres",
		repoPgInit.PostgresPath+"/bin/psql",
		"-t", "-P", "format=unaligned",
		"-c", "SELECT split_part(current_setting('server_version'), '.', 1) AS major_version;",
	)

	if err != nil || pgVersion == nil {
		return responseerror.Clarify("Can't connect to PostgreSQL. Is it running and accessible via postgres user?")
	}

	if !strings.Contains(*pgVersion, util.StringVal(repoPgInit.Version)) {
		return responseerror.Clarify("Postgres version mismatch")
	}

	//TODO: Copy Postgres data to Repo's ZFS Dataset
	pool, err := data.Fetcher.GetPool(ctx, repo.PoolID)
	if err != nil {
		logrus.Errorf("Pool not found for repo: %v", repo)
		return responseerror.Clarify("Associated Pool not found")
	}

	// Get the main dataset for importing Postgres data
	dataset, err := data.Fetcher.GetDatasetByName(ctx, pool.Name+"/main")
	if err != nil {
		logrus.Errorf("Dataset not found for repo: %v and pool: %v", repo, pool)
		return responseerror.Clarify("Associated Dataset not found")
	}

	pg := dao.CreatePgParams{
		PgPath:           repoPgInit.PostgresPath,
		Version:          int64(repoPgInit.Version),
		StopPg:           repoPgInit.StopPostgres,
		PgUser:           repoPgInit.PostgresUser,
		CustomConnection: repoPgInit.CustomConnection,
		Host:             sql.NullString{String: repoPgInit.Host, Valid: repoPgInit.CustomConnection},
		Port:             sql.NullInt64{Int64: int64(repoPgInit.Port), Valid: repoPgInit.CustomConnection},
		Username:         sql.NullString{String: repoPgInit.Username, Valid: repoPgInit.CustomConnection},
		Password:         sql.NullString{String: repoPgInit.Password, Valid: repoPgInit.CustomConnection},
		Status:           dao.PgStarted,
	}

	log.Infof("Creating postgres entry: %v", pg)
	createdPg, err := data.Fetcher.CreatePg(ctx, pg)
	if err != nil {
		log.Errorf("Cannot add Postgres data: %v", err)
		return responseerror.Clarify("Cannot save Postgres info, please check logs")
	}

	go copyPostgresData(&repoPgInit, repo, &pool, &dataset, &createdPg, pgBaseBackupPath)
	return nil
}

func copyPostgresData(
	repoPgInit *repo.PgInitDto,
	repo *dao.Repo,
	pool *dao.ZfsPool,
	dataset *dao.ZfsDataset,
	pg *dao.Pg,
	pgBaseBackupPath string,
) {

	log.Info("Started copying Postgres data to ZFS Dataset")
	log.Infof("RepoPgInit: %v", repoPgInit)
	log.Infof("Repo: %v", repo)
	log.Infof("Pool: %v", pool)
	log.Infof("Dataset: %v", dataset)
	log.Infof("Pg: %v", pg)

	mainDatasetPath := pool.MountPath + "/main/data"
	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()
	cmds.Set(
		"create-data-directory",
		cmd.Get("mkdir", mainDatasetPath),
	)

	cmds.Set(
		"change-data-directory-permissions",
		cmd.Get(
			"chown", "-R",
			fmt.Sprintf("%s:%s", repoPgInit.PostgresUser, repoPgInit.PostgresUser),
			mainDatasetPath,
		),
	)

	cmds.Set(
		"pg-basebackup",
		cmd.Get(
			"sudo", "-u",
			repoPgInit.PostgresUser,
			pgBaseBackupPath, "-w",
			"-D", mainDatasetPath,
		),
	)

	output, err := cmd.Multi(cmds)

	if err != nil {
		errStr := cmd.GetError(output)
		log.Errorf("Failed to copy Postgres. output: %s data: %v", errStr, err)

		updatePgParams := dao.UpdatePgParams{
			Status: dao.PgFailed,
			Output: sql.NullString{String: errStr, Valid: true},
			ID:     pg.ID,
		}

		updatedPg, err := data.Fetcher.UpdatePg(context.Background(), updatePgParams)
		if err != nil {
			log.Errorf("Failed to update Postgres import: %v", err)
		}

		log.Infof("Updated Postgres import: %v", updatedPg)

		return
	}

	updatePgParams := dao.UpdatePgParams{
		Status: dao.PgCompleted,
		Output: sql.NullString{String: cmd.GetError(output), Valid: true},
		ID:     pg.ID,
	}

	updatedPg, err := data.Fetcher.UpdatePg(context.Background(), updatePgParams)
	if err != nil {
		log.Errorf("Failed to update Postgres import: %v", err)
	}

	log.Infof("Updated pg: %v", updatedPg)
	log.Infof("Postgres backup successful for repo: %v", repo)
}
