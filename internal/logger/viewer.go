package logger

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// GetLatestLogFile finds the most recent log file in the logs directory
func GetLatestLogFile() (string, error) {
	logDir := GetLogDirectory()

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return "", fmt.Errorf("failed to read log directory: %w", err)
	}

	// Filter log files matching pattern: encre-YYYY-MM-DD.log
	var logFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "encre-") && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, entry.Name())
		}
	}

	if len(logFiles) == 0 {
		return "", fmt.Errorf("no log files found")
	}

	// Sort in reverse order (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(logFiles)))

	return filepath.Join(logDir, logFiles[0]), nil
}

// OpenLogFile opens the latest log file in the system's default text editor
func OpenLogFile() error {
	// Try to get the latest log file
	logFile, err := GetLatestLogFile()
	if err != nil {
		Info("Failed to find latest log file, opening directory instead", "error", err)
		return OpenLogDirectory()
	}

	// Check if file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		Info("Log file does not exist, opening directory instead", "file", logFile)
		return OpenLogDirectory()
	}

	Info("Opening log file", "file", logFile)

	// Platform-specific command to open file in default editor
	var cmd *exec.Cmd
	var openErr error

	switch runtime.GOOS {
	case "windows":
		// Try notepad first (always available on Windows)
		cmd = exec.Command("notepad.exe", logFile)
		openErr = cmd.Start()
		if openErr != nil {
			// Fallback: Use explorer to open the file with default text editor
			cmd = exec.Command("explorer.exe", logFile)
			openErr = cmd.Start()
		}
	case "darwin":
		cmd = exec.Command("open", logFile)
		openErr = cmd.Start()
	case "linux":
		cmd = exec.Command("xdg-open", logFile)
		openErr = cmd.Start()
	default:
		// Fallback: try to open directory
		return OpenLogDirectory()
	}

	// If failed to open file, try opening the logs directory instead
	if openErr != nil {
		Warn("Failed to open log file, opening directory instead", "error", openErr)
		return OpenLogDirectory()
	}

	return nil
}

// OpenLogDirectory opens the logs directory in the system's file manager
func OpenLogDirectory() error {
	logDir := GetLogDirectory()

	Info("Opening log directory", "directory", logDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// On Windows, use explorer.exe to open the directory
		cmd = exec.Command("explorer.exe", logDir)
	case "darwin":
		cmd = exec.Command("open", logDir)
	case "linux":
		cmd = exec.Command("xdg-open", logDir)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	err := cmd.Start()
	if err != nil {
		Error("Failed to open log directory", "error", err, "directory", logDir)
		return fmt.Errorf("failed to open directory: %w", err)
	}

	return nil
}
