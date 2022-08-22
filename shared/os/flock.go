package os

import (
	"io/fs"
	"os"

	"github.com/darvaza-proxy/darvaza/shared/os/flock"
)

func ReadDirWithLock(dirname string) ([]fs.DirEntry, error) {
	if fl, err := flock.Lock(dirname); err != nil {
		return nil, err
	} else {
		defer fl.Unlock()
	}

	return os.ReadDir(dirname)
}

func ReadFileWithLock(filename string) ([]byte, error) {
	if fl, err := flock.Lock(filename); err != nil {
		return nil, err
	} else {
		defer fl.Unlock()
	}

	return os.ReadFile(filename)
}
