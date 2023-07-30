package mountpoint

import (
	"fmt"
	"golang.org/x/sys/unix"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

func validateStat(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	// check if mount point is a directory
	if !stat.IsDir() {
		return fmt.Errorf("not a directory: %v", path)
	}
	// check if we have write access
	err = syscall.Access(path, unix.W_OK)
	if err != nil {
		return err
	}
	// check for sticky bit
	if stat.Mode()&fs.ModeSticky != 0 {
		return fmt.Errorf("directory has sticky bit set: %v", path)
	}
	return nil
}

func validateNonEmpty(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	} else {
		return fmt.Errorf("directory not empty: %v", path)
	}
}

func ValidateMountpoint(path string, allowNonEmpty bool) (absPath string, err error) {
	absPath, err = filepath.Abs(path)
	if err != nil {
		return
	}

	err = validateStat(absPath)
	if err != nil {
		return
	}

	if !allowNonEmpty {
		err = validateNonEmpty(absPath)
		if err != nil {
			return
		}
	}

	return
}
