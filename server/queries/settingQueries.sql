-- name: GetSetting :one
SELECT *
FROM settings
WHERE key = ?;

-- name: CreateSetting :exec
INSERT INTO settings (key, value, json)
VALUES (?, ?, ?);
