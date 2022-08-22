package os

import (
	"os"
)

type (
	// FileInfo is an alias of the standard os.FileInfo
	FileInfo = os.FileInfo
	// FileMode is an alias of the standard os.FileMode
	FileMode = os.FileMode
)

// IsNotExist is a proxy to the standard os.IsNotExist
func IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// Remove is a proxy to the standard os.Remove
func Remove(name string) error {
	return os.Remove(name)
}

// Stat is a proxy to the standard os.Stat
func Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
