-- name: GetRepo :one
SELECT * FROM repo WHERE id = ?;

-- name: ListRepo :many
SELECT * FROM repo;

-- name: CountRepo :one
SELECT COUNT(*) FROM repo;
