-- name: GetSetting :one
SELECT value
FROM settings
WHERE key = ?;

-- name: CreateSetting :one
INSERT INTO settings (key, value)
VALUES (?, ?)
RETURNING *;
