package db

import (
	"database/sql"
	"github.com/jamius19/postbranch/logger"
	_ "github.com/mattn/go-sqlite3"
)

const DbPath = "/var/lib/postbranch/postbranch.db"

func Initialize() *sql.DB {
	db, err := sql.Open("sqlite3", "/home/jamius19/.postbranch/main.db")
	if err != nil {
		logger.Logger.Fatal(err)
	}

	logger.Logger.Info("Initialized database")
	return db
}
