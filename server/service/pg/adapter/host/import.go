package host

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/dto/pg"
	"github.com/jamius19/postbranch/logger"
	pgSvc "github.com/jamius19/postbranch/service/pg"
	"github.com/jamius19/postbranch/service/zfs"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"path/filepath"
	"strconv"
	"strings"
)

const errMsg = "Can't connect to PostgreSQL. Is it running and is the provided configuration correct?"

var log = logger.Logger

func Validate(pgInit pg.HostImportReqDto) error {
	if err := pgSvc.ValidatePgPath(pgInit.PostgresPath); err != nil {
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
	repoInfo model.Repo,
	pool model.ZfsPool,
	pgInfo *model.Pg,
) (model.Pg, error) {

	createdPg, err := insertPgEntry(ctx, pgConfig, repoInfo, pgInfo)
	if err != nil {
		return model.Pg{}, err
	}

	go copyPostgresData(pgConfig, repoInfo, pool, createdPg)

	return createdPg, nil
}

func GetClusterSize(pgInit pg.HostImportReqDto) (int64, error) {
	var sizeInMb int64

	output, err := pgSvc.Single(&pgInit, pgSvc.ClusterSizeQuery)
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
	output, err := cmd.Single(
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

	if !strings.Contains(output, util.StringVal(pgInit.Version)) {
		log.Error("Postgres version mismatch")
		return responseerror.From("Postgres installation version mismatch")
	}

	output, err = pgSvc.Single(&pgInit, pgSvc.VersionQuery)
	if err != nil {
		log.Errorf("Failed to query Postgres version: %v", err)
		return responseerror.From(errMsg)
	}

	if !strings.Contains(output, util.StringVal(pgInit.Version)) {
		log.Error("Postgres version mismatch")
		return responseerror.From("Database cluster postgres version mismatch")
	}

	return nil
}

func checkPgSuperuser(pgInit pg.HostImportReqDto) error {
	output, err := pgSvc.Single(&pgInit, pgSvc.SuperUserCheckQuery)
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
	var replicationQuery = fmt.Sprintf(pgSvc.ReplicationCheckQuery, pgInit.DbUsername)

	output, err := pgSvc.Single(&pgInit, replicationQuery)
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
	repo model.Repo,
	pgInfo *model.Pg,
) (model.Pg, error) {

	var createdPg model.Pg
	var err error

	if pgInfo != nil {
		log.Infof("Updating existing Postgres entry %v", pgInfo)

		pgUpdate := model.Pg{
			PgPath:  pgInit.PostgresPath,
			Version: pgInit.Version,
			Adapter: "host",
			Status:  string(db.PgStarted),
			ID:      pgInfo.ID,
		}

		createdPg, err = db.UpdatePg(ctx, pgUpdate)
	} else {
		log.Infof("Creating new Postgres entry")

		pgCreate := model.Pg{
			PgPath:  pgInit.PostgresPath,
			Version: pgInit.Version,
			Adapter: "host",
			Status:  string(db.PgStarted),
			RepoID:  *repo.ID,
		}

		createdPg, err = db.CreatePg(ctx, pgCreate)
	}

	if err != nil {
		log.Errorf("Cannot add Postgres data: %v", err)
		return model.Pg{}, responseerror.From("Cannot save Postgres info, please check logs")
	}

	log.Infof("Created postgres entry: %v", createdPg)
	return createdPg, nil
}

func copyPostgresData(
	pgInit pg.HostImportReqDto,
	repo model.Repo,
	pool model.ZfsPool,
	pgInfo model.Pg,
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

	port, err := pgSvc.GetPgPort(ctx)
	if err != nil {
		return
	}

	if err := util.CreateDirectories(mainDatasetPath, pgSvc.PostBranchUser, 0700); err != nil {
		log.Errorf("Failed to create main dataset directory: %v", err)
		return
	}

	if err := util.CreateDirectories(logPath, pgSvc.PostBranchUser, 0700); err != nil {
		log.Errorf("Failed to create log directory: %v", err)
		return
	}

	if err := pgSvc.CreatePgPassFile(&pgInit); err != nil {
		log.Error(err)
		return
	}

	// Backing up postgres
	output, err := cmd.Single(
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

	_ = pgSvc.RemovePgPassFile()

	if err != nil {
		log.Errorf("Failed to copy pg instance. output: %s data: %v", output, err)

		updatedPg, err := db.UpdatePgStatus(ctx, *pgInfo.ID, db.PgFailed, output)
		if err != nil {
			log.Errorf("Failed to update import status of pgInfo: %v", err)
		}

		log.Infof("Updated import status of pgInfo: %v", updatedPg)

		return
	}

	if err := pgSvc.CleanupConfig(mainDatasetPath); err != nil {
		return
	}

	if err := pgSvc.WritePostgresConfig(port, repo.Name, branchName, logPath, mainDatasetPath); err != nil {
		return
	}

	hbaConfigs, err := getHbaFileConfig(&pgInit)
	if err != nil {
		return
	}

	if err := pgSvc.WritePgHbaConfig(hbaConfigs, mainDatasetPath); err != nil {
		return
	}

	// Set the permissions for the main dataset directory to PostBranch user
	// as after the backup, the permissions are set to root
	err = util.SetPermissionsRecursive(mainDatasetPath, pgSvc.PostBranchUser, pgSvc.PostBranchUser)
	if err != nil {
		log.Errorf("Failed to change dataset permissions. output: %s data: %v", output, err)
		return
	}

	// Updating DB
	updatedPg, err := db.UpdatePgStatus(ctx, *pgInfo.ID, db.PgCompleted, output)
	if err != nil {
		log.Errorf("Failed to update status of pgInfo: %v", err)
	}

	log.Infof("Postgres backup successful for repo: %v", repo)

	log.Infof("Updated pg info, pg: %v", updatedPg)

	branch := model.Branch{
		Name:     "main",
		PgPort:   port,
		RepoID:   *repo.ID,
		PgStatus: string(db.BranchPgStarting),
		Status:   string(db.BranchOpen),
	}

	branch, err = db.CreateBranch(ctx, branch)
	if err != nil {
		log.Errorf("Failed to create main branch: %v", err)
		return
	}

	status, err := pgSvc.StartPg(pgInit.PostgresPath, pool.MountPath, branchName)
	if err != nil {
		return
	}

	err = db.UpdateBranchPgStatus(ctx, *branch.ID, status)
	if err != nil {
		log.Errorf("Failed to update branch status: %v", err)
		return
	}
}
