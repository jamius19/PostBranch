-- name: GetDataset :one
SELECT *
FROM zfs_dataset
WHERE id = ?;

-- name: ListDataset :many
SELECT *
FROM zfs_dataset;

-- name: CreateDataset :one
INSERT INTO zfs_dataset (name, pool_id)
VALUES (?, ?)
RETURNING *;

-- name: GetPool :one
SELECT *
FROM zfs_pool
WHERE id = ?;

-- name: ListPool :many
SELECT *
FROM zfs_pool;

-- name: CreatePool :one
INSERT INTO zfs_pool (name, path)
VALUES (?, ?)
RETURNING *;
