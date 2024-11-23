package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	"github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/db/gen/table"
	"time"
)

type BranchStatus string
type BranchPgStatus string

const (
	BranchOpen   BranchStatus = "OPEN"
	BranchMerged BranchStatus = "MERGED"
	BranchClosed BranchStatus = "CLOSED"

	BranchPgStarting BranchPgStatus = "STARTING"
	BranchPgStopped  BranchPgStatus = "STOPPED"
	BranchPgRunning  BranchPgStatus = "RUNNING"
	BranchPgFailed   BranchPgStatus = "FAILED"
)

func CreateBranch(ctx context.Context, branch model.Branch) (model.Branch, error) {
	var newBranch model.Branch

	branch.CreatedAt = time.Now().UTC()
	branch.UpdatedAt = time.Now().UTC()

	stmt := table.Branch.
		INSERT(table.Branch.AllColumns).
		MODEL(branch).
		RETURNING(table.Branch.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &newBranch)
	if err != nil {
		log.Errorf("Can't create branch: %s", err)
		return model.Branch{}, err
	}

	return newBranch, nil
}

func UpdateBranchPgStatus(ctx context.Context, branchId int32, status BranchPgStatus) error {
	stmt := table.Branch.
		UPDATE(table.Branch.PgStatus, table.Branch.UpdatedAt).
		SET(table.Branch.PgStatus.SET(sqlite.String(string(status))), table.Branch.UpdatedAt.SET(sqlite.CURRENT_TIMESTAMP())).
		WHERE(table.Branch.ID.EQ(sqlite.Int(int64(branchId))))

	log.Tracef("Query: %s", stmt.DebugSql())
	_, err := stmt.ExecContext(ctx, Db)
	if err != nil {
		log.Errorf("Can't update branch pg status: %s", err)
		return err
	}

	return nil
}

func GetBranch(ctx context.Context, branchId int32) (model.Branch, error) {
	var branch model.Branch
	stmt := table.Branch.
		SELECT(table.Branch.AllColumns).
		WHERE(table.Branch.ID.EQ(sqlite.Int(int64(branchId))))

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &branch)
	if err != nil {
		log.Errorf("Can't get branch: %s", err)
		return model.Branch{}, err
	}

	return branch, nil
}

func GetBranchPorts(ctx context.Context) ([]int32, error) {
	var ports []int32

	stmt := table.Branch.SELECT(table.Branch.PgPort).
		FROM(table.Branch)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &ports)
	if err != nil {
		log.Errorf("Can't get pg ports: %s", err)
		return nil, err
	}

	return ports, nil
}
