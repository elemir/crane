package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
)

func MountOverlay(lower string, upper string, work string) (string, error) {
	dir, err := ioutil.TempDir("", "merged.")
	if err != nil {
		return "", fmt.Errorf("cannot create temporary dir for merged fs: %s", err)
	}

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work)
	err = syscall.Mount("none", dir, "overlay", 0, options)
	if err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("cannot mount overlay2 with options '%s': %s", options, err)
	}

	return dir, nil
}

func UnmountOverlay(dir string) error {
	if err := syscall.Unmount(dir, 0); err != nil {
		return fmt.Errorf("cannot unmount merged dir '%s': %s", dir, err)
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("cannot remove mounted merged dir '%s': %s", dir, err)
	}

	return nil
}
