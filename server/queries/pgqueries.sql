-- name: GetPg :one
SELECT *
FROM pg
WHERE id = ?;

-- name: CreatePg :one
INSERT INTO pg (pg_path,
                version,
                status,
                repo_id)
VALUES (?, ?, ?, ?)
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
    status          = ?,
    updated_at      = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;
