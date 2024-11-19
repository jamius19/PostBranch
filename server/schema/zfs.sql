CREATE TABLE zfs_pool
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    path       VARCHAR(2048) NOT NULL,
    size_in_mb BIGINT        NOT NULL,
    name       VARCHAR(255)  NOT NULL,
    mount_path VARCHAR(2048) NOT NULL,
    pool_type  VARCHAR(60)   NOT NULL,
    created_at DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP
);
