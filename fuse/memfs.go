package fuse

import (
	"ctb-cli/filesyetem"
	"github.com/winfsp/cgofuse/fuse"
)

//type Memfs struct {
//	fuse.FileSystemBase
//	Cache *Cache
//}

func (c *Cache) Statfs(path string, stat *fuse.Statfs_t) (errc int) {
	stat.Frsize = 4096
	stat.Bsize = stat.Frsize
	stat.Blocks = uint64(2*1024*1024*1024) / stat.Frsize
	stat.Bfree = uint64(2*1024*1024*1024) / stat.Frsize
	stat.Bavail = stat.Bfree
	stat.Namemax = uint64(10 * 1024 * 1024)
	return 0
}

func (c *Cache) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Mode = (node.stat.Mode & fuse.S_IFMT) | mode&07777
	node.stat.Ctim = fuse.Now()
	return 0
}

func (c *Cache) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(path, uid, gid)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
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

func (c *Cache) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
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

func (c *Cache) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	defer c.synchronize()()
	return c.openNode(path, false)
}

func (c *Cache) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	defer c.synchronize()()
	node := c.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	*stat = node.stat
	return 0
}

func (c *Cache) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer c.synchronize()()
	return c.closeNode(fh)
}

func (c *Cache) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if node.explored == false {
		c.exploreDir(path)
	}
	return c.openNode(path, true)
}

func (c *Cache) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {

	defer trace(path, fill, ofst, fh)(&errc)
	defer c.synchronize()()
	node := c.openMap[fh]
	fill(".", &node.stat, 0)
	fill("..", nil, 0)
	for name, chld := range node.chld {
		if !fill(name, &chld.stat, 0) {
			break
		}
	}
	return 0
}

func (c *Cache) Releasedir(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer c.synchronize()()
	return c.closeNode(fh)
}

func (c *Cache) Setxattr(path string, name string, value []byte, flags int) (errc int) {
	defer trace(path, name, value, flags)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
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

func (c *Cache) Getxattr(path string, name string) (errc int, xatr []byte) {
	defer trace(path, name)(&errc, &xatr)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
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

func (c *Cache) Removexattr(path string, name string) (errc int) {
	defer trace(path, name)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
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

func (c *Cache) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(path, fill)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
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

func (c *Cache) Chflags(path string, flags uint32) (errc int) {
	defer trace(path, flags)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Flags = flags
	node.stat.Ctim = fuse.Now()
	return 0
}

func (c *Cache) Setcrtime(path string, tmsp fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Birthtim = tmsp
	node.stat.Ctim = fuse.Now()
	return 0
}

func (c *Cache) Setchgtime(path string, tmsp fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Ctim = tmsp
	return 0
}

func NewMemfs(fs *filesyetem.FileSystem) *Cache {
	self := NewCache(fs)
	defer self.synchronize()()
	return self
}

var _ fuse.FileSystemChflags = (*Cache)(nil)
var _ fuse.FileSystemSetcrtime = (*Cache)(nil)
var _ fuse.FileSystemSetchgtime = (*Cache)(nil)

func (c *Cache) Mount() {
	host := fuse.NewFileSystemHost(c)
	host.SetCapReaddirPlus(true)
	opts := make([]string, 0)
	host.Mount("", opts)
}
