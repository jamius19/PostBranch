package util

import (
	"fmt"
	"github.com/jamius19/postbranch/cmd"
	"io"
	"os"
	"os/user"
	"path/filepath"
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

func SetPermissionsRecursive(path, user, group string) error {
	output, err := cmd.Single(
		"change-permissions-recursive",
		false,
		false,
		"su",
		"-c",
		fmt.Sprintf("chown -R %s:%s %s", user, group, path),
	)

	if err != nil {
		log.Errorf("Failed to change permissions recursively. output: %s data: %v", output, err)
		return err
	}

	return nil
}

func RemoveGlob(path string) (err error) {
	contents, err := filepath.Glob(path)
	if err != nil {
		return
	}

	for _, item := range contents {
		err = os.RemoveAll(item)
		if err != nil {
			return
		}
	}
	return
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
