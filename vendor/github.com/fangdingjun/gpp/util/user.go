package util

import (
	//"log"
	"os/user"
	"strconv"
	"syscall"
)

// DropPrivilege drop privilege to username and group
func DropPrivilege(username, group string) error {

	uid := syscall.Getuid()

	if uid != 0 {
		// only root(uid=0) can call setuid
		// not root, skip
		return nil
	}

	// go1.7 will add user.LookupGroup
	// now use ourself LookupGroup
	if group != "" {
		g, err := LookupGroup(group)
		if err == nil {
			err := Setgid(g.Gid)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if username != "" {
		u, err := user.Lookup(username)
		if err != nil {
			return err
		}
		uid, _ := strconv.Atoi(u.Uid)
		err = Setuid(uid)
		if err != nil {
			return err
		}

	}
	return nil
}
