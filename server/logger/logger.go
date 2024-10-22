package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Logger = &logrus.Logger{
	Out: os.Stderr,
	Formatter: &logrus.TextFormatter{
		FullTimestamp: true,
	},
	Hooks: make(logrus.LevelHooks),
	Level: logrus.DebugLevel,
}
