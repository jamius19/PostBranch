-- name: GetPg :one
SELECT *
FROM pg
WHERE id = ?;

-- name: CreatePg :one
INSERT INTO pg (pg_path, version, stop_pg, pg_user, connection_type, host, port, username, password, status,
                repo_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdatePgStatus :one
UPDATE pg
SET status     = ?,
    output     = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdatePg :one
UPDATE pg
SET pg_path         = ?,
    version         = ?,
    stop_pg         = ?,
    pg_user         = ?,
    connection_type = ?,
    host            = ?,
    port            = ?,
    username        = ?,
    password        = ?,
    status          = ?,
    updated_at      = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;
