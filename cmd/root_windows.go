//go:build windows
// +build windows

package cmd

import (
	"os"
	"path/filepath"
)

func getLogPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		panic("APPDATA environment variable not set")
	}
	logDir := filepath.Join(appData, "com.cognitechbridge.app", "logs")
	logPath := filepath.Join(logDir, "client.log")
	return logPath
}
