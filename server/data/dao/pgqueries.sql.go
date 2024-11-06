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
INSERT INTO pg (pg_path,
                version,
                stop_pg,
                pg_user,
                connection_type,
                host,
                port,
                ssl_mode,
                username,
                password,
                status,
                repo_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, pg_path, version, stop_pg, pg_user, connection_type, host, port, username, password, ssl_mode, status, output, repo_id, created_at, updated_at
`

type CreatePgParams struct {
	PgPath         string
	Version        int64
	StopPg         bool
	PgUser         string
	ConnectionType string
	Host           sql.NullString
	Port           sql.NullInt64
	SslMode        sql.NullString
	Username       sql.NullString
	Password       sql.NullString
	Status         string
	RepoID         int64
}

func (q *Queries) CreatePg(ctx context.Context, arg CreatePgParams) (Pg, error) {
	row := q.db.QueryRowContext(ctx, createPg,
		arg.PgPath,
		arg.Version,
		arg.StopPg,
		arg.PgUser,
		arg.ConnectionType,
		arg.Host,
		arg.Port,
		arg.SslMode,
		arg.Username,
		arg.Password,
		arg.Status,
		arg.RepoID,
	)
	var i Pg
	err := row.Scan(
		&i.ID,
		&i.PgPath,
		&i.Version,
		&i.StopPg,
		&i.PgUser,
		&i.ConnectionType,
		&i.Host,
		&i.Port,
		&i.Username,
		&i.Password,
		&i.SslMode,
		&i.Status,
		&i.Output,
		&i.RepoID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPg = `-- name: GetPg :one
SELECT id, pg_path, version, stop_pg, pg_user, connection_type, host, port, username, password, ssl_mode, status, output, repo_id, created_at, updated_at
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
		&i.ConnectionType,
		&i.Host,
		&i.Port,
		&i.Username,
		&i.Password,
		&i.SslMode,
		&i.Status,
		&i.Output,
		&i.RepoID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updatePg = `-- name: UpdatePg :one
UPDATE pg
SET pg_path         = ?,
    version         = ?,
    stop_pg         = ?,
    pg_user         = ?,
    connection_type = ?,
    host            = ?,
    port            = ?,
    ssl_mode        =?,
    username        = ?,
    password        = ?,
    status          = ?,
    updated_at      = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, pg_path, version, stop_pg, pg_user, connection_type, host, port, username, password, ssl_mode, status, output, repo_id, created_at, updated_at
`

type UpdatePgParams struct {
	PgPath         string
	Version        int64
	StopPg         bool
	PgUser         string
	ConnectionType string
	Host           sql.NullString
	Port           sql.NullInt64
	SslMode        sql.NullString
	Username       sql.NullString
	Password       sql.NullString
	Status         string
	ID             int64
}

func (q *Queries) UpdatePg(ctx context.Context, arg UpdatePgParams) (Pg, error) {
	row := q.db.QueryRowContext(ctx, updatePg,
		arg.PgPath,
		arg.Version,
		arg.StopPg,
		arg.PgUser,
		arg.ConnectionType,
		arg.Host,
		arg.Port,
		arg.SslMode,
		arg.Username,
		arg.Password,
		arg.Status,
		arg.ID,
	)
	var i Pg
	err := row.Scan(
		&i.ID,
		&i.PgPath,
		&i.Version,
		&i.StopPg,
		&i.PgUser,
		&i.ConnectionType,
		&i.Host,
		&i.Port,
		&i.Username,
		&i.Password,
		&i.SslMode,
		&i.Status,
		&i.Output,
		&i.RepoID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updatePgStatus = `-- name: UpdatePgStatus :one
UPDATE pg
SET status     = ?,
    output     = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, pg_path, version, stop_pg, pg_user, connection_type, host, port, username, password, ssl_mode, status, output, repo_id, created_at, updated_at
`

type UpdatePgStatusParams struct {
	Status string
	Output sql.NullString
	ID     int64
}

func (q *Queries) UpdatePgStatus(ctx context.Context, arg UpdatePgStatusParams) (Pg, error) {
	row := q.db.QueryRowContext(ctx, updatePgStatus, arg.Status, arg.Output, arg.ID)
	var i Pg
	err := row.Scan(
		&i.ID,
		&i.PgPath,
		&i.Version,
		&i.StopPg,
		&i.PgUser,
		&i.ConnectionType,
		&i.Host,
		&i.Port,
		&i.Username,
		&i.Password,
		&i.SslMode,
		&i.Status,
		&i.Output,
		&i.RepoID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
