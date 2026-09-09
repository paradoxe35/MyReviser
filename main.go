// Scribe - AI-powered text revision tool
// Author: Paradoxe Ng <contact@pngwasi.me>
// Repository: https://github.com/paradoxe35/scribe

package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	singleinstance "github.com/allan-simon/go-singleinstance"
	"github.com/paradoxe35/scribe/internal/config"
	"github.com/paradoxe35/scribe/internal/logger"
	"github.com/paradoxe35/scribe/internal/platform"
	"github.com/paradoxe35/scribe/internal/utils"
	"github.com/paradoxe35/scribe/internal/version"
)

func main() {
	// Initialize logger first
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Ensure ~/.scribe created
	utils.EnsureAppHomeDir()

	// Check for single instance
	lockPath := utils.AppHomeDir("scribe.lock")
	portPath := utils.AppHomeDir("instance.port")
	lockFile, err := singleinstance.CreateLockFile(lockPath)
	if err != nil {
		if platform.Notify(portPath) {
			logger.Info("Handed this launch to the Scribe already running")
			return
		}
		logger.Error("Another instance is already running", "error", err)
		os.Exit(1)
	}
	defer lockFile.Close()

	handover := platform.NewHandover(portPath)
	defer handover.Close()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		cfg = config.Default()
	}

	// Create Fyne application
	myApp := app.NewWithID(config.APP_ID)
	myApp.SetIcon(resourceIconPng)

	logger.Info("Scribe starting",
		"version", version.GetVersion(myApp),
		"build", version.GetBuildNumber(myApp))

	// Create and start the application
	application, err := NewApplication(myApp, cfg)
	if err != nil {
		logger.Error("Failed to create application", "error", err)
		os.Exit(1)
	}

	handover.Serve(application.ShowWindow)

	// Run the application
	if err := application.Start(); err != nil {
		logger.Error("Failed to start application", "error", err)
		os.Exit(1)
	}
}
