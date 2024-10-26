-- name: CountPg :one
SELECT COUNT(*)
FROM pg;

-- name: ListPg :many
SELECT *
FROM pg;

-- name: GetPg :one
SELECT *
FROM pg
WHERE id = ?;