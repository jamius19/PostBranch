package data

import (
	"database/sql"
	"github.com/jamius19/postbranch/data/fetch"
	"github.com/jamius19/postbranch/logger"
	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "/var/lib/postbranch/postbranch.db"

var Db *sql.DB
var Fetcher *fetch.Queries

func Initialize() *sql.DB {
	var err error

	Db, err = sql.Open("sqlite3", "/home/jamius19/.postbranch/main.db")
	if err != nil {
		logger.Logger.Fatal(err)
	}

	Fetcher = fetch.New(Db)

	logger.Logger.Info("Initialized database")
	return Db
}
