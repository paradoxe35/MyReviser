package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/paradoxe35/scribe/internal/utils"
)

var (
	defaultLogger  *slog.Logger
	currentLogFile string
)

// Init initializes the logger with daily log rotation
func Init() error {
	logDir := utils.AppHomeDir("logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Use daily log file format: scribe-2025-09-29.log
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("scribe-%s.log", today))
	currentLogFile = logFile

	// Open log file with append mode
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create text handler with options
	// Check for DEBUG environment variable
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") != "" {
		logLevel = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(file, opts)
	defaultLogger = slog.New(handler)

	// Set as default
	slog.SetDefault(defaultLogger)

	// Log initialization
	defaultLogger.Info("Logger initialized", "log_file", logFile)

	// Clean up old log files (keep last 30 days)
	go cleanupOldLogs(logDir, 30)

	return nil
}

// Convenience functions for logging
func Info(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, args...)
	}
}

func Debug(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, args...)
	}
}

// GetCurrentLogFile returns the path to the current log file
func GetCurrentLogFile() string {
	if currentLogFile != "" {
		return currentLogFile
	}

	// Fallback: construct path with today's date
	today := time.Now().Format("2006-01-02")
	return utils.AppHomeDir("logs", fmt.Sprintf("scribe-%s.log", today))
}

// GetLogDirectory returns the path to the logs directory
func GetLogDirectory() string {
	return utils.AppHomeDir("logs")
}

// cleanupOldLogs removes log files older than the specified number of days
func cleanupOldLogs(logDir string, maxAgeDays int) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a log file matching our pattern
		if filepath.Ext(entry.Name()) != ".log" {
			continue
		}

		logPath := filepath.Join(logDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Remove if older than cutoff date
		if info.ModTime().Before(cutoffDate) {
			if err := os.Remove(logPath); err == nil {
				Info("Removed old log file", "file", entry.Name())
			}
		}
	}
}
