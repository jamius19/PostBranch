-- name: GetRepo :one
SELECT sqlc.embed(rp), sqlc.embed(zp)
FROM repo rp
         JOIN zfs_pool zp on rp.pool_id = zp.id
WHERE rp.id = ?;

-- name: ListRepo :many
SELECT sqlc.embed(rp), sqlc.embed(zp)
FROM repo rp
         JOIN zfs_pool zp on rp.pool_id = zp.id
ORDER BY rp.created_at DESC;

-- name: ListBranchesByRepoId :many
SELECT *
FROM branch
WHERE repo_id = ?;

-- name: UpdateRepoPg :one
UPDATE repo
SET pg_id      = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: CountRepo :one
SELECT COUNT(*)
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

-- name: CreateBranch :one
INSERT INTO branch (name, repo_id, parent_id, dataset_id)
VALUES (?, ?, ?, ?)
RETURNING *;
