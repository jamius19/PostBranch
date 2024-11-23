CREATE TABLE IF NOT EXISTS repo
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       VARCHAR(255) NOT NULL UNIQUE,
    pg_path    VARCHAR(2048) NOT NULL,
    version    INTEGER       NOT NULL,
    status     VARCHAR(50)   NOT NULL,
    output     TEXT,
    adapter    VARCHAR(50)   NOT NULL,
    pool_id    INTEGER      NOT NULL REFERENCES zfs_pool (id) ON DELETE CASCADE,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
);
