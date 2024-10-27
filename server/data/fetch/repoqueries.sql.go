// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: repoqueries.sql

package fetch

import (
	"context"
	"database/sql"
)

const countRepo = `-- name: CountRepo :one
SELECT COUNT(*)
FROM repo
`

func (q *Queries) CountRepo(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countRepo)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createRepo = `-- name: CreateRepo :execresult
INSERT INTO repo (name, repo_type, size, size_unit, pool_id)
VALUES (?, ?, ?, ?, ?)
`

type CreateRepoParams struct {
	Name     string
	RepoType string
	Size     int64
	SizeUnit string
	PoolID   int64
}

func (q *Queries) CreateRepo(ctx context.Context, arg CreateRepoParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, createRepo,
		arg.Name,
		arg.RepoType,
		arg.Size,
		arg.SizeUnit,
		arg.PoolID,
	)
}

const getRepo = `-- name: GetRepo :one
SELECT id, name, repo_type, size, size_unit, pool_id, pg_id, created_at, updated_at
FROM repo
WHERE id = ?
`

func (q *Queries) GetRepo(ctx context.Context, id int64) (Repo, error) {
	row := q.db.QueryRowContext(ctx, getRepo, id)
	var i Repo
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.RepoType,
		&i.Size,
		&i.SizeUnit,
		&i.PoolID,
		&i.PgID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listRepo = `-- name: ListRepo :many
SELECT id, name, repo_type, size, size_unit, pool_id, pg_id, created_at, updated_at
FROM repo
`

func (q *Queries) ListRepo(ctx context.Context) ([]Repo, error) {
	rows, err := q.db.QueryContext(ctx, listRepo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Repo
	for rows.Next() {
		var i Repo
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.RepoType,
			&i.Size,
			&i.SizeUnit,
			&i.PoolID,
			&i.PgID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
