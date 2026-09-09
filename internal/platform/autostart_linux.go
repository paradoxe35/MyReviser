//go:build linux

package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paradoxe35/scribe/internal/logger"
)

type autoStart struct{}

// escapeExecPath escapes a path for use in desktop entry Exec field
// According to XDG spec: quote reserved characters and escape ", $, `, \
func escapeExecPath(path string) string {
	// Check if path contains reserved characters that require quoting
	reservedChars := " \t\n\"'\\><~|&;$*?#()`"
	needsQuoting := false
	for _, char := range reservedChars {
		if strings.ContainsRune(path, char) {
			needsQuoting = true
			break
		}
	}

	if !needsQuoting {
		return path
	}

	// Escape special characters: ", $, `, \
	escaped := path
	escaped = strings.ReplaceAll(escaped, `\`, `\\`)  // Backslash first!
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)  // Double quote
	escaped = strings.ReplaceAll(escaped, "`", "\\`") // Backtick
	escaped = strings.ReplaceAll(escaped, `$`, `\$`)  // Dollar sign

	// Wrap in double quotes
	return fmt.Sprintf(`"%s"`, escaped)
}

// Enable creates an autostart desktop entry for Linux
func (a *autoStart) Enable() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks to get the real executable path
	// This is important for AppImages and symlinked installations
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		logger.Warn("Failed to resolve symlinks", "error", err)
		// Continue with original path if symlink resolution fails
	}

	autostartDir := filepath.Join(homeDir, ".config", "autostart")
	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return fmt.Errorf("failed to create autostart directory: %w", err)
	}

	desktopFilePath := filepath.Join(autostartDir, "scribe.desktop")

	// Escape the executable path according to XDG desktop entry spec
	escapedExec := escapeExecPath(executable)

	desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Scribe
Comment=AI-powered text revision tool
Exec=%s
Icon=scribe
Terminal=false
StartupNotify=false
X-GNOME-Autostart-enabled=true
Categories=Utility;Office;
`, escapedExec)

	if err := os.WriteFile(desktopFilePath, []byte(desktopContent), 0644); err != nil {
		return fmt.Errorf("failed to write desktop file: %w", err)
	}

	logger.Info("Auto-start enabled for Linux", "desktop_file", desktopFilePath)
	return nil
}

// Disable removes the autostart desktop entry
func (a *autoStart) Disable() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	desktopFilePath := filepath.Join(homeDir, ".config", "autostart", "scribe.desktop")

	if err := os.Remove(desktopFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove desktop file: %w", err)
	}

	logger.Info("Auto-start disabled for Linux")
	return nil
}

// IsEnabled checks if the desktop entry exists
func (a *autoStart) IsEnabled() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	desktopFilePath := filepath.Join(homeDir, ".config", "autostart", "scribe.desktop")
	_, err = os.Stat(desktopFilePath)
	return err == nil
}
