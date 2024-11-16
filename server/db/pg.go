package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/db/gen/table"
)

type PgStatus string

const (
	PgStarted   PgStatus = "STARTED"
	PgCompleted PgStatus = "COMPLETED"
	PgFailed    PgStatus = "FAILED"
)

func CreatePg(ctx context.Context, pg model.Pg) (model.Pg, error) {
	var newPg model.Pg

	stmt := table.Pg.INSERT(table.Pg.AllColumns).
		MODEL(pg).
		RETURNING(table.Pg.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &newPg)
	if err != nil {
		log.Errorf("Can't insert pg: %s", err)
		return model.Pg{}, err
	}

	return newPg, nil
}

func UpdatePg(ctx context.Context, pg model.Pg) (model.Pg, error) {
	var updatedPg model.Pg
	stmt := table.Pg.UPDATE(table.Pg.AllColumns.Except(table.Pg.ID, table.Pg.RepoID)).
		MODEL(pg).
		WHERE(table.Pg.ID.EQ(sqlite.Int(int64(*pg.ID)))).
		RETURNING(table.Pg.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &updatedPg)
	if err != nil {
		log.Errorf("Can't update pg: %s", err)
		return model.Pg{}, err
	}

	return updatedPg, nil
}

func UpdatePgStatus(ctx context.Context, pgId int32, status PgStatus, output string) (model.Pg, error) {
	var pg model.Pg

	stmt := table.Pg.UPDATE(table.Pg.Status).
		SET(
			table.Pg.Status.SET(sqlite.String(string(status))),
			table.Pg.Output.SET(sqlite.String(output)),
			table.Pg.UpdatedAt.SET(sqlite.CURRENT_TIMESTAMP()),
		).
		WHERE(table.Pg.ID.EQ(sqlite.Int(int64(pgId)))).
		RETURNING(table.Pg.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &pg)
	if err != nil {
		log.Errorf("Can't update pg status: %s", err)
		return model.Pg{}, err
	}

	return pg, nil
}
