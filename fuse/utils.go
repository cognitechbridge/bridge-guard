package fuse

import (
	"fmt"
	"github.com/winfsp/cgofuse/examples/shared"
	"github.com/winfsp/cgofuse/fuse"
	"strings"
	"syscall"
)

func split(path string) []string {
	return strings.Split(path, "/")
}

func join(base string, path string) string {
	return strings.TrimRight(base, "/") + "/" + path
}

func errno(err error) int {
	if nil != err {
		return -int(err.(syscall.Errno))
	} else {
		return 0
	}
}

func trace(vals ...interface{}) func(vals ...interface{}) {
	uid, gid, _ := fuse.Getcontext()
	return shared.Trace(1, fmt.Sprintf("[uid=%v,gid=%v]", uid, gid), vals...)
}
