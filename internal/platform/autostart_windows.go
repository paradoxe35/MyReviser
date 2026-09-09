//go:build windows

package platform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paradoxe35/scribe/internal/logger"
	"golang.org/x/sys/windows/registry"
)

type autoStart struct{}

const (
	registryKey  = `Software\Microsoft\Windows\CurrentVersion\Run`
	registryName = "Scribe"
)

// Enable adds the application to Windows startup registry
func (a *autoStart) Enable() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks to get the real executable path
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		logger.Warn("Failed to resolve symlinks", "error", err)
	}

	// Add quotes around path to handle spaces correctly
	// This prevents the "Unquoted Service Path" vulnerability
	quotedPath := fmt.Sprintf(`"%s"`, executable)

	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	if err := key.SetStringValue(registryName, quotedPath); err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	logger.Info("Auto-start enabled for Windows", "executable", executable)
	return nil
}

// Disable removes the application from Windows startup registry
func (a *autoStart) Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	if err := key.DeleteValue(registryName); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	logger.Info("Auto-start disabled for Windows")
	return nil
}

// IsEnabled checks if the registry entry exists
func (a *autoStart) IsEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(registryName)
	return err == nil
}
