package fuse

// @Todo Implement links later
//func (self *Cache) Link(oldpath string, newpath string) (errc int) {
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
//func (self *Cache) Symlink(target string, newpath string) (errc int) {
//	defer trace(target, newpath)(&errc)
//	defer self.synchronize()()
//	return self.makeNode(newpath, fuse.S_IFLNK|00777, 0, []byte(target))
//}
//
//func (self *Cache) Readlink(path string) (errc int, target string) {
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
