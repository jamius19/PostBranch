package zfs

import (
	"fmt"
	"github.com/jamius19/postbranch/internal/util"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func findFreeLoopNo() (int, error) {
	controlFd, err := os.OpenFile("/dev/loop-control", os.O_RDWR, 0)
	if err != nil {
		log.Fatal("Can't open /dev/loop-control for managing loop devices!")
	}
	defer controlFd.Close()

	// Use LOOP_CTL_GET_FREE to get a free loop device number
	loopNo, _, errno := unix.Syscall(unix.SYS_IOCTL, controlFd.Fd(), unix.LOOP_CTL_GET_FREE, 0)
	if errno != 0 {
		return -1, fmt.Errorf("ioctl LOOP_CTL_GET_FREE failed: %w", errno)
	}

	log.Printf("Found free loop device: %d", loopNo)

	return int(loopNo), nil
}

func CreateSparseFile(imgPath string, sizeInMb int64) error {
	_, path := util.SplitPath(imgPath)
	log.Infof("Creating virtual disk file at %s", imgPath)

	err := util.CreateDirectories(path, "root", 0600)
	if err != nil {
		return fmt.Errorf("failed to create directories for the sparse file. Error: %s", err)
	}

	// Open or create the file
	file, err := os.Create(imgPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	sizeInBytes := sizeInMb * 1024 * 1024

	if err := file.Truncate(sizeInBytes); err != nil {
		return fmt.Errorf("failed to set file size: %w", err)
	}

	if err := os.Chmod(imgPath, 0600); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	log.Infof("Virtual disk file successfully created at %s with size %d MB", imgPath, sizeInMb)
	return nil
}

func SetupLoopDevice(imgFilePath string) (int, error) {
	loopNo, err := findFreeLoopNo()
	if err != nil {
		log.Errorf("Failed to find free loop device. Error: %s", err)
		return -1, err
	}

	file, err := os.OpenFile(imgFilePath, os.O_RDWR, 0)
	if err != nil {
		log.Errorf("Failed to open img file: %v", err)
		return -1, err
	}
	defer file.Close()

	loopDevice := fmt.Sprintf("/dev/loop%d", loopNo)

	if _, err := os.Stat(loopDevice); os.IsNotExist(err) {
		var loopBackDeviceId = 7<<8 | loopNo

		if err := unix.Mknod(loopDevice, unix.S_IFBLK|0660, loopBackDeviceId); err != nil {
			return -1, fmt.Errorf("failed to create loop device %s: %w", loopDevice, err)
		}

		if err := os.Chown(loopDevice, 0, 6); err != nil {
			return -1, fmt.Errorf("failed to set ownership for %s: %w", loopDevice, err)
		}
	}

	loopFd, err := os.OpenFile(loopDevice, os.O_RDWR, 0)
	if err != nil {
		return -1, fmt.Errorf("failed to open loop device %s: %w", loopDevice, err)
	}
	defer loopFd.Close()

	// Set the file descriptor for the loop device
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, loopFd.Fd(), unix.LOOP_SET_FD, file.Fd()); errno != 0 {
		return -1, fmt.Errorf("ioctl LOOP_SET_FD failed: %w", errno)
	}

	log.Printf("Successfully attached file to %s", loopDevice)
	return loopNo, nil
}

func ReleaseLoopDevice(loopDevice string) error {
	loopFd, err := os.OpenFile(loopDevice, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open loop device %s: %w", loopDevice, err)
	}
	defer loopFd.Close()

	// Use the IOCTL command LOOP_CLR_FD to release the loop device
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, loopFd.Fd(), unix.LOOP_CLR_FD, 0); errno != 0 {
		return fmt.Errorf("ioctl LOOP_CLR_FD failed: %w", errno)
	}

	log.Infof("Successfully released loop device %s\n", loopDevice)
	return nil
}

func FindLoopDeviceFromSys(targetFile string) ([]string, error) {
	var loopbackDevices []string

	files, err := filepath.Glob("/sys/block/loop*/loop/backing_file")
	if err != nil {
		return loopbackDevices, err
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		if strings.TrimSpace(string(content)) == targetFile {
			// Extract loop device number from path
			parts := strings.Split(file, "/")
			if len(parts) >= 4 {
				devicePath := "/dev/" + parts[3]
				loopbackDevices = append(loopbackDevices, devicePath)
			}
		}
	}

	return loopbackDevices, nil
}
