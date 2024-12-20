package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	"github.com/jamius19/postbranch/internal/db/gen/model"
	"github.com/jamius19/postbranch/internal/db/gen/table"
	"time"
)

type PoolDetail struct {
	Repo model.Repo    `alias:"repo"`
	Pool model.ZfsPool `alias:"pool"`
}

func ListPoolDetail(ctx context.Context) ([]PoolDetail, error) {
	var pools []PoolDetail

	stmt := sqlite.
		SELECT(
			table.ZfsPool.AllColumns.As("pool"),
			table.Repo.AllColumns.As("repo"),
		).
		FROM(
			table.ZfsPool.
				INNER_JOIN(table.Repo, table.Repo.PoolID.EQ(table.ZfsPool.ID)).
				INNER_JOIN(table.Repo, table.Repo.PoolID.EQ(table.ZfsPool.ID)),
		)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &pools)
	if err != nil {
		log.Errorf("Can't query pools: %s", err)
		return nil, err
	}

	return pools, nil
}

func CreatePool(ctx context.Context, pool model.ZfsPool) (model.ZfsPool, error) {
	var newPool model.ZfsPool

	pool.CreatedAt = time.Now().UTC()
	pool.UpdatedAt = time.Now().UTC()

	stmt := table.ZfsPool.
		INSERT(table.ZfsPool.AllColumns).
		MODEL(pool).
		RETURNING(table.ZfsPool.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &newPool)
	if err != nil {
		log.Errorf("Can't insert pool: %s", err)
		return model.ZfsPool{}, err
	}

	return newPool, nil
}

func DeletePool(ctx context.Context, poolId int32) error {
	stmt := table.ZfsPool.
		DELETE().
		WHERE(table.ZfsPool.ID.EQ(sqlite.Int32(poolId)))

	log.Tracef("Query: %s", stmt.DebugSql())

	_, err := stmt.ExecContext(ctx, Db)
	if err != nil {
		log.Errorf("Can't delete pool: %s", err)
		return err
	}

	return nil
}
