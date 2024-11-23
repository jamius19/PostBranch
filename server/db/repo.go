package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/db/gen/table"
	"strings"
	"time"
)

type RepoDetail struct {
	Repo     model.Repo
	Pool     model.ZfsPool `alias:"pool"`
	Pg       model.Pg
	Branches []model.Branch `alias:"branches"`
}

func ListRepo(ctx context.Context) ([]RepoDetail, error) {
	var repoDetailList []RepoDetail

	stmt := sqlite.SELECT(
		table.Repo.AllColumns,
		table.ZfsPool.AllColumns.As("pool"),
		table.Pg.AllColumns,
		table.Branch.AllColumns.As("branches"),
	).
		FROM(table.Repo.
			INNER_JOIN(table.ZfsPool, table.Repo.PoolID.EQ(table.ZfsPool.ID)).
			INNER_JOIN(table.Pg, table.Pg.RepoID.EQ(table.Repo.ID)).
			LEFT_JOIN(table.Branch, table.Branch.RepoID.EQ(table.Repo.ID))).
		ORDER_BY(table.Repo.CreatedAt.DESC())

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
		table.Pg.AllColumns,
		table.Branch.AllColumns.As("branches"),
	).
		FROM(table.Repo.
			INNER_JOIN(table.ZfsPool, table.Repo.PoolID.EQ(table.ZfsPool.ID)).
			INNER_JOIN(table.Pg, table.Pg.RepoID.EQ(table.Repo.ID)).
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
		table.Pg.AllColumns,
		table.Branch.AllColumns.As("branches"),
	).
		FROM(table.Repo.
			INNER_JOIN(table.ZfsPool, table.Repo.PoolID.EQ(table.ZfsPool.ID)).
			INNER_JOIN(table.Pg, table.Pg.RepoID.EQ(table.Repo.ID)).
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

	stmt := table.Repo.INSERT(table.Repo.Name, table.Repo.PoolID).
		VALUES(repo.Name, repo.PoolID).
		RETURNING(table.Repo.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &newRepo)
	if err != nil {
		log.Errorf("Can't create repo: %s", err)
		return model.Repo{}, err
	}

	return newRepo, nil
}

func DeleteRepo(ctx context.Context, repoId int64) error {
	stmt := table.Repo.DELETE().
		WHERE(table.Repo.ID.EQ(sqlite.Int(repoId)))

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
