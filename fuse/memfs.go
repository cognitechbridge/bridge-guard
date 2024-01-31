package fuse

import (
	"ctb-cli/filesyetem"
	"github.com/winfsp/cgofuse/fuse"
)

type Memfs struct {
	fuse.FileSystemBase
	Cache *Cache
}

func (self *Memfs) Statfs(path string, stat *fuse.Statfs_t) (errc int) {
	stat.Frsize = 4096
	stat.Bsize = stat.Frsize
	stat.Blocks = uint64(2*1024*1024*1024) / stat.Frsize
	stat.Bfree = uint64(2*1024*1024*1024) / stat.Frsize
	stat.Bavail = stat.Bfree
	stat.Namemax = uint64(10 * 1024 * 1024)
	return 0
}

func (self *Memfs) Mknod(path string, mode uint32, dev uint64) (errc int) {
	return self.Cache.Mknod(path, mode, dev)
}

func (self *Memfs) Mkdir(path string, mode uint32) (errc int) {
	return self.Cache.Mkdir(path, mode)
}

func (self *Memfs) Unlink(path string) (errc int) {
	return self.Cache.Unlink(path)
}

func (self *Memfs) Rmdir(path string) (errc int) {
	return self.Cache.Rmdir(path)
}

// @Todo Implement links later
//func (self *Memfs) Link(oldpath string, newpath string) (errc int) {
//	defer trace(oldpath, newpath)(&errc)
//	defer self.synchronize()()
//	_, _, oldnode := self.lookupNode(oldpath, nil)
//	if nil == oldnode {
//		return -fuse.ENOENT
//	}
//	newprnt, newname, newnode := self.lookupNode(newpath, nil)
//	if nil == newprnt {
//		return -fuse.ENOENT
//	}
//	if nil != newnode {
//		return -fuse.EEXIST
//	}
//	oldnode.stat.Nlink++
//	newprnt.chld[newname] = oldnode
//	tmsp := fuse.Now()
//	oldnode.stat.Ctim = tmsp
//	newprnt.stat.Ctim = tmsp
//	newprnt.stat.Mtim = tmsp
//	return 0
//}
//
//
//func (self *Memfs) Symlink(target string, newpath string) (errc int) {
//	defer trace(target, newpath)(&errc)
//	defer self.synchronize()()
//	return self.makeNode(newpath, fuse.S_IFLNK|00777, 0, []byte(target))
//}
//
//func (self *Memfs) Readlink(path string) (errc int, target string) {
//	defer trace(path)(&errc, &target)
//	defer self.synchronize()()
//	_, _, node := self.lookupNode(path, nil)
//	if nil == node {
//		return -fuse.ENOENT, ""
//	}
//	if fuse.S_IFLNK != node.stat.Mode&fuse.S_IFMT {
//		return -fuse.EINVAL, ""
//	}
//	return 0, string(node.data)
//}

func (self *Memfs) Rename(oldpath string, newpath string) (errc int) {
	return self.Cache.Rename(oldpath, newpath)
}

func (self *Memfs) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Mode = (node.stat.Mode & fuse.S_IFMT) | mode&07777
	node.stat.Ctim = fuse.Now()
	return 0
}

func (self *Memfs) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(path, uid, gid)(&errc)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if ^uint32(0) != uid {
		node.stat.Uid = uid
	}
	if ^uint32(0) != gid {
		node.stat.Gid = gid
	}
	node.stat.Ctim = fuse.Now()
	return 0
}

func (self *Memfs) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Ctim = fuse.Now()
	if nil == tmsp {
		tmsp0 := node.stat.Ctim
		tmsa := [2]fuse.Timespec{tmsp0, tmsp0}
		tmsp = tmsa[:]
	}
	node.stat.Atim = tmsp[0]
	node.stat.Mtim = tmsp[1]
	return 0
}

func (self *Memfs) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	defer self.Cache.synchronize()()
	return self.Cache.openNode(path, false)
}

func (self *Memfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	defer self.Cache.synchronize()()
	node := self.Cache.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	*stat = node.stat
	return 0
}

func (self *Memfs) Truncate(path string, size int64, fh uint64) (errc int) {
	defer trace(path, size, fh)(&errc)
	defer self.Cache.synchronize()()
	return self.Cache.Truncate(path, size, fh)

}

func (self *Memfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer self.Cache.synchronize()()
	return self.Cache.Read(path, buff, ofst, fh)
}

func (self *Memfs) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	return self.Cache.Write(path, buff, ofst, fh)
}

func (self *Memfs) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer self.Cache.synchronize()()
	return self.Cache.closeNode(fh)
}

func (self *Memfs) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if node.explored == false {
		self.Cache.exploreDir(path)
	}
	return self.Cache.openNode(path, true)
}

func (self *Memfs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {

	defer trace(path, fill, ofst, fh)(&errc)
	defer self.Cache.synchronize()()
	node := self.Cache.openMap[fh]
	fill(".", &node.stat, 0)
	fill("..", nil, 0)
	for name, chld := range node.chld {
		if !fill(name, &chld.stat, 0) {
			break
		}
	}
	return 0
}

func (self *Memfs) Releasedir(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer self.Cache.synchronize()()
	return self.Cache.closeNode(fh)
}

func (self *Memfs) Setxattr(path string, name string, value []byte, flags int) (errc int) {
	defer trace(path, name, value, flags)(&errc)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if "com.apple.ResourceFork" == name {
		return -fuse.ENOTSUP
	}
	if fuse.XATTR_CREATE == flags {
		if _, ok := node.xatr[name]; ok {
			return -fuse.EEXIST
		}
	} else if fuse.XATTR_REPLACE == flags {
		if _, ok := node.xatr[name]; !ok {
			return -fuse.ENOATTR
		}
	}
	xatr := make([]byte, len(value))
	copy(xatr, value)
	if nil == node.xatr {
		node.xatr = map[string][]byte{}
	}
	node.xatr[name] = xatr
	return 0
}

func (self *Memfs) Getxattr(path string, name string) (errc int, xatr []byte) {
	defer trace(path, name)(&errc, &xatr)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT, nil
	}
	if "com.apple.ResourceFork" == name {
		return -fuse.ENOTSUP, nil
	}
	xatr, ok := node.xatr[name]
	if !ok {
		return -fuse.ENOATTR, nil
	}
	return 0, xatr
}

func (self *Memfs) Removexattr(path string, name string) (errc int) {
	defer trace(path, name)(&errc)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if "com.apple.ResourceFork" == name {
		return -fuse.ENOTSUP
	}
	if _, ok := node.xatr[name]; !ok {
		return -fuse.ENOATTR
	}
	delete(node.xatr, name)
	return 0
}

func (self *Memfs) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(path, fill)(&errc)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	for name := range node.xatr {
		if !fill(name) {
			return -fuse.ERANGE
		}
	}
	return 0
}

func (self *Memfs) Chflags(path string, flags uint32) (errc int) {
	defer trace(path, flags)(&errc)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Flags = flags
	node.stat.Ctim = fuse.Now()
	return 0
}

func (self *Memfs) Setcrtime(path string, tmsp fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Birthtim = tmsp
	node.stat.Ctim = fuse.Now()
	return 0
}

func (self *Memfs) Setchgtime(path string, tmsp fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer self.Cache.synchronize()()
	_, _, node := self.Cache.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Ctim = tmsp
	return 0
}

func NewMemfs(fs *filesyetem.FileSystem) *Memfs {
	self := Memfs{}
	self.Cache = NewCache(fs)
	defer self.Cache.synchronize()()
	return &self
}

var _ fuse.FileSystemChflags = (*Memfs)(nil)
var _ fuse.FileSystemSetcrtime = (*Memfs)(nil)
var _ fuse.FileSystemSetchgtime = (*Memfs)(nil)

func (self *Memfs) Mount() {
	host := fuse.NewFileSystemHost(self)
	host.SetCapReaddirPlus(true)
	opts := make([]string, 0)
	host.Mount("", opts)
}
