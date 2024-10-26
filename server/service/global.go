package service

import (
	"database/sql"
	"github.com/jamius19/postbranch/data/fetch"
	"github.com/jamius19/postbranch/logger"
)

var log = logger.Logger
var f *fetch.Queries

func Initialize(db *sql.DB) {
	f = fetch.New(db)
}
