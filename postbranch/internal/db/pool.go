package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	model2 "github.com/jamius19/postbranch/internal/db/gen/model"
	table2 "github.com/jamius19/postbranch/internal/db/gen/table"
	"time"
)

type PoolDetail struct {
	Pool model2.ZfsPool `alias:"pool"`
	Pg   model2.Pg      `alias:"pg"`
}

func ListPoolDetail(ctx context.Context) ([]PoolDetail, error) {
	var pools []PoolDetail

	stmt := sqlite.
		SELECT(
			table2.ZfsPool.AllColumns.As("pool"),
			table2.Pg.AllColumns.As("pg"),
		).
		FROM(
			table2.ZfsPool.
				INNER_JOIN(table2.Repo, table2.Repo.PoolID.EQ(table2.ZfsPool.ID)).
				INNER_JOIN(table2.Pg, table2.Pg.RepoID.EQ(table2.Repo.ID)),
		)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &pools)
	if err != nil {
		log.Errorf("Can't query pools: %s", err)
		return nil, err
	}

	return pools, nil
}

func CreatePool(ctx context.Context, pool model2.ZfsPool) (model2.ZfsPool, error) {
	var newPool model2.ZfsPool

	pool.CreatedAt = time.Now().UTC()
	pool.UpdatedAt = time.Now().UTC()

	stmt := table2.ZfsPool.
		INSERT(table2.ZfsPool.AllColumns).
		MODEL(pool).
		RETURNING(table2.ZfsPool.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &newPool)
	if err != nil {
		log.Errorf("Can't insert pool: %s", err)
		return model2.ZfsPool{}, err
	}

	return newPool, nil
}

func DeletePool(ctx context.Context, poolId int32) error {
	stmt := table2.ZfsPool.
		DELETE().
		WHERE(table2.ZfsPool.ID.EQ(sqlite.Int32(poolId)))

	log.Tracef("Query: %s", stmt.DebugSql())

	_, err := stmt.ExecContext(ctx, Db)
	if err != nil {
		log.Errorf("Can't delete pool: %s", err)
		return err
	}

	return nil
}
