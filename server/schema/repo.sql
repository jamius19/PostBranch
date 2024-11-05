CREATE TABLE IF NOT EXISTS repo
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       VARCHAR(255) NOT NULL UNIQUE,
    pool_id    INTEGER      NOT NULL REFERENCES zfs_pool (id) ON DELETE CASCADE,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
)
