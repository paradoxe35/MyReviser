package utils

import (
	"log"
	"os"
	"path/filepath"
)

const appDirName = ".scribe"

func AppHomeDir(elem ...string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory", "error", err)
	}

	parts := append([]string{homeDir, appDirName}, elem...)
	return filepath.Join(parts...)
}

func EnsureAppHomeDir() {
	if err := os.MkdirAll(AppHomeDir(), 0755); err != nil {
		log.Fatal("Failed to create app home directory", "error", err)
	}
}
