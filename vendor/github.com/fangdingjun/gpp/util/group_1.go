// +build !cgo

package util

import (
	"fmt"
)

// Group is group struct
type Group struct {
	Gid  int
	Name string
}

// LookupGroupID return a Group by the group id
func LookupGroupID(gid int) (*Group, error) {
	return nil, fmt.Errorf("lookup group by id not implemented")
}

// LookupGroup return a Group by the group name
func LookupGroup(name string) (*Group, error) {
	return nil, fmt.Errorf("lookup group not implemented")
}
