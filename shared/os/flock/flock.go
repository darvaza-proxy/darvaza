// Package flock provides a wrapper around syscall.Flock
package flock

import (
	"os"
	"syscall"

	"darvaza.org/x/fs/fssyscall"
)

// Handle is a OS-agnostic syscall handler
type Handle = fssyscall.Handle

// OpenerFunc is a function that opens a file
type OpenerFunc func(string) (Handle, error)

// Flock implements a simple wrapper around syscall.Flock
type Flock struct {
	filename string
	opener   OpenerFunc
	h        Handle
}

// New instantiates a Flock for a given filename
func New(filename string) *Flock {
	return NewWithOpener(filename, nil)
}

// NewWithOpener instantiates a Flock for a given filename
func NewWithOpener(filename string, opener OpenerFunc) *Flock {
	if opener == nil {
		opener = defaultOpener
	}

	fl := &Flock{
		filename: filename,
		opener:   opener,
		h:        fssyscall.ZeroHandle,
	}
	return fl
}

func defaultOpener(path string) (Handle, error) {
	return fssyscall.Open(path, os.O_RDONLY, DefaultFileMode)
}

// Lock flocks a file by name
func Lock(filename string) (*Flock, error) {
	fl := New(filename)
	if err := fl.Lock(); err != nil {
		return nil, err
	}
	return fl, nil
}

func (lock *Flock) open() error {
	if !lock.h.IsZero() {
		// already open
		return syscall.EBUSY
	}

	h, err := lock.opener(lock.filename)
	if err != nil {
		// failed to open
		return err
	}

	lock.h = h
	return nil
}

func (lock *Flock) close() {
	if h := lock.h; !h.IsZero() {
		_ = h.Close()
		lock.h = fssyscall.ZeroHandle
	}
}

// Lock Flocks the file
func (lock *Flock) Lock() error {
	if err := lock.open(); err != nil {
		// failed to open
		return err
	} else if err := fssyscall.LockEx(lock.h); err != nil {
		// failed to flock
		defer lock.close()
		return err
	}

	return nil
}

// Unlock releases the flock
func (lock *Flock) Unlock() {
	lock.close()
}
