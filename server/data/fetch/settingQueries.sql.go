// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: settingQueries.sql

package fetch

import (
	"context"
	"database/sql"
)

const createSetting = `-- name: CreateSetting :exec
INSERT INTO settings (key, value, json)
VALUES (?, ?, ?)
`

type CreateSettingParams struct {
	Key   string
	Value string
	Json  sql.NullString
}

func (q *Queries) CreateSetting(ctx context.Context, arg CreateSettingParams) error {
	_, err := q.db.ExecContext(ctx, createSetting, arg.Key, arg.Value, arg.Json)
	return err
}

const getSetting = `-- name: GetSetting :one
SELECT id, "key", value, json, created_at, updated_at
FROM settings
WHERE key = ?
`

func (q *Queries) GetSetting(ctx context.Context, key string) (Setting, error) {
	row := q.db.QueryRowContext(ctx, getSetting, key)
	var i Setting
	err := row.Scan(
		&i.ID,
		&i.Key,
		&i.Value,
		&i.Json,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
