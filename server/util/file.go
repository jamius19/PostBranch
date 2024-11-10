package util

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
)

func CopyFile(src, dst, user string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		log.Errorf("Failed to open source file: %v", err)
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		log.Errorf("Failed to create destination file: %v", err)
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		log.Errorf("Failed to copy file: %v", err)
		return err
	}

	return SetPermissions(dst, user)
}

func CreateDirectories(path, user string, perm os.FileMode) error {
	err := os.MkdirAll(path, perm)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	if err := SetPermissions(path, user); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

func SetPermissions(path, username string) error {
	osUser, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("failed to lookup user: %s, error: %v", username, err)
	}

	uid, _ := strconv.Atoi(osUser.Uid)
	gid, _ := strconv.Atoi(osUser.Gid)

	if err := os.Chown(path, uid, gid); err != nil {
		return fmt.Errorf("failed to change ownership of path: %s, error: %v", path, err)
	}

	return nil
}

func RemoveFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to remove file: %s, error: %v", path, err)
	}

	return nil
}
