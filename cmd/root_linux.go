//go:build linux
// +build linux

package cmd

import (
	"os"
	"path/filepath"
)

func getLogPath() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		panic("HOME environment variable not set")
	}
	logDir := filepath.Join(homeDir, ".cognitechbridge", "logs")
	logPath := filepath.Join(logDir, "client.log")
	return logPath
}
