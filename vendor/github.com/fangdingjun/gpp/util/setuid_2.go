// +build !cgo

package util

import (
	"fmt"
	"os"
)

//Setuid set the uid to uid
func Setuid(uid int) error {
	fmt.Fprintf(os.Stderr, "WARNING: setuid not supported\n")
	return nil
}

//Setgid set the gid to gid
func Setgid(gid int) error {
	fmt.Fprintf(os.Stderr, "WARNING: setgid not supported\n")
	return nil
}
