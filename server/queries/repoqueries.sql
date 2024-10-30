-- name: GetRepo :one
SELECT *
FROM repo
WHERE id = ?;

-- name: ListRepo :many
SELECT sqlc.embed(rp), sqlc.embed(zp)
FROM repo rp
         JOIN zfs_pool zp on rp.pool_id = zp.id;

-- name: CountRepo :one
SELECT COUNT(*)
FROM repo;

-- name: ListRepoNames :many
SELECT name
FROM repo;

-- name: CountRepoByNameOrPath :one
SELECT COUNT(*)
FROM repo rp
         JOIN zfs_pool zp on rp.pool_id = zp.id
WHERE rp.name = ?
   OR zp.path = ?;

-- name: CreateRepo :one
INSERT INTO repo (name, repo_type, pool_id)
VALUES (?, ?, ?)
RETURNING *;

-- UpdateRepoPg :one
UPDATE repo
SET pg_id = ?
WHERE id = ?
RETURNING *;
