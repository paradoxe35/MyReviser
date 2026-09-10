package platform

import (
	"os"
	"os/exec"
	"strings"

	"github.com/paradoxe35/encre/internal/logger"
)

// RestartApplication restarts the application using platform-specific methods
func RestartApplication() error {
	executable, err := os.Executable()
	if err != nil {
		logger.Error("Failed to get executable path", "error", err)
		return err
	}

	logger.Info("Restarting application", "executable", executable)

	// On macOS, use 'open' command to properly launch the app bundle
	// This handles both .app bundles and direct executables
	var cmd *exec.Cmd

	// Check if we're inside a .app bundle
	if strings.Contains(executable, ".app/Contents/MacOS/") {
		// Extract the .app path
		appPath := executable[:strings.Index(executable, ".app/Contents/MacOS/")+4]
		logger.Info("Detected .app bundle, using 'open' command", "app", appPath)
		// Use 'open -n' to force a new instance
		cmd = exec.Command("open", "-n", appPath)
	} else {
		// Direct executable
		logger.Info("Using direct executable launch")
		cmd = exec.Command(executable)
	}

	err = cmd.Start()
	if err != nil {
		logger.Error("Failed to start new instance", "error", err)
		return err
	}

	logger.Info("New instance started successfully")
	return nil
}
