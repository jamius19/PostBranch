-- name: GetPg :one
SELECT *
FROM pg
WHERE id = ?;

-- name: CreatePg :one
INSERT INTO pg (pg_path, version, stop_pg, pg_user, custom_connection, host, port, username, password, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdatePg :one
UPDATE pg
SET status = ?,
    output = ?
WHERE id = ?
RETURNING *;
