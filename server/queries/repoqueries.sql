-- name: GetRepo :one
SELECT *
FROM repo
WHERE id = ?;

-- name: ListRepo :many
SELECT *
FROM repo;

-- name: CountRepo :one
SELECT COUNT(*)
FROM repo;

-- name: CreateRepo :one
INSERT INTO repo (name, repo_type, size, size_unit, pool_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- UpdateRepoPg :one
UPDATE repo
SET pg_id = ?
WHERE id = ?
RETURNING *;
