package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	"github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/db/gen/table"
	"strings"
	"time"
)

type RepoStatus string
type RepoPgAdapter string

const (
	RepoStarted   RepoStatus = "STARTED"
	RepoCompleted RepoStatus = "READY"
	RepoFailed    RepoStatus = "FAILED"

	HostAdapter RepoPgAdapter = "HOST"
)

type RepoDetail struct {
	Repo     model.Repo
	Pool     model.ZfsPool  `alias:"pool"`
	Branches []model.Branch `alias:"branches"`
}

func ListRepo(ctx context.Context) ([]RepoDetail, error) {
	var repoDetailList []RepoDetail

	stmt := sqlite.
		SELECT(
			table.Repo.AllColumns,
			table.ZfsPool.AllColumns.As("pool"),
			table.Branch.AllColumns.As("branches"),
		).
		FROM(
			table.Repo.
				INNER_JOIN(table.ZfsPool, table.Repo.PoolID.EQ(table.ZfsPool.ID)).
				LEFT_JOIN(table.Branch, table.Branch.RepoID.EQ(table.Repo.ID)),
		).
		ORDER_BY(table.Repo.CreatedAt.DESC())

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &repoDetailList)
	if err != nil {
		log.Errorf("Can't query repos: %s", err)
		return nil, err
	}

	return repoDetailList, nil
}

func ListRepoWithStatus(ctx context.Context, status ...RepoStatus) ([]RepoDetail, error) {
	var repoDetailList []RepoDetail
	var stringExpressions []sqlite.Expression
	for _, status := range status {
		stringExpressions = append(stringExpressions, sqlite.String(string(status)))
	}

	stmt := sqlite.
		SELECT(
			table.Repo.AllColumns,
			table.ZfsPool.AllColumns.As("pool"),
			table.Branch.AllColumns.As("branches"),
		).
		FROM(
			table.Repo.
				INNER_JOIN(table.ZfsPool, table.Repo.PoolID.EQ(table.ZfsPool.ID)).
				LEFT_JOIN(table.Branch, table.Branch.RepoID.EQ(table.Repo.ID)),
		).
		ORDER_BY(table.Repo.CreatedAt.DESC()).
		WHERE(table.Repo.Status.IN(stringExpressions...))

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &repoDetailList)
	if err != nil {
		log.Errorf("Can't query repos: %s", err)
		return nil, err
	}

	return repoDetailList, nil
}

func GetRepo(ctx context.Context, repoId int64) (RepoDetail, error) {
	var repoDetail RepoDetail

	stmt := sqlite.SELECT(
		table.Repo.AllColumns,
		table.ZfsPool.AllColumns.As("pool"),
		table.Branch.AllColumns.As("branches"),
	).
		FROM(table.Repo.
			INNER_JOIN(table.ZfsPool, table.Repo.PoolID.EQ(table.ZfsPool.ID)).
			LEFT_JOIN(table.Branch, table.Branch.RepoID.EQ(table.Repo.ID))).
		WHERE(table.Repo.ID.EQ(sqlite.Int(repoId))).
		ORDER_BY(table.Branch.CreatedAt.DESC())

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &repoDetail)
	if err != nil {
		log.Errorf("Can't query repos: %s", err)
		return RepoDetail{}, err
	}

	return repoDetail, nil
}

func GetRepoByName(ctx context.Context, repoName string) (RepoDetail, error) {
	var repoDetail RepoDetail

	stmt := sqlite.SELECT(
		table.Repo.AllColumns,
		table.ZfsPool.AllColumns.As("pool"),
		table.Branch.AllColumns.As("branches"),
	).
		FROM(table.Repo.
			INNER_JOIN(table.ZfsPool, table.Repo.PoolID.EQ(table.ZfsPool.ID)).
			LEFT_JOIN(table.Branch, table.Branch.RepoID.EQ(table.Repo.ID))).
		WHERE(table.Repo.Name.EQ(sqlite.String(strings.TrimSpace(repoName)))).
		ORDER_BY(table.Branch.CreatedAt.DESC())

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &repoDetail)
	if err != nil {
		log.Errorf("Can't query repos: %s", err)
		return RepoDetail{}, err
	}

	return repoDetail, nil
}

func CreateRepo(ctx context.Context, repo model.Repo) (model.Repo, error) {
	var newRepo model.Repo

	repo.CreatedAt = time.Now().UTC()
	repo.UpdatedAt = time.Now().UTC()

	stmt := table.Repo.INSERT(table.Repo.AllColumns).
		MODEL(repo).
		RETURNING(table.Repo.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &newRepo)
	if err != nil {
		log.Errorf("Can't create repo: %s", err)
		return model.Repo{}, err
	}

	return newRepo, nil
}

func UpdateRepoPg(ctx context.Context,
	repoId int32,
	pgPath string,
	version int32,
	adapter RepoPgAdapter,
	status RepoStatus,
) (model.Repo, error) {

	var updatedRepo model.Repo

	stmt := table.Repo.
		UPDATE(table.Repo.PgPath, table.Repo.Version, table.Repo.Adapter, table.Repo.Status, table.Repo.UpdatedAt).
		SET(
			sqlite.String(pgPath),
			sqlite.Int(int64(version)),
			sqlite.String(string(adapter)),
			sqlite.String(string(status)),
			sqlite.CURRENT_TIMESTAMP(),
		).
		WHERE(table.Repo.ID.EQ(sqlite.Int(int64(repoId)))).
		RETURNING(table.Repo.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &updatedRepo)
	if err != nil {
		log.Errorf("Can't update repo: %s", err)
		return model.Repo{}, err
	}

	return updatedRepo, nil
}

func UpdateRepoStatus(ctx context.Context, repoId int32, status RepoStatus, output string) (model.Repo, error) {
	var repo model.Repo

	stmt := table.Repo.
		UPDATE(table.Repo.Status, table.Repo.Output, table.Repo.UpdatedAt).
		SET(sqlite.String(string(status)), sqlite.String(output), sqlite.CURRENT_TIMESTAMP()).
		WHERE(table.Repo.ID.EQ(sqlite.Int(int64(repoId)))).
		RETURNING(table.Repo.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &repo)
	if err != nil {
		log.Errorf("Can't update repo: %s", err)
		return model.Repo{}, err
	}

	return repo, nil
}

func DeleteRepo(ctx context.Context, repoId int32) error {
	stmt := table.Repo.DELETE().
		WHERE(table.Repo.ID.EQ(sqlite.Int(int64(repoId))))

	log.Tracef("Query: %s", stmt.DebugSql())
	_, err := stmt.ExecContext(ctx, Db)
	if err != nil {
		log.Errorf("Can't delete repo: %s", err)
		return err
	}

	return nil
}

func CountRepoByNameOrPath(ctx context.Context, repoName, repoPath string) (int64, error) {
	var count struct {
		Count int64
	}

	stmt := sqlite.SELECT(sqlite.COUNT(table.Repo.Name).AS("count")).
		FROM(table.Repo.
			INNER_JOIN(table.ZfsPool, table.Repo.PoolID.EQ(table.ZfsPool.ID))).
		WHERE(
			table.Repo.Name.EQ(sqlite.String(repoName)).
				OR(table.ZfsPool.Path.EQ(sqlite.String(repoPath))),
		)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &count)
	if err != nil {
		log.Errorf("Can't query repo count: %s", err)
		return -1, err
	}

	return count.Count, nil
}
