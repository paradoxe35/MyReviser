//go:build !darwin

package platform

import (
	"os"
	"os/exec"

	"github.com/paradoxe35/scribe/internal/logger"
)

// RestartApplication restarts the application using platform-specific methods
func RestartApplication() error {
	executable, err := os.Executable()
	if err != nil {
		logger.Error("Failed to get executable path", "error", err)
		return err
	}

	logger.Info("Restarting application", "executable", executable)

	// For Linux/Windows, use direct execution
	cmd := exec.Command(executable)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		logger.Error("Failed to start new instance", "error", err)
		return err
	}

	logger.Info("New instance started successfully")
	return nil
}
