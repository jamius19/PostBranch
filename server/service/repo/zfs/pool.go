package zfs

import (
	"context"
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/dto"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
)

var log = logger.Logger

func VirtualPool(ctx context.Context, repoinit *dto.RepoInit) (*dao.ZfsPool, error) {
	log.Infof("ZFS Pool init %v", *repoinit)

	_, path := util.SplitPath(repoinit.Path)
	log.Infof("Creating virtual disk file at %s", repoinit.Path)

	//
	// Set the Commands
	//
	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()
	cmds.Set("create-folder", cmd.Get("mkdir", "-p", path))

	filesize := fmt.Sprintf("%d%s", repoinit.Size, repoinit.SizeUnit)
	cmds.Set("create-file", cmd.Get("fallocate", "-l", filesize, repoinit.Path))

	loopNo, err := findFreeLoopDevice()
	if err != nil {
		log.Errorf("Failed to find free loop device: %s", err)
		return nil, err
	}

	log.Infof("Found free loop device: %d", loopNo)

	loopDevice := fmt.Sprintf("/dev/loop%d", loopNo)

	cmds.Set("create-loop-node", cmd.Get("mknod", loopDevice, "b", "7", fmt.Sprintf("%d", loopNo)))
	cmds.Set("setup-loopback", cmd.Get("losetup", loopDevice, repoinit.Path))

	//
	// Run the Commands
	//
	_, err = cmd.Multi(cmds)
	if err != nil {
		log.Errorf("Failed to create virtual disk: %s", err)
		return nil, err
	}

	pool, err := pool(ctx, repoinit, loopDevice)
	if err != nil {
		log.Errorf("Failed to create pool: %s", err)
		return nil, err
	}

	return pool, nil
}

func pool(ctx context.Context, repoinit *dto.RepoInit, path string) (*dao.ZfsPool, error) {
	//
	// Set the Commands
	//
	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()

	mountPath := fmt.Sprintf("/mnt/pb-%s", repoinit.Name)
	cmds.Set(
		"create-zpool",
		cmd.Get(
			"zpool", "create", "-m", mountPath, repoinit.Name, path,
		),
	)

	//
	// Run the Commands
	//
	_, err := cmd.Multi(cmds)

	// Create a new pool
	poolData := dao.CreatePoolParams{
		Name: repoinit.Name,
		Path: repoinit.Path,
	}

	pool, err := data.Fetcher.CreatePool(ctx, poolData)
	if err != nil {
		log.Errorf("Failed to insert pool. Repo:%v Path: %s Error:%s", poolData, path, err)
		return nil, err
	}

	log.Infof("Pool insertion successful %s", pool.Name)

	return &pool, nil
}
