package fuse

import (
	"ctb-cli/core"
	"fmt"
	"os"
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/winfsp/cgofuse/fuse"
)

type CtbFs struct {
	mountPoint string
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
	modePerm := fs.GetUserFileAccess("/", true)
	c.root = c.newNode(0, true, "/", uint32(modePerm))
	return &c
}

func (c *CtbFs) FindMountPoint() string {
	mount := ""
	if runtime.GOOS == "windows" {
		mount = c.FindUnusedDrive()
	} else if runtime.GOOS == "darwin" {
		mount = "/Volumes/ctbfs"
	} else if runtime.GOOS == "linux" {
		mount = "/mnt/ctbfs"
	}
	c.mountPoint = mount
	return c.mountPoint
}

func (c *CtbFs) Mount() {
	host := fuse.NewFileSystemHost(c)
	host.SetCapReaddirPlus(true)
	opts := make([]string, 0)
	mount := c.mountPoint
	if runtime.GOOS == "windows" {
		opts = append(opts, "-o", "volname=CTB-Secure-Drive")
	}
	host.Mount(mount, opts)
}

// FindUnusedDrive finds the first unused drive letter in the system.
// It iterates through drive letters from 'Z' to 'A' and checks if each drive is accessible.
// If an unused drive is found, it prints the drive letter and exits the loop.
func (c *CtbFs) FindUnusedDrive() string {
	for drive := 'Z'; drive >= 'A'; drive-- {
		_, err := os.Open(string(drive) + ":\\")
		if err != nil {
			return string(drive) + ":"
		}
	}
	return ""
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
		if c != "" {
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
	if node == nil {
		log.Error("Error opening node: ", path, " does not exist.")
		return -fuse.ENOENT, ^uint64(0)
	}
	if !dir && fuse.S_IFDIR == node.stat.Mode&fuse.S_IFMT {
		log.Error("Error opening node: ", path, " is a directory and requested as a file.")
		return -fuse.EISDIR, ^uint64(0)
	}
	if dir && fuse.S_IFDIR != node.stat.Mode&fuse.S_IFMT {
		log.Error("Error opening node: ", path, " is not a directory and requested as a directory.")
		return -fuse.ENOTDIR, ^uint64(0)
	}
	node.opencnt++
	if node.opencnt == 1 {
		c.openMap[node.stat.Ino] = node
	}
	return 0, node.stat.Ino
}

func (c *CtbFs) closeNode(fh uint64) int {
	node := c.openMap[fh]
	node.opencnt--
	if node.opencnt == 0 {
		err := c.commit(node)
		if err != nil {
			return errno(err)
		}
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
			nodePath := join(path, info.Name())
			node := c.newNode(0, info.IsDir(), nodePath, uint32(info.Mode()))
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
	if prnt == nil {
		log.Error("Error creating node: ", path, ". Parent does not exist.")
		return -fuse.ENOENT
	}
	if node != nil {
		log.Error("Error creating node: ", path, ". Node already exists.")
		return -fuse.EEXIST
	}
	_ = c.fs.CreateFile(path)
	node = c.newNode(0, false, path, 0777)
	prnt.chld[name] = node
	return 0
}

func (c *CtbFs) Mkdir(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer c.synchronize()()
	prnt, name, node := c.lookupNode(path, nil)
	if prnt == nil {
		log.Error("Error creating directory: ", path, ". Parent does not exist.")
		return -fuse.ENOENT
	}
	if node != nil {
		log.Error("Error creating directory: ", path, ". Directory already exists.")
		return -fuse.EEXIST
	}
	node = c.newNode(0, true, path, 0777)
	prnt.chld[name] = node
	_ = c.fs.CreateDir(path)
	return 0
}

func (c *CtbFs) Rmdir(path string) (errc int) {
	defer trace(path)(&errc)
	defer c.synchronize()()
	if err := c.removeNode(path, true); err != 0 {
		log.Error("Error removing node while removing directory: ", path, ". error: ", err)
		return err
	}
	if err := c.fs.RemoveDir(path); err != nil {
		log.Error("Error removing directory: ", path, ". error: ", err)
		return errno(err)
	}
	return 0
}

func (c *CtbFs) removeNode(path string, dir bool) int {
	prnt, name, node := c.lookupNode(path, nil)
	if node == nil {
		log.Error("Error removing node: ", path, ". Node does not exist.")
		return -fuse.ENOENT
	}
	if !dir && fuse.S_IFDIR == node.stat.Mode&fuse.S_IFMT {
		log.Error("Error removing node: ", path, ". Node is a directory and requested as a file.")
		return -fuse.EISDIR
	}
	if dir && fuse.S_IFDIR != node.stat.Mode&fuse.S_IFMT {
		log.Error("Error removing node: ", path, ". Node is not a directory and requested as a directory.")
		return -fuse.ENOTDIR
	}
	if 0 < len(node.chld) {
		log.Error("Error removing node: ", path, ". Directory is not empty.")
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
	if node == nil {
		log.Error("Error writing to node: ", path, ". Node does not exist.")
		return -fuse.ENOENT
	}
	n, _ = c.fs.Write(path, buff, ofst)
	return
}

func (c *CtbFs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer c.synchronize()()
	node := c.getNode(path, fh)
	if node == nil {
		log.Error("Error reading from node: ", path, ". Node does not exist.")
		return -fuse.ENOENT
	}
	n, _ = c.fs.Read(path, buff, ofst)
	return
}

func (c *CtbFs) newNode(dev uint64, isDir bool, path string, modePerm uint32) *Node {
	uid := uint32(0)
	gid := uint32(0)
	if path != "/" {
		uid, gid = c.getUid()
	}
	tmsp := fuse.Now()
	ino := c.getIno()
	mode := c.getMode(isDir, modePerm)
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

func (c *CtbFs) getMode(isDir bool, modePerm uint32) uint32 {
	if isDir {
		return fuse.S_IFDIR | modePerm
	} else {
		return fuse.S_IFREG | modePerm
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
	if node == nil {
		log.Error("Error truncating node: ", path, ". Node does not exist.")
		return -fuse.ENOENT
	}
	if err := c.fs.Resize(path, size); err != nil {
		log.Error("Error resizing file while truncating node: ", path, ". error: ", err)
		return errno(err)
	}
	node.stat.Size = size
	return 0
}

func (c *CtbFs) Rename(oldPath string, newPath string) (errc int) {
	defer trace(oldPath, newPath)(&errc)
	defer c.synchronize()()
	oldPrnt, oldName, oldNode := c.lookupNode(oldPath, nil)
	if oldNode == nil {
		log.Error("Error renaming node: ", oldPath, ". Node does not exist.")
		return -fuse.ENOENT
	}
	newPrnt, newName, newNode := c.lookupNode(newPath, nil)
	if newPrnt == nil {
		log.Error("Error renaming node: ", newPath, ". New parent does not exist.")
		return -fuse.ENOENT
	}
	if newName == "" {
		log.Error("Error renaming node: ", newPath, ". New name is empty. (directory loop)")
		// guard against directory loop creation
		return -fuse.EINVAL
	}
	if oldPrnt == newPrnt && oldName == newName {
		log.Warn("Renaming node: ", oldPath, " to ", newPath, ". No change.")
		return 0
	}
	if newNode != nil {
		log.Error("Error renaming node: ", newPath, ". Node already exists.")
		return -fuse.ENOENT
	}
	err := c.fs.Rename(oldPath, newPath)
	if err != nil {
		log.Error("Error renaming node: ", oldPath, " to ", newPath, ". error: ", err)
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
		log.Error("Error removing (unlink) node: ", path, ". error: ", err)
		return -fuse.ENOENT
	}
	if err := c.removeNode(path, false); err != 0 {
		return err
	}
	return 0
}

// Statfs returns file system statistics.
// It populates the provided `stat` structure with information about the file system.
// The `stat` structure contains fields such as block size, total blocks, free blocks, and maximum filename length.
// If an error occurs while retrieving disk usage information, it returns -fuse.ENOENT.
func (c *CtbFs) Statfs(_ string, stat *fuse.Statfs_t) (errc int) {
	reserved := uint64(1 * 1024 * 1024 * 1024) // 1GB
	stat.Frsize = 4096
	stat.Bsize = stat.Frsize
	totalBytes, freeBytesAvailable, err := c.fs.GetDiskUsage()
	if err != nil {
		log.Error("Error getting disk usage: ", err)
		return -fuse.ENOENT
	}
	stat.Blocks = totalBytes / stat.Frsize
	if freeBytesAvailable < reserved {
		stat.Bfree = 0
	} else {
		stat.Bfree = (freeBytesAvailable - 1*1024*1024*1024) / stat.Frsize
	}
	stat.Bavail = stat.Bfree
	stat.Namemax = uint64(10 * 1024 * 1024)
	return 0
}

func (c *CtbFs) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if node == nil {
		log.Error("Error changing mode of node: ", path, ". Node does not exist.")
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
	if node == nil {
		log.Error("Error changing ownership of node: ", path, ". Node does not exist.")
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
	if node == nil {
		log.Error("Error setting time of node: ", path, ". Node does not exist.")
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
	if node == nil {
		log.Error("Error getting attributes of node: ", path, ". Node does not exist.")
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
	if !node.explored {
		err := c.exploreDir(path)
		if err != nil {
			log.Error("Error opening directory: ", path, ". error: ", err)
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
	if node == nil {
		return -fuse.ENOENT
	}
	if name == "com.apple.ResourceFork" {
		return -fuse.ENOTSUP
	}
	if fuse.XATTR_CREATE == flags {
		if _, ok := node.xatr[name]; ok {
			log.Error("Error setting extended attribute: ", path, ". Extended attribute already exists.")
			return -fuse.EEXIST
		}
	} else if fuse.XATTR_REPLACE == flags {
		if _, ok := node.xatr[name]; !ok {
			log.Error("Error setting extended attribute: ", path, ". Extended attribute does not exist.")
			return -fuse.ENOATTR
		}
	}
	xatr := make([]byte, len(value))
	copy(xatr, value)
	if node.xatr == nil {
		node.xatr = map[string][]byte{}
	}
	node.xatr[name] = xatr
	return 0
}

func (c *CtbFs) Getxattr(path string, name string) (errc int, xatr []byte) {
	defer trace(path, name)(&errc, &xatr)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if node == nil {
		log.Error("Error getting extended attribute: ", path, ". Node does not exist.")
		return -fuse.ENOENT, nil
	}
	if name == "com.apple.ResourceFork" {
		log.Error("Error getting extended attribute: ", path, ". Resource fork is not supported.")
		return -fuse.ENOTSUP, nil
	}
	xatr, ok := node.xatr[name]
	if !ok {
		log.Error("Error getting extended attribute: ", path, ". Extended attribute does not exist.")
		return -fuse.ENOATTR, nil
	}
	return 0, xatr
}

func (c *CtbFs) Removexattr(path string, name string) (errc int) {
	defer trace(path, name)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if node == nil {
		log.Error("Error removing extended attribute: ", path, ". Node does not exist.")
		return -fuse.ENOENT
	}
	if name == "com.apple.ResourceFork" {
		log.Error("Error removing extended attribute: ", path, ". Resource fork is not supported.")
		return -fuse.ENOTSUP
	}
	if _, ok := node.xatr[name]; !ok {
		log.Error("Error removing extended attribute: ", path, ". Extended attribute does not exist.")
		return -fuse.ENOATTR
	}
	delete(node.xatr, name)
	return 0
}

func (c *CtbFs) Listxattr(path string, fill func(name string) bool) (errc int) {
	defer trace(path, fill)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if node == nil {
		log.Error("Error listing extended attributes: ", path, ". Node does not exist.")
		return -fuse.ENOENT
	}
	for name := range node.xatr {
		if !fill(name) {
			log.Error("Error listing extended attributes: ", path, ". Error filling extended attributes.")
			return -fuse.ERANGE
		}
	}
	return 0
}

func (c *CtbFs) Chflags(path string, flags uint32) (errc int) {
	defer trace(path, flags)(&errc)
	defer c.synchronize()()
	_, _, node := c.lookupNode(path, nil)
	if node == nil {
		log.Error("Error changing flags of node: ", path, ". Node does not exist.")
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
	if node == nil {
		log.Error("Error setting creation time of node: ", path, ". Node does not exist.")
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
	if node == nil {
		log.Error("Error setting change time of node: ", path, ". Node does not exist.")
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
