// Package mountpoint provides utilities for mountpoint validation.
package mountpoint

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

var (
	ErrNotADirectory = errors.New("not a directory")
	ErrStickyBitSet  = errors.New("directory has sticky bit set")
	ErrNotEmpty      = errors.New("directory not empty")
)

// validateStat checks if the mountpoint has the correct type and permissions. If this is the case, it returns nil,
// otherwise an error is returned specifying what exactly is wrong.
func validateStat(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}
	// check if mount point is a directory
	if !stat.IsDir() {
		return fmt.Errorf("%v: %w", path, ErrNotADirectory)
	}
	// check if we have write access
	err = syscall.Access(path, unix.W_OK)
	if err != nil {
		return err
	}
	// check for sticky bit
	if stat.Mode()&fs.ModeSticky != 0 {
		return fmt.Errorf("%v: %w", path, ErrStickyBitSet)
	}
	return nil
}

// validateNonEmpty checks if the specified directory is empty. If it is, it returns nil, otherwise - an error.
func validateNonEmpty(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	} else {
		return fmt.Errorf("%v: %w", path, ErrNotEmpty)
	}
}

// ValidateMountpoint validates if a given directory is a valid mountpoint. If allowNonEmpty is false, it additionally
// checks, if the given directory is empty. It returns the argument resolved to an absolute path.
// If the mountpoint is invalid, an error is returned specifying what is wrong.
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
