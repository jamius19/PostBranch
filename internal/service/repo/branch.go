package repo

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/internal/db"
	"github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/dto/repo"
	"github.com/jamius19/postbranch/internal/runner"
	"github.com/jamius19/postbranch/internal/service/pg"
	"github.com/jamius19/postbranch/internal/util"
	"path/filepath"
)

func CreateBranch(ctx context.Context, repoDetail db.RepoDetail, branchInit repo.BranchInit) (model.Branch, error) {
	parentBranch, err := db.GetBranch(ctx, branchInit.ParentId)
	if err != nil {
		log.Error("Can't get parent branch: %s", err)
		return model.Branch{}, err
	}

	// TODO: Add a checkpoint to parent branch

	snapshotName := fmt.Sprintf("%s/%s@pb-branch-%s", repoDetail.Pool.Name, parentBranch.Name, branchInit.Name)
	_, err = runner.Single(
		"create-zfs-branch-snapshot",
		false,
		false,
		"zfs",
		"snapshot",
		snapshotName,
	)

	if err != nil {
		log.Errorf("Can't create branch snapshot: %s", err)
		return model.Branch{}, err
	}

	log.Infof("Created branch snapshot %s", snapshotName)

	_, err = runner.Single(
		"clone-zfs-branch",
		false,
		false,
		"zfs",
		"clone",
		snapshotName,
		fmt.Sprintf("%s/%s", repoDetail.Pool.Name, branchInit.Name),
	)
	if err != nil {
		log.Errorf("Can't clone branch: %s", err)
		return model.Branch{}, err
	}

	port, err := pg.GetPgPort(ctx)
	if err != nil {
		log.Errorf("Can't get pg port: %s", err)
		return model.Branch{}, err
	}

	branch := model.Branch{
		Name:     branchInit.Name,
		Status:   string(db.BranchOpen),
		PgStatus: string(db.BranchPgStarting),
		PgPort:   port,
		RepoID:   *repoDetail.Repo.ID,
		ParentID: parentBranch.ID,
	}

	branch, err = db.CreateBranch(ctx, branch)
	if err != nil {
		log.Errorf("Can't create branch: %s", err)
		return model.Branch{}, err
	}

	go startBranchPg(repoDetail, branch)

	log.Infof("Created new branch %s", branchInit.Name)
	return model.Branch{}, nil
}

func CloseBranch(ctx context.Context, repoDetail db.RepoDetail, branchClose repo.BranchClose) error {
	branch, err := db.GetBranchByName(ctx, branchClose.Name)
	if err != nil {
		log.Error("Can't get branch: %s", err)
		return err
	}

	err = pg.StopPg(repoDetail.Repo.PgPath, repoDetail.Pool.MountPath, branch.Name, false)
	if err != nil {
		return err
	}

	branchDatasetName := fmt.Sprintf("%s/%s", repoDetail.Pool.Name, branch.Name)

	_, err = runner.Single(
		"delete-zfs-dataset",
		false,
		false,
		"zfs",
		"destroy",
		"-r",
		branchDatasetName,
	)

	if err != nil {
		log.Errorf("Can't close branch: %s", err)
		return err
	}

	err = db.UpdateBranchStatus(ctx, *branch.ID, db.BranchClosed)
	if err != nil {
		return err
	}

	return nil
}

func startBranchPg(repoDetail db.RepoDetail, branch model.Branch) {
	datasetPath := filepath.Join(repoDetail.Pool.MountPath, branch.Name, "data")
	logPath := filepath.Join(repoDetail.Pool.MountPath, branch.Name, "logs")

	err := pg.UpdatePostgresConfig(datasetPath, "port", util.StringVal(branch.PgPort))
	if err != nil {
		log.Errorf("Can't update postgres port, branch: %s, err: %s", branch.Name, err)
		return
	}

	err = pg.UpdatePostgresConfig(
		datasetPath,
		"log_filename",
		fmt.Sprintf("'%s_%s__%s.log'", repoDetail.Repo.Name, branch.Name, "%Y-%m-%d_%H-%M-%S"),
	)

	if err != nil {
		log.Errorf("Can't update postgres log file pattern, branch: %s, err: %s", branch.Name, err)
		return
	}

	err = pg.UpdatePostgresConfig(
		datasetPath,
		"log_directory",
		fmt.Sprintf("'%s'", logPath),
	)

	if err != nil {
		log.Errorf("Can't update postgres log dir, branch: %s, err: %s", branch.Name, err)
		return
	}

	logDirGlobPattern := filepath.Join(repoDetail.Pool.MountPath, branch.Name, "logs", "*")
	err = util.RemoveGlob(logDirGlobPattern)
	if err != nil {
		log.Errorf("Can't remove log files, branch: %s, err: %s", branch.Name, err)
		return
	}

	err = pg.CleanPidFile(datasetPath)
	if err != nil {
		log.Errorf("Can't clean pid file, branch: %s, err: %s", branch.Name, err)
		return
	}

	status, err := pg.StartPg(repoDetail.Repo.PgPath, repoDetail.Pool.MountPath, branch.Name)
	if err != nil {
		log.Errorf("Can't start Postgres: %s", err)
		return
	}

	err = db.UpdateBranchPgStatus(context.Background(), *branch.ID, status)
	if err != nil {
		log.Errorf("Can't update branch status: %s", err)
		return
	}

	log.Infof("Started Postgres on branch %s", branch.Name)
}
