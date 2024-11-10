package zfs

import (
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
	"strings"
)

const (
	FindLoopBackCmd          = "losetup -a | grep '%s' | cut -d ':' -f 1"
	FindLoopBackFromZpoolCmd = "zpool list -v %s | grep '^  loop' | awk '{print $1}'"
)

var zfsVersions = []string{
	"2.1.5",
}

func Version() (*string, bool) {
	log.Info("Checking ZFS availability")

	zfsOutput, err := cmd.Single("zfs-version-check", true, false, "zfs", "--version")
	zfsVersion := util.TrimmedString(zfsOutput)

	if err != nil || zfsVersion == cmd.EmptyOutput {
		return nil, false
	}

	version := strings.Replace(zfsVersion, "\n", "/", -1)

	logger.Logger.Infof("ZFS version: %s", version)
	for _, v := range zfsVersions {
		if strings.Contains(version, v) {
			return &v, true
		}
	}

	return nil, false
}
