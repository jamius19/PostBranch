package db

import (
	"database/sql"
	"github.com/jamius19/postbranch/logger"
	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "/var/lib/postbranch/postbranch.db"

var log = logger.Logger
var Db *sql.DB

func Initialize() func() {
	var err error

	Db, err = sql.Open("sqlite3", "/home/jamius19/.postbranch/main.db")
	if err != nil {
		logger.Logger.Fatal(err)
	}

	_, err = Db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		log.Errorf("Can't set PRAGMA foreign_keys: %s", err)
		return nil
	}

	log.Info("Initialized database")
	return func() {
		err := Db.Close()
		if err != nil {
			log.Errorf("Can't close database: %s", err)
			return
		}
	}
}
