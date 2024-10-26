package opts

import (
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
	"os"
)

type Opts struct {
	Server struct {
		Port int `yaml:"port" validate:"required,min=1,max=65535"`
	} `yaml:"server"`
}

const defaultConfigPath = "/etc/postbranch/config.yml"

var Config *Opts

func Load() error {
	flags := loadFlags()

	if _, err := os.Stat(flags.configFile); os.IsNotExist(err) {
		log.Errorf("Config file is not found on %s", flags.configFile)
		log.Infof(
			"Usage: postbranch --config-file <config_file> or the default config file at %s",
			defaultConfigPath,
		)

		return err
	}

	buf, err := os.ReadFile(flags.configFile)
	if err != nil {
		return err
	}

	config := &Opts{}

	err = yaml.Unmarshal(buf, config)

	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		return err
	}

	Config = config

	return nil
}
