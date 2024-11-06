package dao

import (
	"database/sql"
	"fmt"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service/pg"
	"github.com/jamius19/postbranch/web/responseerror"
)

var log = logger.Logger
var PgSuperUserCheckQuery = `select usesuper from pg_user where usename = CURRENT_USER;`

var PgConfigFilePathsQuery = `SELECT DISTINCT(sourcefile) FROM pg_settings WHERE source = 'configuration file';`

var PgHbaFilePathsQuery = `SELECT DISTINCT(file_name) FROM pg_hba_file_rules;`

// TODO: Fix potential sql injection
var PgLocalReplicationCheckQuery = `SELECT CASE 
           WHEN EXISTS (
               SELECT 1
               FROM pg_hba_file_rules
               WHERE type = 'local'
                 AND 'replication' = ANY(database)
                 AND auth_method IN ('trust', 'peer')
                 AND ('%s' = ANY(user_name) OR 'all' = ANY(user_name))
           ) 
           THEN 'REPLICATION_ALLOWED'
           ELSE 'REPLICATION_NOT_ALLOWED'
       END AS replication_status;`

// TODO: Fix potential sql injection
var PgHostReplicationCheckQuery = `SELECT CASE 
           WHEN EXISTS (
               SELECT 1
               FROM pg_hba_file_rules
               WHERE type = 'host'
                 AND 'replication' = ANY(database)
                 AND auth_method IN ('md5', 'scram-sha-256')
                 AND ('%s' = ANY(user_name) OR 'all' = ANY(user_name))
           ) 
           THEN 'REPLICATION_ALLOWED'
           ELSE 'REPLICATION_NOT_ALLOWED'
       END AS replication_status;`

const errMsg = "Can't connect to PostgreSQL. Is it running and is the provided configuration correct?"

func GetConnString(pg pg.AuthInfo) string {
	return fmt.Sprintf(
		"user=%s host=%s port=%d password=%s dbname=postgres sslmode=%s",
		pg.GetDbUsername(),
		pg.GetHost(),
		pg.GetPort(),
		pg.GetPassword(),
		pg.GetSslMode(),
	)
}

func RunQuery(pgInit *repo.PgInitDto, query string) (*sql.DB, *sql.Rows, func(), error) {
	cleanup := func() {}

	db, err := sql.Open("postgres", GetConnString(pgInit))
	if err != nil {
		log.Errorf("Failed to open db: %v", err)
		return nil, nil, cleanup, responseerror.From(errMsg)
	}

	rows, err := db.Query(query)
	if err != nil {
		log.Errorf("Failed to run query: %s, error: %v", query, err)

		if db != nil {
			if err := db.Close(); err != nil {
				log.Errorf("Failed to close db: %v", err)
			}
		}

		return nil, nil, cleanup, responseerror.From(errMsg)
	}

	cleanup = func() {
		if db != nil {
			if err := db.Close(); err != nil {
				log.Errorf("Failed to close db: %v", err)
			}
		}

		if rows != nil {
			if err := rows.Close(); err != nil {
				log.Errorf("Failed to close rows: %v", err)
			}
		}
	}
	return db, rows, cleanup, err
}
