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

	cmd := exec.Command("repo", "--version")
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

func InitializeVirtual(req *dto.Virtual) error {
	log.Infof("repo data %v", *req)

	_, path := util.SplitPath(req.Path)
	filesize := fmt.Sprintf("%d%s", req.Size, req.SizeUnit[:1])
	log.Infof("Creating virtual disk file %s", req.Path)

	loopNo, err := findFreeLoopDevice()
	if err != nil {
		log.Errorf("Failed to find free loop device: %s", err)
		return err
	}

	log.Infof("Found free loop device: %d", loopNo)

	loopDevice := fmt.Sprintf("/dev/pb-loop-%d", loopNo)
	mountPath := fmt.Sprintf("/mnt/%s", req.Name)

	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()
	cmds.Set("create-folder", cmd.Get("mkdir", "-p", path))
	cmds.Set("create-file", cmd.Get("fallocate", "-l", filesize, req.Path))
	cmds.Set("create-loop-node", cmd.Get("mknod", loopDevice, "b", "7", string(rune(loopNo))))
	cmds.Set("setup-loopback", cmd.Get("losetup", loopDevice, req.Path))
	cmds.Set(
		"create-zpool",
		cmd.Get(
			"zpool", "create", "-o", fmt.Sprintf("mountpoint=%s", mountPath), req.Name, loopDevice,
		),
	)

	multi, err := cmd.Multi(cmds)
	if err != nil {
		log.Errorf("Failed to create virtual disk: %s", err)
		return err
	}

	for el := multi.Front(); el != nil; el = el.Next() {
		log.Infof("%s: %v", el.Key, el.Value.Output)
	}

	return nil
}
