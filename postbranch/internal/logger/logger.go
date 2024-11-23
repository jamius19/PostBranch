package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"runtime"
)

var Logger = &logrus.Logger{
	Out: os.Stderr,
	Formatter: &logrus.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return "", fmt.Sprintf("[%12s:%-3d]", filename, f.Line)
		},
	},
	Hooks:        make(logrus.LevelHooks),
	Level:        logrus.DebugLevel,
	ReportCaller: true,
}
