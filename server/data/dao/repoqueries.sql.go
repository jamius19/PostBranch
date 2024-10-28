// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: repoqueries.sql

package dao

import (
	"context"
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

const countRepoByNameOrPath = `-- name: CountRepoByNameOrPath :one
SELECT COUNT(*)
FROM repo rp
         JOIN zfs_pool zp on rp.pool_id = zp.id
WHERE rp.name = ?
   OR zp.path = ?
`

type CountRepoByNameOrPathParams struct {
	Name string
	Path string
}

func (q *Queries) CountRepoByNameOrPath(ctx context.Context, arg CountRepoByNameOrPathParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, countRepoByNameOrPath, arg.Name, arg.Path)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createRepo = `-- name: CreateRepo :one
INSERT INTO repo (name, repo_type, size, size_unit, pool_id)
VALUES (?, ?, ?, ?, ?)
RETURNING id, name, repo_type, size, size_unit, pool_id, pg_id, created_at, updated_at
`

type CreateRepoParams struct {
	Name     string
	RepoType string
	Size     int64
	SizeUnit string
	PoolID   int64
}

func (q *Queries) CreateRepo(ctx context.Context, arg CreateRepoParams) (Repo, error) {
	row := q.db.QueryRowContext(ctx, createRepo,
		arg.Name,
		arg.RepoType,
		arg.Size,
		arg.SizeUnit,
		arg.PoolID,
	)
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
SELECT rp.id, rp.name, rp.repo_type, rp.size, rp.size_unit, rp.pool_id, rp.pg_id, rp.created_at, rp.updated_at, zp.id, zp.path, zp.name, zp.created_at, zp.updated_at
FROM repo rp
         JOIN zfs_pool zp on rp.pool_id = zp.id
`

type ListRepoRow struct {
	Repo    Repo
	ZfsPool ZfsPool
}

func (q *Queries) ListRepo(ctx context.Context) ([]ListRepoRow, error) {
	rows, err := q.db.QueryContext(ctx, listRepo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListRepoRow
	for rows.Next() {
		var i ListRepoRow
		if err := rows.Scan(
			&i.Repo.ID,
			&i.Repo.Name,
			&i.Repo.RepoType,
			&i.Repo.Size,
			&i.Repo.SizeUnit,
			&i.Repo.PoolID,
			&i.Repo.PgID,
			&i.Repo.CreatedAt,
			&i.Repo.UpdatedAt,
			&i.ZfsPool.ID,
			&i.ZfsPool.Path,
			&i.ZfsPool.Name,
			&i.ZfsPool.CreatedAt,
			&i.ZfsPool.UpdatedAt,
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

const listRepoNames = `-- name: ListRepoNames :many
SELECT name
FROM repo
`

func (q *Queries) ListRepoNames(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, listRepoNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
