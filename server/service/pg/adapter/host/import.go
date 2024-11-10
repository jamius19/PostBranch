package host

import (
	"context"
	"errors"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/dto/pg"
	"github.com/jamius19/postbranch/dto/repo"
	"github.com/jamius19/postbranch/logger"
	pgSvc "github.com/jamius19/postbranch/service/pg"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	_ "github.com/lib/pq"
	"os"
	"strconv"
	"strings"
)

const errMsg = "Can't connect to PostgreSQL. Is it running and is the provided configuration correct?"

var log = logger.Logger

func Validate(pgInit pg.HostImportReqDto) error {
	pgBaseBackupPath := pgInit.PostgresPath + "/bin/pg_basebackup"
	if _, err := os.Stat(pgBaseBackupPath); errors.Is(err, os.ErrNotExist) {
		return responseerror.From("Invalid Postgres path, please check the path")
	}

	err := checkPgVersion(pgInit)
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

func Import(
	ctx context.Context,
	repoInit repo.InitDto[pg.HostImportReqDto],
	repoInfo model.Repo,
	pool model.ZfsPool,
	pgInfo *model.Pg,
) (model.Pg, error) {

	pgConfig := repoInit.PgConfig

	// Get the main dataset for importing Postgres data
	dataset, err := db.GetDatasetByName(ctx, pool.Name+"/main")
	if err != nil {
		log.Errorf("Dataset not found for repo: %v and pool: %v", repo.MinSizeInMb, pool)
		return model.Pg{}, responseerror.From("Associated Dataset not found")
	}

	createdPg, err := insertPgEntry(ctx, pgConfig, repoInfo, pgInfo)
	if err != nil {
		return model.Pg{}, err
	}

	go copyPostgresData(pgConfig, repoInfo, pool, dataset, &createdPg)

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
	output, err := pgSvc.Single(&pgInit, pgSvc.VersionQuery)
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
	var replicationQuery = fmt.Sprintf(pgSvc.HostReplicationCheckQuery, pgInit.DbUsername)

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

		pgUpdateParams := model.Pg{
			PgPath:  pgInit.PostgresPath,
			Version: pgInit.Version,
			Status:  pgSvc.Started,
			ID:      pgInfo.ID,
		}

		createdPg, err = db.UpdatePg(ctx, pgUpdateParams)
	} else {
		log.Infof("Creating new Postgres entry")
		pgParams := model.Pg{
			PgPath:  pgInit.PostgresPath,
			Version: pgInit.Version,
			Status:  pgSvc.Started,
			RepoID:  *repo.ID,
		}

		createdPg, err = db.CreatePg(ctx, pgParams)
	}

	if err != nil {
		log.Errorf("Cannot add Postgres data: %v", err)
		return model.Pg{}, responseerror.From("Cannot save Postgres info, please check logs")
	}

	log.Infof("Created postgres entry: %v", createdPg)
	return createdPg, nil
}

//func getConfFilePath(pgInit *pg.LocalImportReqDto) (string, error) {
//	output, err := pgSvc.Single(pgInit, ConfigFilePathQuery)
//	if err != nil {
//		log.Errorf("Failed to query Postgres config file, output: %v error: %v", output, err)
//		return "", fmt.Errorf("failed to query Postgres config file. error: %v", err)
//	}
//
//	return output, nil
//}
//
//func getHbaFilePath(pgInit *pg.LocalImportReqDto) (string, error) {
//	output, err := pgSvc.Single(pgInit, HbaFilePathQuery)
//	if err != nil {
//		log.Errorf("Failed to query Postgres hba file, output: %v error: %v", output, err)
//		return "", fmt.Errorf("failed to query Postgres hba file. error: %v", err)
//	}
//
//	return output, nil
//}
//
//func getIdentFilePath(pgInit *pg.LocalImportReqDto) (string, error) {
//	output, err := pgSvc.Single(pgInit, IdentFilePathQuery)
//	if err != nil {
//		log.Errorf("Failed to query Postgres ident file, output: %v error: %v", output, err)
//		return "", fmt.Errorf("failed to query Postgres ident file. error: %v", err)
//	}
//
//	return output, nil
//}

func copyPostgresData(
	pgInit pg.HostImportReqDto,
	repo model.Repo,
	pool model.ZfsPool,
	dataset model.ZfsDataset,
	pgInstance *model.Pg,
) {

	log.Info("Started copying Local Postgres data to ZFS Dataset")
	log.Infof("Repo: %v", repo)
	log.Infof("Pool: %v", pool)
	log.Infof("Dataset: %v", dataset)
	log.Infof("Pg: %v", pgInstance)

	ctx := context.Background()
	pgBaseBackupPath := pgInit.PostgresPath + "/bin/pg_basebackup"
	mainDatasetPath := pool.MountPath + "/main/data"

	// Cleaning the new dataset directory
	if err := util.RemoveFile(mainDatasetPath); err != nil {
		log.Errorf("Failed to cleanup main dataset directory: %v", err)
		return
	}

	// TODO: Change to postbranch user
	if err := util.CreateDirectories(mainDatasetPath, "postgres", 0700); err != nil {
		log.Errorf("Failed to create main dataset directory: %v", err)
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
		log.Errorf("Failed to copy pg instance. output: %s data: %v", util.SafeStringVal(output), err)

		updatedPg, err := db.UpdatePgStatus(ctx, *pgInstance.ID, pgSvc.Failed, util.SafeStringVal(output))
		if err != nil {
			log.Errorf("Failed to update import status of pgInstance: %v", err)
		}

		log.Infof("Updated import status of pgInstance: %v", updatedPg)

		return
	}

	err = cleanupConfig(mainDatasetPath)
	if err != nil {
		return
	}

	// Updating DB
	updatedPg, err := db.UpdatePgStatus(ctx, *pgInstance.ID, pgSvc.Completed, util.SafeStringVal(output))
	if err != nil {
		log.Errorf("Failed to update import status of pgInstance: %v", err)
	}
	log.Infof("Updated pgInstance: %v", updatedPg)

	branch := model.Branch{
		Name:      "main",
		RepoID:    *repo.ID,
		DatasetID: *dataset.ID,
	}

	branch, err = db.CreateBranch(ctx, branch)
	if err != nil {
		log.Errorf("Failed to create main branch: %v", err)
		return
	}

	log.Infof("Postgres backup successful for repo: %v", repo)
}

func cleanupConfig(mainDatasetPath string) error {
	if err := util.RemoveFile(mainDatasetPath + "/postgresql.conf"); err != nil {
		log.Errorf("Failed to remove postgresql.conf")
		return err
	}

	if err := util.RemoveFile(mainDatasetPath + "/pg_hba.conf"); err != nil {
		log.Errorf("Failed to remove pg_hba.conf")
		return err
	}

	if err := util.RemoveFile(mainDatasetPath + "/pg_ident.conf"); err != nil {
		log.Errorf("Failed to remove pg_ident.conf")
		return err
	}

	return nil
}
