package opts

import (
	"flag"
	"github.com/jamius19/postbranch/logger"
)

type flags struct {
	configFile string
}

var log = logger.Logger

func loadFlags() *flags {
	log.Info("Loading config")

	configPath := flag.String("config-file", defaultConfigPath, "Path to config file")
	flag.Parse()

	args := flags{
		configFile: *configPath,
	}

	return &args
}
