CREATE TABLE pg
(
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    pg_path         VARCHAR(2048) NOT NULL,
    version         INTEGER       NOT NULL,
    status          VARCHAR(50)   NOT NULL,
    output          TEXT,
    repo_id         INTEGER       NOT NULL REFERENCES repo (id) ON DELETE CASCADE,
    created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP
);
