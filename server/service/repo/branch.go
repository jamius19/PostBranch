package repo

import (
	"context"
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/dto/repo"
	"github.com/jamius19/postbranch/service/pg"
	"github.com/jamius19/postbranch/util"
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
	_, err = cmd.Single(
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

	_, err = cmd.Single(
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

	dataset := model.ZfsDataset{
		Name:   branchInit.Name,
		PoolID: *repoDetail.Pool.ID,
	}

	dataset, err = db.CreateDataset(ctx, dataset)
	if err != nil {
		log.Errorf("Can't create dataset: %s", err)
		return model.Branch{}, err
	}

	port, err := pg.GetPgPort(ctx)
	if err != nil {
		log.Errorf("Can't get pg port: %s", err)
		return model.Branch{}, err
	}

	branch := model.Branch{
		Name:      branchInit.Name,
		Status:    string(db.BranchOpen),
		PgStatus:  string(db.BranchPgStarting),
		PgPort:    port,
		RepoID:    *repoDetail.Repo.ID,
		ParentID:  parentBranch.ID,
		DatasetID: dataset.ID,
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

func startBranchPg(repoDetail db.RepoDetail, branch model.Branch) {
	dataDir := filepath.Join(repoDetail.Pool.MountPath, branch.Name, "data")
	logPath := filepath.Join(repoDetail.Pool.MountPath, branch.Name, "logs")
	logDir := filepath.Join(repoDetail.Pool.MountPath, branch.Name, "logs", "*")

	err := pg.UpdatePostgresConfig(dataDir, "port", util.StringVal(branch.PgPort))
	if err != nil {
		log.Errorf("Can't update postgres port, branch: %s, err: %s", branch.Name, err)
		return
	}

	err = pg.UpdatePostgresConfig(
		dataDir,
		"log_filename",
		fmt.Sprintf("'%s_%s__%s.log'", repoDetail.Repo.Name, branch.Name, "%Y-%m-%d_%H-%M-%S"),
	)

	if err != nil {
		log.Errorf("Can't update postgres log file pattern, branch: %s, err: %s", branch.Name, err)
		return
	}

	err = pg.UpdatePostgresConfig(
		dataDir,
		"log_directory",
		fmt.Sprintf("'%s'", logPath),
	)

	if err != nil {
		log.Errorf("Can't update postgres log dir, branch: %s, err: %s", branch.Name, err)
		return
	}

	err = util.RemoveGlob(logDir)
	if err != nil {
		log.Errorf("Can't remove log files, branch: %s, err: %s", branch.Name, err)
		return
	}

	status, err := pg.StartPg(repoDetail.Pg.PgPath, repoDetail.Pool.MountPath, branch.Name)
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
