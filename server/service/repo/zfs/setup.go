package zfs

import (
	"fmt"
	"os"
)

func findFreeLoopDevice() (int, error) {
	const maxMinor = 255

	//Check existing /dev/loop* devices
	for i := 50; i <= maxMinor; i++ {
		loopDevice := fmt.Sprintf("/dev/loop%d", i)
		if _, err := os.Stat(loopDevice); err != nil {
			log.Infof("Found free loop device: %d", i)

			return i, nil
		}
	}

	return -1, fmt.Errorf("no free loop device minor number available in range 0-255")
}
