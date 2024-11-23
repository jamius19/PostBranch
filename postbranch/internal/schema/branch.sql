CREATE TABLE IF NOT EXISTS branch
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       VARCHAR(255) NOT NULL,
    status     VARCHAR(50)  NOT NULL,
    pg_status  VARCHAR(50)  NOT NULL,
    pg_port       INTEGER      NOT NULL,
    repo_id    INTEGER      NOT NULL REFERENCES repo (id) ON DELETE CASCADE,
    parent_id  INTEGER REFERENCES branch (id) ON DELETE CASCADE,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (repo_id, name)
);
