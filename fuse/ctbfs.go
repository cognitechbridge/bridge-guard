package fuse

import (
	"ctb-cli/core"
	"fmt"
	"github.com/winfsp/cgofuse/fuse"
	"sync"
)

type CtbFs struct {
	fuse.FileSystemBase

	sync.Mutex

	fs core.FileSystemService

	root    *Node
	openMap map[uint64]*Node

	ino Ino
	uid uint32
	gid uint32
}

type Node struct {
	stat     fuse.Stat_t
	xatr     map[string][]byte
	chld     map[string]*Node
	opencnt  int
	explored bool
	path     string
}

type Ino struct {
	sync.Mutex
	counter uint64
}

func New(fs core.FileSystemService) *CtbFs {
	c := CtbFs{
		openMap: make(map[uint64]*Node),
		fs:      fs,
	}
	defer c.synchronize()()
	c.root = c.newNode(0, true, "/")
	return &c
}

func (c *CtbFs) Mount() {
	host := fuse.NewFileSystemHost(c)
	host.SetCapReaddirPlus(true)
	opts := make([]string, 0)
	host.Mount("", opts)
}

func (c *CtbFs) lookupNode(path string, ancestor *Node) (prnt *Node, name string, node *Node) {
	if ancestor == nil {
		prnt = c.root
		node = c.root
	} else {
		prnt = ancestor
		node = ancestor
	}
	name = ""
	for _, c := range split(path) {
		if "" != c {
			if 255 < len(c) {
				panic(fuse.Error(-fuse.ENAMETOOLONG))
			}
			prnt, name = node, c
			if node == nil {
				return
			}
			node = node.chld[c]
			if nil != ancestor && node == ancestor {
				name = "" // special case loop condition
				return
			}
		}
	}
	return
}

func (c *CtbFs) getNode(path string, fh uint64) *Node {
	if ^uint64(0) == fh {
		_, _, node := c.lookupNode(path, nil)
		return node
	} else {
		return c.openMap[fh]
	}
}

func (c *CtbFs) openNode(path string, dir bool) (errc int, fh uint64) {
	_, _, node := c.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT, ^uint64(0)
	}
	if !dir && fuse.S_IFDIR == node.stat.Mode&fuse.S_IFMT {
		return -fuse.EISDIR, ^uint64(0)
	}
	if dir && fuse.S_IFDIR != node.stat.Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR, ^uint64(0)
	}
	node.opencnt++
	if 1 == node.opencnt {
		c.openMap[node.stat.Ino] = node
	}
	return 0, node.stat.Ino
}

func (c *CtbFs) closeNode(fh uint64) int {
	node := c.openMap[fh]
	node.opencnt--
	if 0 == node.opencnt {
		c.commit(node)
		delete(c.openMap, node.stat.Ino)
	}
	return 0
}

func (c *CtbFs) exploreDir(path string) (err error) {
	names, err := c.fs.GetSubFiles(path)
	if err != nil {
		return fmt.Errorf("error exploring directory: %v", err)
	}
	_, _, parent := c.lookupNode(path, nil)
	for _, info := range names {
		_, _, node := c.lookupNode(info.Name(), parent)
		if node == nil {
			node := c.newNode(0, info.IsDir(), path)
			node.path = join(path, info.Name())
			node.stat.Size = info.Size()
			parent.chld[info.Name()] = node
		}
	}
	parent.explored = true
	return nil
}

func (c *CtbFs) Mknod(path string, mode uint32, dev uint64) (errc int) {
	defer trace(path, mode, dev)(&errc)
	defer c.synchronize()()
	prnt, name, node := c.lookupNode(path, nil)
	if nil == prnt {
		return -fuse.ENOENT
	}
	if nil != node {
		return -fuse.EEXIST
	}
	_ = c.fs.CreateFile(path)
	node = c.newNode(0, false, path)
	prnt.chld[name] = node
	return 0
}

func (c *CtbFs) Mkdir(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer c.synchronize()()
	prnt, name, node := c.lookupNode(path, nil)
	if nil == prnt {
		return -fuse.ENOENT
	}
	if nil != node {
		return -fuse.EEXIST
	}
	node = c.newNode(0, true, path)
	prnt.chld[name] = node
	_ = c.fs.CreateDir(path)
	return 0
}

func (c *CtbFs) Rmdir(path string) (errc int) {
	defer trace(path)(&errc)
	defer c.synchronize()()
	if err := c.removeNode(path, true); err != 0 {
		return err
	}
	if err := c.fs.RemoveDir(path); err != nil {
		return errno(err)
	}
	return 0
}

func (c *CtbFs) removeNode(path string, dir bool) int {
	prnt, name, node := c.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	if !dir && fuse.S_IFDIR == node.stat.Mode&fuse.S_IFMT {
		return -fuse.EISDIR
	}
	if dir && fuse.S_IFDIR != node.stat.Mode&fuse.S_IFMT {
		return -fuse.ENOTDIR
	}
	if 0 < len(node.chld) {
		return -fuse.ENOTEMPTY
	}
	node.stat.Nlink--
	delete(prnt.chld, name)
	return 0
}

func (c *CtbFs) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer c.synchronize()()
	node := c.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	n, _ = c.fs.Write(path, buff, ofst)
	return
}

func (c *CtbFs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer c.synchronize()()
	node := c.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	n, _ = c.fs.Read(path, buff, ofst)
	return
}

func (c *CtbFs) newNode(dev uint64, isDir bool, path string) *Node {
	uid, gid := c.getUid()
	tmsp := fuse.Now()
	ino := c.getIno()
	mode := c.getMode(isDir)
	self := Node{
		stat: fuse.Stat_t{
			Dev:      dev,
			Ino:      ino,
			Mode:     mode,
			Nlink:    1,
			Uid:      uid,
			Gid:      gid,
			Atim:     tmsp,
			Mtim:     tmsp,
			Ctim:     tmsp,
			Birthtim: tmsp,
			Flags:    0,
		},
		path: path,
	}
	if isDir {
		self.chld = map[string]*Node{}
	}
	return &self
}

func (c *CtbFs) getMode(isDir bool) uint32 {
	if isDir {
		return fuse.S_IFDIR | 0777
	} else {
		return fuse.S_IFREG | 0777
	}
}

func (c *CtbFs) getUid() (uint32, uint32) {
	uid, gid, _ := fuse.Getcontext()
	if uid != ^uint32(0) {
		if c.root != nil {
			c.root.stat.Uid = uid
			c.root.stat.Gid = gid
		}
		c.uid = uid
		c.gid = gid
	}
	return c.uid, c.gid
}

func (c *CtbFs) getIno() uint64 {
	c.ino.Lock()
	defer c.ino.Unlock()
	c.ino.counter++
	return c.ino.counter
}

func (c *CtbFs) Truncate(path string, size int64, fh uint64) (errc int) {
	defer trace(path, size, fh)(&errc)
	defer c.synchronize()()
	node := c.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	if err := c.fs.Resize(path, size); err != nil {
		return errno(err)
	}
	node.stat.Size = size
	return 0
}

func (c *CtbFs) Rename(oldPath string, newPath string) (errc int) {
	defer trace(oldPath, newPath)(&errc)
	defer c.synchronize()()
	oldPrnt, oldName, oldNode := c.lookupNode(oldPath, nil)
	if nil == oldNode {
		return -fuse.ENOENT
	}
	newPrnt, newName, newNode := c.lookupNode(newPath, nil)
	if nil == newPrnt {
		return -fuse.ENOENT
	}
	if "" == newName {
		// guard against directory loop creation
		return -fuse.EINVAL
	}
	if oldPrnt == newPrnt && oldName == newName {
		return 0
	}
	if nil != newNode {
		return -fuse.ENOENT
	}
	err := c.fs.Rename(oldPath, newPath)
	if err != nil {
		return -fuse.ENOENT
	}
	delete(oldPrnt.chld, oldName)
	newPrnt.chld[newName] = oldNode
	return 0
}

func (c *CtbFs) Unlink(path string) (errc int) {
	defer trace(path)(&errc)
	defer c.synchronize()()
	err := c.fs.RemovePath(path)
	if err != nil {
		return -fuse.ENOENT
	}
	if err := c.removeNode(path, false); err != 0 {
		return err
	}
	return 0
}

func (c *CtbFs) Statfs(_ string, stat *fuse.Statfs_t) (errc int) {
	stat.Frsize = 4096
	stat.Bsize = stat.Frsize
	stat.Blocks = uint64(2*1024*1024*1024) / stat.Frsize
	stat.Bfree = uint64(2*1024*1024*1024) / stat.Frsize
	stat.Bavail = stat.Bfree
	stat.Namemax = uint64(10 * 1024 * 1024)
	return 0
}

func (c *CtbFs) Chmod(path string, mode uint32) (errc int) {
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

func (c *CtbFs) Chown(path string, uid uint32, gid uint32) (errc int) {
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

func (c *CtbFs) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
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

func (c *CtbFs) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	defer c.synchronize()()
	return c.openNode(path, false)
}

func (c *CtbFs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	defer c.synchronize()()
	node := c.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	*stat = node.stat
	return 0
}

func (c *CtbFs) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer c.synchronize()()
	return c.closeNode(fh)
}

func (c *CtbFs) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if node.explored == false {
		err := c.exploreDir(path)
		if err != nil {
			return errno(err), ^uint64(0)
		}
	}
	return c.openNode(path, true)
}

func (c *CtbFs) Readdir(path string,
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

func (c *CtbFs) Releasedir(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer c.synchronize()()
	return c.closeNode(fh)
}

func (c *CtbFs) Setxattr(path string, name string, value []byte, flags int) (errc int) {
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

func (c *CtbFs) Getxattr(path string, name string) (errc int, xatr []byte) {
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

func (c *CtbFs) Removexattr(path string, name string) (errc int) {
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

func (c *CtbFs) Listxattr(path string, fill func(name string) bool) (errc int) {
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

func (c *CtbFs) Chflags(path string, flags uint32) (errc int) {
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

func (c *CtbFs) Setcrtime(path string, tmsp fuse.Timespec) (errc int) {
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

func (c *CtbFs) Setchgtime(path string, tmsp fuse.Timespec) (errc int) {
	defer trace(path, tmsp)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if nil == node {
		return -fuse.ENOENT
	}
	node.stat.Ctim = tmsp
	return 0
}

func (c *CtbFs) synchronize() func() {
	c.Lock()
	return func() {
		c.Unlock()
	}
}

func (c *CtbFs) commit(node *Node) error {
	return c.fs.Commit(node.path)
}
