package flock

import (
	"syscall"
)

// Flock implements a simple wrapper around syscall.Flock
type Flock struct {
	filename string
	opener   func(string) (int, error)
	fd       int
}

// New instantiates a Flock for a given filename
func New(filename string) *Flock {
	return &Flock{
		filename: filename,
		fd:       -1,
	}
}

// Lock flocks a file by name
func Lock(filename string) (*Flock, error) {
	fl := New(filename)
	if err := fl.Lock(); err != nil {
		return nil, err
	} else {
		return fl, nil
	}
}

func (lock *Flock) open() error {
	var fd int
	var err error

	if lock.fd >= 0 {
		// already open
		return syscall.EBUSY
	}

	if lock.opener != nil {
		// use given opener
		fd, err = lock.opener(lock.filename)
	} else {
		// default opener
		fd, err = syscall.Open(lock.filename, syscall.O_RDONLY, 0)
	}

	if err != nil {
		// failed to open
		return err
	} else {
		// openned
		lock.fd = fd
		return nil
	}
}

func (lock *Flock) close() {
	if fd := lock.fd; fd >= 0 {
		defer syscall.Close(fd)
		lock.fd = -1
	}
}

// Lock Flocks the file
func (lock *Flock) Lock() error {
	if err := lock.open(); err != nil {
		// failed to open
		return err
	} else if err := syscall.Flock(lock.fd, syscall.LOCK_EX); err != nil {
		// failed to clock
		defer lock.close()
		return err
	} else {
		// success
		return nil
	}
}

// Unlocks releases the flock
func (lock *Flock) Unlock() {
	lock.close()
}
