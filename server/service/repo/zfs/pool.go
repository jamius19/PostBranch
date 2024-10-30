package zfs

import (
	"context"
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/cmd"
	"github.com/jamius19/postbranch/data"
	"github.com/jamius19/postbranch/data/dao"
	"github.com/jamius19/postbranch/data/dto/repo"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
)

var log = logger.Logger

func VirtualPool(ctx context.Context, repoinit *repo.InitDto) (*dao.ZfsPool, error) {
	log.Infof("ZFS Pool init %v", *repoinit)

	_, path := util.SplitPath(repoinit.Path)
	log.Infof("Creating virtual disk file at %s", repoinit.Path)

	//
	// Set the Commands
	//
	cmds := orderedmap.NewOrderedMap[string, cmd.Command]()
	cmds.Set("create-folder", cmd.Get("mkdir", "-p", path))

	filesize := fmt.Sprintf("%dM", repoinit.SizeInMb)
	cmds.Set("create-file", cmd.Get("fallocate", "-l", filesize, repoinit.Path))

	loopNo, err := findFreeLoopDevice()
	if err != nil {
		log.Errorf("Failed to find free loop device: %s", err)
		return nil, responseerror.Clarify("Failed to find free loop device")
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
		return nil, responseerror.Clarify("Error creating virtual disk")
	}

	pool, err := pool(ctx, repoinit, loopDevice)
	if err != nil {
		log.Errorf("Failed to create pool: %s", err)
		return nil, err
	}

	return pool, nil
}

func pool(ctx context.Context, repoinit *repo.InitDto, path string) (*dao.ZfsPool, error) {
	mountPath := fmt.Sprintf("/mnt/pb-%s", repoinit.Name)

	//
	// Run the Commands
	//
	_, err := cmd.Single("create-zpool", false, "zpool", "create", "-m", mountPath, repoinit.Name, path)
	if err != nil {
		log.Errorf("Failed to create pool: %s", err)
		return nil, err
	}

	// Create a new pool
	poolData := dao.CreatePoolParams{
		Name:      repoinit.Name,
		Path:      repoinit.Path,
		SizeInMb:  repoinit.SizeInMb,
		MountPath: mountPath,
	}

	pool, err := data.Fetcher.CreatePool(ctx, poolData)
	if err != nil {
		// TODO: Cleanup Pool
		log.Errorf("Failed to insert pool. Repo:%v Path: %s Error:%s", poolData, path, err)
		return nil, responseerror.Clarify("Failed to create pool")
	}

	log.Infof("Pool insertion successful %s", pool.Name)

	return &pool, nil
}
