//go:build windows
// +build windows

package filesyetem_service

import (
	"golang.org/x/sys/windows"
)

// getDiskUsage returns the total and free bytes available in the directory's disk partition
func (f *FileSystem) GetDiskUsage() (totalBytes, freeBytes uint64, err error) {
	var freeBytesAvailable uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	err = windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(f.linkRepo.GetRootPath()),
		&freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		return 0, 0, err
	}
	return totalNumberOfBytes, freeBytesAvailable, nil
}
