package zfs

import (
	"github.com/jamius19/postbranch/logger"
	"os/exec"
	"strings"
)

var zfsVersions = []string{
	"2.1.5",
}

func Version() (*string, bool) {
	log.Info("Checking ZFS availability")

	cmd := exec.Command("zfs", "--version")
	output, err := cmd.Output()

	if err != nil {
		return nil, false
	}

	version := strings.TrimSpace(string(output))
	version = strings.Replace(version, "\n", "/", 1)

	logger.Logger.Infof("ZFS version: %s", version)
	for _, v := range zfsVersions {
		if strings.Contains(version, v) {
			return &v, true
		}
	}

	return nil, false
}
