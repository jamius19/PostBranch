package host

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/internal/db"
	"github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/dto/pg"
	"github.com/jamius19/postbranch/internal/logger"
	"github.com/jamius19/postbranch/internal/runner"
	pgSvc "github.com/jamius19/postbranch/internal/service/pg"
	"github.com/jamius19/postbranch/internal/service/zfs"
	"github.com/jamius19/postbranch/internal/util"
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

func Import(pgConfig pg.HostImportReqDto, repoInfo model.Repo, pool model.ZfsPool) {
	go copyPostgresData(pgConfig, repoInfo, pool)
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

func copyPostgresData(
	pgInit pg.HostImportReqDto,
	repo model.Repo,
	pool model.ZfsPool,
) {

	log.Info("Started copying host Postgres data to main branch")
	log.Infof("Repo: %v", repo)
	log.Infof("Pool: %v", pool)

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

	_ = pgSvc.RemovePgPassFile()

	if err != nil {
		log.Errorf("Failed to copy pg instance. output: %s data: %v", output, err)

		updatedPg, err := db.UpdateRepoStatus(ctx, *repo.ID, db.RepoFailed, output)
		if err != nil {
			log.Errorf("Failed to update import status of repo pg: %v", err)
		}

		log.Infof("Updated import status of repo pg: %v", updatedPg)

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
	updatedPg, err := db.UpdateRepoStatus(ctx, *repo.ID, db.RepoCompleted, output)
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
