package utils

import (
	"os"
	"path/filepath"
)

func CreateTempFile(pattern string) (*os.File, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	cachePath := filepath.Join(cacheDir, ".ctb")
	tempFile, err := os.CreateTemp(cachePath, pattern)
	return tempFile, err
}

func CloseDeleteTempFile(file *os.File) {
	_ = file.Close()
	_ = os.Remove(file.Name())
}
