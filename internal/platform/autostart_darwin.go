//go:build darwin

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/paradoxe35/encre/internal/logger"
)

type autoStart struct{}

// getAppPath returns the path to the .app bundle
func getAppPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		logger.Warn("Failed to resolve symlinks", "error", err)
	}

	// Check if running from .app bundle
	// Path should be: Encre.app/Contents/MacOS/Encre
	if strings.Contains(executable, ".app/Contents/MacOS/") {
		// Extract .app path
		parts := strings.Split(executable, ".app/Contents/MacOS/")
		if len(parts) >= 1 {
			return parts[0] + ".app", nil
		}
	}

	// If not in .app bundle, return executable path directly
	return executable, nil
}

// Enable adds the app to macOS login items using osascript
func (a *autoStart) Enable() error {
	appPath, err := getAppPath()
	if err != nil {
		return err
	}

	// Use osascript to add login item - appears in "Open at Login"
	script := fmt.Sprintf(`tell application "System Events" to make new login item with properties {path:"%s", hidden:false} at end`, appPath)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if already exists
		if strings.Contains(string(output), "already exists") || strings.Contains(string(output), "duplicate") {
			logger.Info("Login item already exists", "path", appPath)
			return nil
		}
		return fmt.Errorf("failed to add login item: %w, output: %s", err, string(output))
	}

	logger.Info("Auto-start enabled for macOS (Open at Login)", "path", appPath)
	return nil
}

// Disable removes the app from macOS login items using osascript
func (a *autoStart) Disable() error {
	appPath, err := getAppPath()
	if err != nil {
		return err
	}

	// Extract app name from path
	appName := filepath.Base(appPath)

	// Use osascript to remove login item
	script := fmt.Sprintf(`tell application "System Events" to delete login item "%s"`, strings.TrimSuffix(appName, ".app"))

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if item doesn't exist
		if strings.Contains(string(output), "doesn't understand") || strings.Contains(string(output), "not found") {
			logger.Info("Login item not found (already removed)", "name", appName)
			return nil
		}
		return fmt.Errorf("failed to remove login item: %w, output: %s", err, string(output))
	}

	logger.Info("Auto-start disabled for macOS", "name", appName)
	return nil
}

// IsEnabled checks if the app is in macOS login items
func (a *autoStart) IsEnabled() bool {
	appPath, err := getAppPath()
	if err != nil {
		return false
	}

	// Use osascript to check if login item exists
	script := `tell application "System Events" to get the name of every login item`

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Extract app name
	appName := filepath.Base(appPath)
	appName = strings.TrimSuffix(appName, ".app")

	// Check if app name is in the list
	loginItems := string(output)
	return strings.Contains(loginItems, appName)
}
