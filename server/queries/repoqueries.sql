-- name: GetRepo :one
SELECT rp.id              AS repo_id,
       rp.name            AS repo_name,
       rp.created_at      AS repo_created_at,
       rp.updated_at      AS repo_updated_at,
       zp.id              AS pool_id,
       zp.path            AS pool_path,
       zp.size_in_mb      AS pool_size_in_mb,
       zp.name            AS pool_name,
       zp.mount_path      AS pool_mount_path,
       zp.pool_type       AS pool_type,
       zp.created_at      AS pool_created_at,
       zp.updated_at      AS pool_updated_at,
       pg.id              AS pg_id,
       pg.pg_path         AS pg_path,
       pg.version         AS pg_version,
       pg.stop_pg         AS pg_stop_pg,
       pg.pg_user         AS pg_pg_user,
       pg.connection_type AS pg_connection_type,
       pg.host            AS pg_host,
       pg.port            AS pg_port,
       pg.username        AS pg_username,
       pg.password        AS pg_password,
       pg.status          AS pg_status,
       pg.output          AS pg_output,
       pg.created_at      AS pg_created_at,
       pg.updated_at      AS pg_updated_at
FROM repo rp
         JOIN zfs_pool zp on rp.pool_id = zp.id
         LEFT JOIN main.pg pg on rp.id = pg.repo_id
WHERE rp.id = ?;

-- name: ListRepo :many
SELECT rp.id              AS repo_id,
       rp.name            AS repo_name,
       rp.created_at      AS repo_created_at,
       rp.updated_at      AS repo_updated_at,
       zp.id              AS pool_id,
       zp.path            AS pool_path,
       zp.size_in_mb      AS pool_size_in_mb,
       zp.name            AS pool_name,
       zp.mount_path      AS pool_mount_path,
       zp.pool_type       AS pool_type,
       zp.created_at      AS pool_created_at,
       zp.updated_at      AS pool_updated_at,
       pg.id              AS pg_id,
       pg.pg_path         AS pg_path,
       pg.version         AS pg_version,
       pg.stop_pg         AS pg_stop_pg,
       pg.pg_user         AS pg_pg_user,
       pg.connection_type AS pg_connection_type,
       pg.host            AS pg_host,
       pg.port            AS pg_port,
       pg.username        AS pg_username,
       pg.password        AS pg_password,
       pg.status          AS pg_status,
       pg.output          AS pg_output,
       pg.created_at      AS pg_created_at,
       pg.updated_at      AS pg_updated_at
FROM repo rp
         JOIN zfs_pool zp on rp.pool_id = zp.id
         LEFT JOIN main.pg pg on rp.id = pg.repo_id
ORDER BY rp.created_at DESC;

-- name: ListBranchesByRepoId :many
SELECT *
FROM branch
WHERE repo_id = ?;

-- name: UpdatePgRepo :one
UPDATE pg
SET repo_id    = ?,
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
INSERT INTO repo (name, pool_id)
VALUES (?, ?)
RETURNING *;

-- name: CreateBranch :one
INSERT INTO branch (name, repo_id, parent_id, dataset_id)
VALUES (?, ?, ?, ?)
RETURNING *;
