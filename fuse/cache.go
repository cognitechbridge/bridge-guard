package fuse

import (
	"ctb-cli/filesyetem"
	"github.com/winfsp/cgofuse/fuse"
	"sync"
)

type Cache struct {
	sync.Mutex

	fs *filesyetem.FileSystem

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

func NewCache(fs *filesyetem.FileSystem) *Cache {
	c := Cache{}
	c.openMap = make(map[uint64]*Node)
	c.root = c.newNode(0, true)
	c.root.path = "/"
	c.fs = fs
	return &c
}

func (c *Cache) lookupNode(path string, ancestor *Node) (prnt *Node, name string, node *Node) {
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

func (c *Cache) getNode(path string, fh uint64) *Node {
	if ^uint64(0) == fh {
		_, _, node := c.lookupNode(path, nil)
		return node
	} else {
		return c.openMap[fh]
	}
}

func (c *Cache) openNode(path string, dir bool) (int, uint64) {
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

func (c *Cache) closeNode(fh uint64) int {
	node := c.openMap[fh]
	node.opencnt--
	if 0 == node.opencnt {
		delete(c.openMap, node.stat.Ino)
	}
	return 0
}

func (c *Cache) exploreDir(path string) {
	names := c.fs.GetSubFiles(path)
	_, _, parent := c.lookupNode(path, nil)
	for _, info := range names {
		_, _, node := c.lookupNode(info.Name, parent)
		if node == nil {
			node := c.newNode(0, info.IsDir)
			node.path = join(path, info.Name)
			node.stat.Size = info.Size
			parent.chld[info.Name] = node
		}
	}
	parent.explored = true
}

func (c *Cache) createFile(path string) int {
	prnt, name, node := c.lookupNode(path, nil)
	if nil == prnt {
		return -fuse.ENOENT
	}
	if nil != node {
		return -fuse.EEXIST
	}
	_ = c.fs.CreateFile(path)
	node = c.newNode(0, false)
	prnt.chld[name] = node
	return 0
}

func (c *Cache) createDir(path string) int {
	_ = c.fs.CreateDir(path)
	prnt, name, node := c.lookupNode(path, nil)
	if nil == prnt {
		return -fuse.ENOENT
	}
	if nil != node {
		return -fuse.EEXIST
	}
	node = c.newNode(0, true)
	prnt.chld[name] = node
	return 0
}

func (c *Cache) rmDir(path string) int {
	if err := c.removeNode(path, true); err != 0 {
		return err
	}
	c.fs.RemoveDir(path)
	return 0
}

func (c *Cache) removeNode(path string, dir bool) int {
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

func (c *Cache) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	node := c.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	n, _ = c.fs.Write(path, buff, ofst)
	return
}

func (c *Cache) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	node := c.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	n, _ = c.fs.Read(path, buff, ofst)
	return
}

func (c *Cache) newNode(dev uint64, isDir bool) *Node {
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
	}
	if isDir {
		self.chld = map[string]*Node{}
	}
	return &self
}

func (c *Cache) getMode(isDir bool) uint32 {
	if isDir {
		return fuse.S_IFDIR | 0777
	} else {
		return fuse.S_IFREG | 0777
	}
}

func (c *Cache) getUid() (uint32, uint32) {
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

func (c *Cache) getIno() uint64 {
	c.ino.Lock()
	defer c.ino.Unlock()
	c.ino.counter++
	return c.ino.counter
}

func (c *Cache) Truncate(path string, size int64, fh uint64) (errc int) {
	node := c.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	c.fs.Resize(path, size)
	node.stat.Size = size
	return 0
}

func (c *Cache) Rename(oldpath string, newpath string) int {
	oldprnt, oldname, oldnode := c.lookupNode(oldpath, nil)
	if nil == oldnode {
		return -fuse.ENOENT
	}
	newprnt, newname, newnode := c.lookupNode(newpath, nil)
	if nil == newprnt {
		return -fuse.ENOENT
	}
	if "" == newname {
		// guard against directory loop creation
		return -fuse.EINVAL
	}
	if oldprnt == newprnt && oldname == newname {
		return 0
	}
	if nil != newnode {
		return -fuse.ENOENT
	}
	err := c.fs.Rename(oldpath, newpath)
	if err != nil {
		return -fuse.ENOENT
	}
	delete(oldprnt.chld, oldname)
	newprnt.chld[newname] = oldnode
	return 0
}

func (c *Cache) RemoveFile(path string) int {
	err := c.fs.RemovePath(path)
	if err != nil {
		return -fuse.ENOENT
	}
	if err := c.removeNode(path, false); err != 0 {
		return err
	}
	return 0
}
