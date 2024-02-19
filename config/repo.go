package config

import (
	"os"
)

func GetRepoRoot() (string, error) {
	return os.Getwd()
}

func GetRepoCtbRoot() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path, nil
}
