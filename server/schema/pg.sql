CREATE TABLE pg
(
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    pg_path         VARCHAR(2048) NOT NULL,
    version         INTEGER       NOT NULL,
    stop_pg         BOOLEAN       NOT NULL,
    pg_user         VARCHAR(255)  NOT NULL,
    connection_type VARCHAR(20)   NOT NULL DEFAULT false,
    host            VARCHAR(255),
    port            INTEGER,
    username        VARCHAR(255),
    password        VARCHAR(255),
    ssl_mode        VARCHAR(20),
    status          VARCHAR(50)   NOT NULL,
    output          TEXT,
    repo_id         INTEGER       NOT NULL REFERENCES repo (id) ON DELETE CASCADE,
    created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP
);
