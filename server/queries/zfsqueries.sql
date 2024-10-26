-- name: GetDataset :one
SELECT * FROM zfs_dataset WHERE id = ?;

-- name: ListDataset :many
SELECT * FROM zfs_dataset;

-- name: GetPool :one
SELECT * FROM zfs_pool WHERE id = ?;

-- name: ListPool :many
SELECT * FROM zfs_pool;
