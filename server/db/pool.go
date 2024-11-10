package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/db/gen/table"
)

func ListPool(ctx context.Context) ([]model.ZfsPool, error) {
	var pools []model.ZfsPool

	stmt := sqlite.SELECT(table.ZfsPool.AllColumns).
		FROM(table.ZfsPool)

	log.Debugf("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &pools)
	if err != nil {
		log.Errorf("Can't query pools: %s", err)
		return nil, err
	}

	return pools, nil
}

func CreatePool(ctx context.Context, pool model.ZfsPool) (model.ZfsPool, error) {
	var newPool model.ZfsPool

	stmt := table.ZfsPool.INSERT(table.ZfsPool.AllColumns).
		MODEL(pool).
		RETURNING(table.ZfsPool.AllColumns)

	log.Debugf("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &newPool)
	if err != nil {
		log.Errorf("Can't insert pool: %s", err)
		return model.ZfsPool{}, err
	}

	return newPool, nil
}

func DeletePool(ctx context.Context, poolId int32) error {
	stmt := table.ZfsPool.DELETE().
		WHERE(table.ZfsPool.ID.EQ(sqlite.Int32(poolId)))

	log.Debugf("Query: %s", stmt.DebugSql())

	_, err := stmt.ExecContext(ctx, Db)
	if err != nil {
		log.Errorf("Can't delete pool: %s", err)
		return err
	}

	return nil
}
