package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	model2 "github.com/jamius19/postbranch/internal/db/gen/model"
	table2 "github.com/jamius19/postbranch/internal/db/gen/table"
	"strings"
	"time"
)

type RepoDetail struct {
	Repo     model2.Repo
	Pool     model2.ZfsPool `alias:"pool"`
	Pg       model2.Pg
	Branches []model2.Branch `alias:"branches"`
}

func ListRepo(ctx context.Context) ([]RepoDetail, error) {
	var repoDetailList []RepoDetail

	stmt := sqlite.SELECT(
		table2.Repo.AllColumns,
		table2.ZfsPool.AllColumns.As("pool"),
		table2.Pg.AllColumns,
		table2.Branch.AllColumns.As("branches"),
	).
		FROM(table2.Repo.
			INNER_JOIN(table2.ZfsPool, table2.Repo.PoolID.EQ(table2.ZfsPool.ID)).
			INNER_JOIN(table2.Pg, table2.Pg.RepoID.EQ(table2.Repo.ID)).
			LEFT_JOIN(table2.Branch, table2.Branch.RepoID.EQ(table2.Repo.ID))).
		ORDER_BY(table2.Repo.CreatedAt.DESC())

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
		table2.Repo.AllColumns,
		table2.ZfsPool.AllColumns.As("pool"),
		table2.Pg.AllColumns,
		table2.Branch.AllColumns.As("branches"),
	).
		FROM(table2.Repo.
			INNER_JOIN(table2.ZfsPool, table2.Repo.PoolID.EQ(table2.ZfsPool.ID)).
			INNER_JOIN(table2.Pg, table2.Pg.RepoID.EQ(table2.Repo.ID)).
			LEFT_JOIN(table2.Branch, table2.Branch.RepoID.EQ(table2.Repo.ID))).
		WHERE(table2.Repo.ID.EQ(sqlite.Int(repoId))).
		ORDER_BY(table2.Branch.CreatedAt.DESC())

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
		table2.Repo.AllColumns,
		table2.ZfsPool.AllColumns.As("pool"),
		table2.Pg.AllColumns,
		table2.Branch.AllColumns.As("branches"),
	).
		FROM(table2.Repo.
			INNER_JOIN(table2.ZfsPool, table2.Repo.PoolID.EQ(table2.ZfsPool.ID)).
			INNER_JOIN(table2.Pg, table2.Pg.RepoID.EQ(table2.Repo.ID)).
			LEFT_JOIN(table2.Branch, table2.Branch.RepoID.EQ(table2.Repo.ID))).
		WHERE(table2.Repo.Name.EQ(sqlite.String(strings.TrimSpace(repoName)))).
		ORDER_BY(table2.Branch.CreatedAt.DESC())

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &repoDetail)
	if err != nil {
		log.Errorf("Can't query repos: %s", err)
		return RepoDetail{}, err
	}

	return repoDetail, nil
}

func CreateRepo(ctx context.Context, repo model2.Repo) (model2.Repo, error) {
	var newRepo model2.Repo

	repo.CreatedAt = time.Now().UTC()
	repo.UpdatedAt = time.Now().UTC()

	stmt := table2.Repo.INSERT(table2.Repo.Name, table2.Repo.PoolID).
		VALUES(repo.Name, repo.PoolID).
		RETURNING(table2.Repo.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &newRepo)
	if err != nil {
		log.Errorf("Can't create repo: %s", err)
		return model2.Repo{}, err
	}

	return newRepo, nil
}

func DeleteRepo(ctx context.Context, repoId int64) error {
	stmt := table2.Repo.DELETE().
		WHERE(table2.Repo.ID.EQ(sqlite.Int(repoId)))

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

	stmt := sqlite.SELECT(sqlite.COUNT(table2.Repo.Name).AS("count")).
		FROM(table2.Repo.
			INNER_JOIN(table2.ZfsPool, table2.Repo.PoolID.EQ(table2.ZfsPool.ID))).
		WHERE(
			table2.Repo.Name.EQ(sqlite.String(repoName)).
				OR(table2.ZfsPool.Path.EQ(sqlite.String(repoPath))),
		)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &count)
	if err != nil {
		log.Errorf("Can't query repo count: %s", err)
		return -1, err
	}

	return count.Count, nil
}
