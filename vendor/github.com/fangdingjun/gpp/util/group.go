// +build linux,cgo darwin,cgo

package util

/*
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <grp.h>

int get_gid_by_name(char *name){
	struct group *grp;
	grp = getgrnam(name);
	if (grp != NULL){
		return grp->gr_gid;
	}else{
		return -1;
	}
}

char * get_name_by_gid(int gid){
	struct group *grp;
	grp = getgrgid(gid);
	if (grp == NULL){
		return NULL;
	}
	return grp->gr_name;
}

*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Group is group struct
type Group struct {
	Gid  int
	Name string
}

// LookupGroupID return a Group by the group id
func LookupGroupID(gid int) (*Group, error) {
	name, err := C.get_name_by_gid(C.int(gid))
	if err != nil {
		return nil, err
	}
	n := C.GoString(name)
	if n == "" {
		return nil, fmt.Errorf("gid %d does not exists", gid)
	}
	return &Group{gid, n}, nil
}

// LookupGroup return a Group by the group name
func LookupGroup(name string) (*Group, error) {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	gid, err := C.get_gid_by_name(n)
	if int(gid) == -1 || err != nil {
		if err == nil {
			return nil, fmt.Errorf("group %s does not exists", name)
		}
		return nil, err
	}

	return &Group{int(gid), name}, nil
}
