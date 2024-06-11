//go:build linux
// +build linux

package filesyetem_service

import (
	"syscall"
)

// getDiskUsage returns the total and free bytes available in the directory's disk partition
func (f *FileSystem) GetDiskUsage() (totalBytes, freeBytes uint64, err error) {
	var stat syscall.Statfs_t

	err = syscall.Statfs(f.linkRepo.GetRootPath(), &stat)
	if err != nil {
		return 0, 0, err
	}

	totalBytes = stat.Blocks * uint64(stat.Bsize)
	freeBytes = stat.Bfree * uint64(stat.Bsize)

	return totalBytes, freeBytes, nil
}
