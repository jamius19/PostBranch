package zfs

import (
	"github.com/jamius19/postbranch/internal/logger"
	"github.com/jamius19/postbranch/internal/runner"
	"strings"
)

const (
	FindLoopBackFromZpoolCmd = "zpool list -v %s | grep '^  loop' | awk '{print $1}'"
)

var zfsVersions = []string{
	"2.1.5",
}

func Version() (*string, bool) {
	log.Info("Checking ZFS availability")

	zfsOutput, err := runner.Single("zfs-version-check", true, false, "zfs", "--version")

	if err != nil || zfsOutput == runner.EmptyOutput {
		return nil, false
	}

	version := strings.Replace(zfsOutput, "\n", "\\\\", -1)

	logger.Logger.Infof("ZFS version: %s", version)
	for _, v := range zfsVersions {
		if strings.Contains(version, v) {
			return &v, true
		}
	}

	return nil, false
}
