// +build linux,cgo darwin,cgo

package util

/*
#include <sys/types.h>
#include <unistd.h>
*/
import "C"

//Setuid set the uid to uid
func Setuid(uid int) error {
	ret, err := C.setuid(C.__uid_t(uid))
	if ret == C.int(0) {
		return nil
	}

	return err
}

//Setgid set the gid to gid
func Setgid(gid int) error {
	ret, err := C.setgid(C.__gid_t(gid))
	if ret == C.int(0) {
		return nil
	}

	return err
}
