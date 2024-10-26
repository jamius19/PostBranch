package zfs

import (
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/dto"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
	"os/exec"
	"strings"
)

var zfsVersions = []string{
	"2.1.5",
}

var log = logger.Logger

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

func InitializeVirtual(repoinit *dto.RepoInit) error {
	log.Infof("Repo init data %v", *repoinit)

	_, path := util.SplitPath(repoinit.Path)
	log.Infof("Creating virtual disk file at %s", repoinit.Path)

	loopNo, err := findFreeLoopDevice()
	if err != nil {
		log.Errorf("Failed to find free loop device: %s", err)
		return err
	}

	log.Infof("Found free loop device: %d", loopNo)

	loopDevice := fmt.Sprintf("/dev/loop%d", loopNo)
	mountPath := fmt.Sprintf("/mnt/pb-%s", repoinit.Name)

	// Run the Commands
	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()
	cmds.Set("create-folder", cmd.Get("mkdir", "-p", path))

	filesize := fmt.Sprintf("%d%s", repoinit.Size, repoinit.SizeUnit)
	cmds.Set("create-file", cmd.Get("fallocate", "-l", filesize, repoinit.Path))

	cmds.Set("create-loop-node", cmd.Get("mknod", loopDevice, "b", "7", fmt.Sprintf("%d", loopNo)))
	cmds.Set("setup-loopback", cmd.Get("losetup", loopDevice, repoinit.Path))

	//mountPathFlag := fmt.Sprintf("mountpoint=%s", mountPath)
	cmds.Set(
		"create-zpool",
		cmd.Get(
			"zpool", "create", "-m", mountPath, repoinit.Name, loopDevice,
		),
	)

	multi, err := cmd.Multi(cmds)
	if err != nil {
		log.Errorf("Failed to create virtual disk: %s", err)
		return err
	}

	for el := multi.Front(); el != nil; el = el.Next() {
		var output = "<nil>"
		if el.Value.Output != nil && *el.Value.Output != "" {
			output = *el.Value.Output
		}

		log.Infof("%s: %v", el.Key, output)
	}

	return nil
}
