//go:build linux

package flock

import (
	"io/fs"
	"os"
	"syscall"
)

// Handle represets a OS specific file descriptor
type Handle int

const deadHandle Handle = -1

// revive:disable:flag-parameter

func openHandle(path string, create bool, perm fs.FileMode) (Handle, error) {
	// revive:enable:flag-parameter
	mode := os.O_RDONLY
	if create {
		mode |= os.O_CREATE
	}

	fd, err := syscall.Open(path, mode, uint32(perm))
	return Handle(fd), err
}

func closeHandle(h Handle) error {
	return syscall.Close(int(h))
}

func lockHandle(h Handle) error {
	return syscall.Flock(int(h), syscall.LOCK_EX)
}
