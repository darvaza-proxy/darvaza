// Package os provides some extensions over the standard os package
package os

import (
	"io/fs"
	"os"

	"darvaza.org/darvaza/shared/os/flock"
)

// ReadDirWithLock reads a directory using a syscall.Flock
func ReadDirWithLock(dirname string) ([]fs.DirEntry, error) {
	fl, err := flock.Lock(dirname)
	if err != nil {
		return nil, err
	}
	defer fl.Unlock()

	return os.ReadDir(dirname)
}

// ReadFileWithLock reads a file using a syscall.Flock
func ReadFileWithLock(filename string) ([]byte, error) {
	fl, err := flock.Lock(filename)
	if err != nil {
		return nil, err
	}
	defer fl.Unlock()

	return os.ReadFile(filename)
}
