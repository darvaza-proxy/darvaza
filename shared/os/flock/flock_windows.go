//go:build windows

package flock

import (
	"io/fs"
	"os"
	"syscall"
)

// Handle represets a OS specific file descriptor
type Handle syscall.Handle

const deadHandle Handle = 0

// revive:disable:flag-parameter

func openHandle(path string, create bool, perm fs.FileMode) (Handle, error) {
	// revive:enable:flag-parameter
	mode := os.O_RDONLY
	if create {
		mode |= os.O_CREATE
	}

	h, err := syscall.Open(path, mode, uint32(perm))
	return Handle(h), err
}

func closeHandle(h Handle) error {
	return syscall.Close(syscall.Handle(h))
}

func lockHandle(_ Handle) error {
	return nil
}
