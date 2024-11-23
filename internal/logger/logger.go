package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

const logPathLen = 45

var Logger = &logrus.Logger{
	Out: os.Stderr,
	Formatter: &logrus.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := extractLogPath(f.File, "PostBranch/")
			return "", fmt.Sprintf("[ %45s:%s ]", filename, getPaddedLineNumber(f.Line))
		},
	},
	Hooks:        make(logrus.LevelHooks),
	Level:        logrus.DebugLevel,
	ReportCaller: true,
}

func getPaddedLineNumber(linenum int) string {
	if linenum < 10 {
		return fmt.Sprintf("00%d", linenum)
	} else if linenum < 100 {
		return fmt.Sprintf("0%d", linenum)
	} else {
		return fmt.Sprintf("%d", linenum)
	}
}

func extractLogPath(fullPath, prefix string) string {
	startIndex := strings.Index(fullPath, prefix)
	if startIndex == -1 {
		return ""
	}

	path := fullPath[startIndex+len(prefix):]

	if len(path) > logPathLen {
		path = "..." + path[len(path)-(logPathLen-3):]
	}

	return path
}
