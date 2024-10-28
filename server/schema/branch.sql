CREATE TABLE IF NOT EXISTS branch
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       VARCHAR(255) NOT NULL UNIQUE,
    path       VARCHAR(255) NOT NULL,
    parent_id  INTEGER REFERENCES branch (id) ON DELETE CASCADE,
    dataset_id INTEGER      NOT NULL REFERENCES zfs_dataset (id) ON DELETE CASCADE,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
);
