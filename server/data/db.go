package data

import (
	"database/sql"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/service"
	_ "github.com/mattn/go-sqlite3"
)

const DbPath = "/var/lib/postbranch/postbranch.db"

var Db *sql.DB

func Initialize() *sql.DB {
	db, err := sql.Open("sqlite3", "/home/jamius19/.postbranch/main.db")
	if err != nil {
		logger.Logger.Fatal(err)
	}

	Db = db
	service.Initialize(db)

	logger.Logger.Info("Initialized database")
	return db
}