package main

import (
	"github.com/jamius19/postbranch/db"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/opts"
	"github.com/jamius19/postbranch/web"
)

var log = logger.Logger

func main() {
	err := opts.Load()

	if err != nil {
		log.Fatal("Failed to load config")
	}

	log.Info("Config loaded")

	dbConn := db.Initialize()
	defer dbConn.Close()

	web.Initialize()
}
