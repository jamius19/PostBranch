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

const errMsg = "Can't connect to PostgreSQL. Is it running and is the provided configuration correct?"

func Validate(pgInit *repo.PgInitDto) error {
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

	return nil
}

func Import(ctx context.Context, repoinit *repo.InitDto, repo *dao.Repo, pool *dao.ZfsPool, pgInfo *dao.Pg) (*dao.Pg, error) {
	pgInit := &repoinit.PgInitDto

	if err := Validate(pgInit); err != nil {
		return nil, err
	}

	if repoinit.SizeInMb < max(256, repoinit.ClusterSizeInMb) {
		log.Errorf("Client requested size of %d MB is too small. Cluster size should be at least %d MB",
			repoinit.SizeInMb, max(256, repoinit.ClusterSizeInMb))

		return nil, responseerror.From(
			fmt.Sprintf("Cluster size should be at least %d MB", max(256, repoinit.ClusterSizeInMb)),
		)
	}

	// Get the main dataset for importing Postgres data
	dataset, err := data.Db.GetDatasetByName(ctx, pool.Name+"/main")
	if err != nil {
		log.Errorf("Dataset not found for repo: %v and pool: %v", repo, pool)
		return nil, responseerror.From("Associated Dataset not found")
	}

	createdPg, err := insertPgEntry(ctx, pgInit, repo, pgInfo)
	if err != nil {
		return nil, err
	}

	go copyPostgresData(pgInit, repo, pool, &dataset, &createdPg)

	return &createdPg, nil
}

func GetClusterSize(pgInit *repo.PgInitDto) (int64, error) {
	var sizeInMb int64

	output, err := Single(pgInit, ClusterSizeQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres Cluster size: %v", err)
		return 0, responseerror.From(errMsg)
	}

	sizeInMb, err = strconv.ParseInt(output, 10, 64)
	if err != nil {
		log.Errorf("Failed to convert size to int: %v", err)
		return -1, responseerror.From(errMsg)
	}

	return sizeInMb, nil
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
	output, err := Single(pgInit, VersionQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres version: %v", err)
		return responseerror.From(errMsg)
	}

	if !strings.Contains(output, util.StringVal(pgInit.Version)) {
		log.Error("Postgres version mismatch")
		return responseerror.From("Postgres version mismatch")
	}

	return nil
}

func checkPgSuperuser(pgInit *repo.PgInitDto) error {
	output, err := Single(pgInit, SuperUserCheckQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres superuser: %v", err)
		return responseerror.From(errMsg)
	}

	if !strings.Contains(output, "t") {
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
	var replicationQuery string

	if pgInit.IsHostConnection() {
		replicationQuery = fmt.Sprintf(HostReplicationCheckQuery, pgInit.GetDbUsername())
	} else {
		replicationQuery = fmt.Sprintf(LocalReplicationCheckQuery, pgInit.GetPostgresOsUser())
	}

	output, err := Single(pgInit, replicationQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres replication: %v", err)
		return responseerror.From(errMsg)
	}

	if "REPLICATION_NOT_ALLOWED" == output {
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
			PgPath:  pgInit.PostgresPath,
			Version: int64(pgInit.Version),
			Status:  Started,
			ID:      pgInfo.ID,
		}

		createdPg, err = data.Db.UpdatePg(ctx, pgUpdateParams)
	} else {
		log.Infof("Creating new Postgres entry")
		pgParams := dao.CreatePgParams{
			PgPath:  pgInit.PostgresPath,
			Version: int64(pgInit.Version),
			Status:  Started,
			RepoID:  repo.ID,
		}

		createdPg, err = data.Db.CreatePg(ctx, pgParams)
	}

	if err != nil {
		log.Errorf("Cannot add Postgres data: %v", err)
		return dao.Pg{}, responseerror.From("Cannot save Postgres info, please check logs")
	}

	log.Infof("Created postgres entry: %v", createdPg)
	return createdPg, nil
}

func getConfFiles(pgInit *repo.PgInitDto) ([]string, error) {
	output, err := Single(pgInit, ConfigFilePathsQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres config files, output: %v error: %v", output, err)
		return nil, fmt.Errorf("failed to query Postgres config files. error: %v", err)
	}

	return strings.Split(output, ";"), nil
}

func getHbaFiles(pgInit *repo.PgInitDto) ([]string, error) {
	output, err := Single(pgInit, HbaFilePathsQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres hba files, output: %v error: %v", output, err)
		return nil, fmt.Errorf("failed to query Postgres hba files. error: %v", err)
	}

	return strings.Split(output, ";"), nil
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

	// Cleaning the new dataset directory
	if err := os.RemoveAll(mainDatasetPath); err != nil {
		log.Errorf("Failed to cleanup main dataset directory: %v", err)
		return
	}

	if err := zfs.CreateDirectories(mainDatasetPath, 0700); err != nil {
		log.Errorf("Failed to create main dataset directory: %v", err)
		return
	}

	if err := zfs.SetPermissions(mainDatasetPath, pgInit.PostgresOsUser); err != nil {
		log.Errorf("Failed to change ownership of main dataset directory: %v", err)
		return
	}

	// Backing up postgres
	var output *string
	var cmderr error

	if err := CreatePgPassFile(pgInit); err != nil {
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

	_ = RemovePgPassFile()

	if cmderr != nil {
		log.Errorf("Failed to copy pg instance. output: %s data: %v", util.SafeStringVal(output), cmderr)

		updatePgParams := dao.UpdatePgStatusParams{
			Status: Failed,
			Output: sql.NullString{String: util.SafeStringVal(output), Valid: true},
			ID:     pgInstance.ID,
		}

		updatedPg, err := data.Db.UpdatePgStatus(ctx, updatePgParams)
		if err != nil {
			log.Errorf("Failed to update import status of pgInstance: %v", err)
		}

		log.Infof("Updated import status of pgInstance: %v", updatedPg)

		return
	}

	// Updating DB
	updatePgParams := dao.UpdatePgStatusParams{
		Status: Completed,
		Output: sql.NullString{String: util.SafeStringVal(output), Valid: true},
		ID:     pgInstance.ID,
	}

	updatedPg, err := data.Db.UpdatePgStatus(ctx, updatePgParams)
	if err != nil {
		log.Errorf("Failed to update import status of pgInstance: %v", err)
	}
	log.Infof("Updated pgInstance: %v", updatedPg)

	branchParams := dao.CreateBranchParams{
		RepoID:    repo.ID,
		Name:      "main",
		DatasetID: dataset.ID,
	}

	_, err = data.Db.CreateBranch(ctx, branchParams)
	if err != nil {
		log.Errorf("Failed to create main branch: %v", err)
		return
	}

	log.Infof("Postgres backup successful for repo: %v", repo)
}
