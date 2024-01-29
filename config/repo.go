package config

import (
	"os"
	"path/filepath"
)

func GetRepoRoot() (string, error) {
	return os.Getwd()
}

func GetRepoCtbRoot() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, ".ctb")
	return path, nil
}
