package data

import (
	"database/sql"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/logger"
	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "/var/lib/postbranch/postbranch.db"

var log = logger.Logger
var Db *sql.DB
var Fetcher *dao.Queries

func Initialize() *sql.DB {
	var err error

	Db, err = sql.Open("sqlite3", "/home/jamius19/.postbranch/main.db")
	if err != nil {
		logger.Logger.Fatal(err)
	}

	Fetcher = dao.New(Db)

	_, err = Db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		log.Errorf("Can't set PRAGMA foreign_keys: %s", err)
		return nil
	}

	log.Info("Initialized database")
	return Db
}
