package config

import (
	"os"
	"path/filepath"
)

// Config represents the configuration of the application
type Config struct {
	repoPath string // path to the repository
	tempPath string // path to the temporary folder of the application
}

// New returns a new Config
func New(repoPath string, tempPath string, cfgFile string) (*Config, error) {
	return &Config{
		repoPath: repoPath,
		tempPath: tempPath,
	}, nil
}

// GetTempRoot returns the root path of the temporary folder.
func (c *Config) GetTempRoot() (string, error) {
	if err := os.MkdirAll(c.tempPath, os.ModePerm); err != nil {
		panic("Cannot create temp path")
	}
	return c.tempPath, nil
}

// GetRepoCtbRoot returns the root path of the repository.
func (c *Config) GetRepoCtbRoot() (string, error) {
	if err := os.MkdirAll(c.repoPath, os.ModePerm); err != nil {
		panic("Cannot create repo root path")
	}
	return c.repoPath, nil
}

// GetTempRoot returns the root path of the temporary folder.
func (c *Config) GetCacheRoot() (string, error) {
	path := filepath.Join(c.tempPath, "cache")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic("Cannot create cache path")
	}
	return path, nil
}
