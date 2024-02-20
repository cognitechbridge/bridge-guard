package config

import (
	"os"
	"path/filepath"
)

func GetTempRoot() (string, error) {
	path := filepath.Join(os.TempDir(), ".ctb")
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

func GetRepoCtbRoot() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path, nil
}
