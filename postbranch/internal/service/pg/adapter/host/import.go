package host

import (
	"context"
	"fmt"
	db2 "github.com/jamius19/postbranch/internal/db"
	model2 "github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/dto/pg"
	"github.com/jamius19/postbranch/internal/logger"
	"github.com/jamius19/postbranch/internal/runner"
	pg2 "github.com/jamius19/postbranch/internal/service/pg"
	"github.com/jamius19/postbranch/internal/service/zfs"
	util2 "github.com/jamius19/postbranch/internal/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"path/filepath"
	"strconv"
	"strings"
)

const errMsg = "Can't connect to PostgreSQL. Is it running and is the provided configuration correct?"

var log = logger.Logger

func Validate(pgInit pg.HostImportReqDto) error {
	if err := pg2.ValidatePgPath(pgInit.PostgresPath); err != nil {
		return err
	}

	if err := checkPgVersion(pgInit); err != nil {
		return err
	}

	//if err := checkPgSuperuser(pgInit); err != nil {
	//	return err
	//}
	//
	//if err := checkPgReplication(pgInit); err != nil {
	//	return err
	//}

	return nil
}

func Import(
	ctx context.Context,
	pgConfig pg.HostImportReqDto,
	repoInfo model2.Repo,
	pool model2.ZfsPool,
	pgInfo *model2.Pg,
) (model2.Pg, error) {

	createdPg, err := insertPgEntry(ctx, pgConfig, repoInfo, pgInfo)
	if err != nil {
		return model2.Pg{}, err
	}

	go copyPostgresData(pgConfig, repoInfo, pool, createdPg)

	return createdPg, nil
}

func GetClusterSize(pgInit pg.HostImportReqDto) (int64, error) {
	var sizeInMb int64

	output, err := pg2.Single(&pgInit, pg2.ClusterSizeQuery)
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

func checkPgVersion(pgInit pg.HostImportReqDto) error {
	output, err := runner.Single(
		"local-postgres-version",
		false,
		false,
		filepath.Join(pgInit.PostgresPath, "bin", "postgres"),
		"-V",
	)

	if err != nil {
		log.Errorf("Failed to query Postgres version: %v", err)
		return responseerror.From(errMsg)
	}

	if !strings.Contains(output, util2.StringVal(pgInit.Version)) {
		log.Error("Postgres version mismatch")
		return responseerror.From("Postgres installation version mismatch")
	}

	output, err = pg2.Single(&pgInit, pg2.VersionQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres version: %v", err)
		return responseerror.From(errMsg)
	}

	if !strings.Contains(output, util2.StringVal(pgInit.Version)) {
		log.Error("Postgres version mismatch")
		return responseerror.From("Database cluster postgres version mismatch")
	}

	return nil
}

func checkPgSuperuser(pgInit pg.HostImportReqDto) error {
	output, err := pg2.Single(&pgInit, pg2.SuperUserCheckQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres superuser: %v", err)
		return responseerror.From(errMsg)
	}

	if !strings.Contains(output, "t") {
		errMsg := fmt.Sprintf(
			"%s is not a superuser. Please connect using a superuser credentials.",
			pgInit.DbUsername,
		)

		log.Error(errMsg)
		return responseerror.From(errMsg)
	}

	return nil
}

func checkPgReplication(pgInit pg.HostImportReqDto) error {
	var replicationQuery = fmt.Sprintf(pg2.ReplicationCheckQuery, pgInit.DbUsername)

	output, err := pg2.Single(&pgInit, replicationQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres replication: %v", err)
		return responseerror.From(errMsg)
	}

	if "REPLICATION_NOT_ALLOWED" == output {
		errMsg := fmt.Sprintf(
			"Replication is not enabled for user %s on host connection.",
			pgInit.DbUsername,
		)

		log.Error(errMsg)
		return responseerror.From(errMsg)
	}

	return nil
}

func insertPgEntry(
	ctx context.Context,
	pgInit pg.HostImportReqDto,
	repo model2.Repo,
	pgInfo *model2.Pg,
) (model2.Pg, error) {

	var createdPg model2.Pg
	var err error

	if pgInfo != nil {
		log.Infof("Updating existing Postgres entry %v", pgInfo)

		pgUpdate := model2.Pg{
			PgPath:  pgInit.PostgresPath,
			Version: pgInit.Version,
			Adapter: "host",
			Status:  string(db2.PgStarted),
			ID:      pgInfo.ID,
		}

		createdPg, err = db2.UpdatePg(ctx, pgUpdate)
	} else {
		log.Infof("Creating new Postgres entry")

		pgCreate := model2.Pg{
			PgPath:  pgInit.PostgresPath,
			Version: pgInit.Version,
			Adapter: "host",
			Status:  string(db2.PgStarted),
			RepoID:  *repo.ID,
		}

		createdPg, err = db2.CreatePg(ctx, pgCreate)
	}

	if err != nil {
		log.Errorf("Cannot add Postgres data: %v", err)
		return model2.Pg{}, responseerror.From("Cannot save Postgres info, please check logs")
	}

	log.Infof("Created postgres entry: %v", createdPg)
	return createdPg, nil
}

func copyPostgresData(
	pgInit pg.HostImportReqDto,
	repo model2.Repo,
	pool model2.ZfsPool,
	pgInfo model2.Pg,
) {

	log.Info("Started copying host Postgres data to main branch")
	log.Infof("Repo: %v", repo)
	log.Infof("Pool: %v", pool)
	log.Infof("Pg: %v", pgInfo)

	ctx := context.Background()
	branchName := "main"
	pgBaseBackupPath := filepath.Join(pgInit.PostgresPath, "bin", "pg_basebackup")
	mainDatasetPath := filepath.Join(pool.MountPath, branchName, "data")
	logPath := filepath.Join(pool.MountPath, branchName, "logs")

	err := zfs.EmptyDataset(pool, branchName)
	if err != nil {
		return
	}

	port, err := pg2.GetPgPort(ctx)
	if err != nil {
		return
	}

	if err := util2.CreateDirectories(mainDatasetPath, pg2.PostBranchUser, 0700); err != nil {
		log.Errorf("Failed to create main dataset directory: %v", err)
		return
	}

	if err := util2.CreateDirectories(logPath, pg2.PostBranchUser, 0700); err != nil {
		log.Errorf("Failed to create log directory: %v", err)
		return
	}

	if err := pg2.CreatePgPassFile(&pgInit); err != nil {
		log.Error(err)
		return
	}

	// Backing up postgres
	output, err := runner.Single(
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

	_ = pg2.RemovePgPassFile()

	if err != nil {
		log.Errorf("Failed to copy pg instance. output: %s data: %v", output, err)

		updatedPg, err := db2.UpdatePgStatus(ctx, *pgInfo.ID, db2.PgFailed, output)
		if err != nil {
			log.Errorf("Failed to update import status of pgInfo: %v", err)
		}

		log.Infof("Updated import status of pgInfo: %v", updatedPg)

		return
	}

	if err := pg2.CleanupConfig(mainDatasetPath); err != nil {
		return
	}

	if err := pg2.WritePostgresConfig(port, repo.Name, branchName, logPath, mainDatasetPath); err != nil {
		return
	}

	hbaConfigs, err := getHbaFileConfig(&pgInit)
	if err != nil {
		return
	}

	if err := pg2.WritePgHbaConfig(hbaConfigs, mainDatasetPath); err != nil {
		return
	}

	// Set the permissions for the main dataset directory to PostBranch user
	// as after the backup, the permissions are set to root
	err = util2.SetPermissionsRecursive(mainDatasetPath, pg2.PostBranchUser, pg2.PostBranchUser)
	if err != nil {
		log.Errorf("Failed to change dataset permissions. output: %s data: %v", output, err)
		return
	}

	// Updating DB
	updatedPg, err := db2.UpdatePgStatus(ctx, *pgInfo.ID, db2.PgCompleted, output)
	if err != nil {
		log.Errorf("Failed to update status of pgInfo: %v", err)
	}

	log.Infof("Postgres backup successful for repo: %v", repo)

	log.Infof("Updated pg info, pg: %v", updatedPg)

	branch := model2.Branch{
		Name:     "main",
		PgPort:   port,
		RepoID:   *repo.ID,
		PgStatus: string(db2.BranchPgStarting),
		Status:   string(db2.BranchOpen),
	}

	branch, err = db2.CreateBranch(ctx, branch)
	if err != nil {
		log.Errorf("Failed to create main branch: %v", err)
		return
	}

	status, err := pg2.StartPg(pgInit.PostgresPath, pool.MountPath, branchName)
	if err != nil {
		return
	}

	err = db2.UpdateBranchPgStatus(ctx, *branch.ID, status)
	if err != nil {
		log.Errorf("Failed to update branch status: %v", err)
		return
	}
}
