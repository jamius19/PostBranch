// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: pgqueries.sql

package dao

import (
	"context"
	"database/sql"
)

const createPg = `-- name: CreatePg :one
INSERT INTO pg (pg_path, version, stop_pg, pg_user, custom_connection, host, port, username, password, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, pg_path, version, stop_pg, pg_user, custom_connection, host, port, username, password, status, output, created_at, updated_at
`

type CreatePgParams struct {
	PgPath           string
	Version          int64
	StopPg           bool
	PgUser           string
	CustomConnection bool
	Host             sql.NullString
	Port             sql.NullInt64
	Username         sql.NullString
	Password         sql.NullString
	Status           string
}

func (q *Queries) CreatePg(ctx context.Context, arg CreatePgParams) (Pg, error) {
	row := q.db.QueryRowContext(ctx, createPg,
		arg.PgPath,
		arg.Version,
		arg.StopPg,
		arg.PgUser,
		arg.CustomConnection,
		arg.Host,
		arg.Port,
		arg.Username,
		arg.Password,
		arg.Status,
	)
	var i Pg
	err := row.Scan(
		&i.ID,
		&i.PgPath,
		&i.Version,
		&i.StopPg,
		&i.PgUser,
		&i.CustomConnection,
		&i.Host,
		&i.Port,
		&i.Username,
		&i.Password,
		&i.Status,
		&i.Output,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPg = `-- name: GetPg :one
SELECT id, pg_path, version, stop_pg, pg_user, custom_connection, host, port, username, password, status, output, created_at, updated_at
FROM pg
WHERE id = ?
`

func (q *Queries) GetPg(ctx context.Context, id int64) (Pg, error) {
	row := q.db.QueryRowContext(ctx, getPg, id)
	var i Pg
	err := row.Scan(
		&i.ID,
		&i.PgPath,
		&i.Version,
		&i.StopPg,
		&i.PgUser,
		&i.CustomConnection,
		&i.Host,
		&i.Port,
		&i.Username,
		&i.Password,
		&i.Status,
		&i.Output,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updatePg = `-- name: UpdatePg :one
UPDATE pg
SET status = ?,
    output = ?
WHERE id = ?
RETURNING id, pg_path, version, stop_pg, pg_user, custom_connection, host, port, username, password, status, output, created_at, updated_at
`

type UpdatePgParams struct {
	Status string
	Output sql.NullString
	ID     int64
}

func (q *Queries) UpdatePg(ctx context.Context, arg UpdatePgParams) (Pg, error) {
	row := q.db.QueryRowContext(ctx, updatePg, arg.Status, arg.Output, arg.ID)
	var i Pg
	err := row.Scan(
		&i.ID,
		&i.PgPath,
		&i.Version,
		&i.StopPg,
		&i.PgUser,
		&i.CustomConnection,
		&i.Host,
		&i.Port,
		&i.Username,
		&i.Password,
		&i.Status,
		&i.Output,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
