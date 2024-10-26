// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: databaseQueries.sql

package fetch

import (
	"context"
)

const countDb = `-- name: CountDb :one
SELECT COUNT(*)
FROM databases
`

func (q *Queries) CountDb(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, countDb)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getDb = `-- name: GetDb :one
SELECT id, name, path, parent, created_at, updated_at
FROM databases
WHERE id = ?
`

func (q *Queries) GetDb(ctx context.Context, id int64) (Database, error) {
	row := q.db.QueryRowContext(ctx, getDb, id)
	var i Database
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Path,
		&i.Parent,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
