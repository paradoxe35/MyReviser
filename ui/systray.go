package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/paradoxe35/scribe/internal/logger"
)

func SetupSystemTray(desk desktop.App, mainWindow *MainWindow, onQuit func() error) {
	// Every item touches the UI, and a tray callback does not run on Fyne's thread.
	menu := fyne.NewMenu("Scribe",
		fyne.NewMenuItem("Settings", func() {
			fyne.Do(mainWindow.ShowWindow)
		}),
		fyne.NewMenuItem("View Logs", func() {
			if err := logger.OpenLogFile(); err != nil {
				logger.Error("Failed to open log file", "error", err)
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			if err := onQuit(); err != nil {
				fyne.Do(mainWindow.app.Quit)
			}
		}),
	)

	desk.SetSystemTrayMenu(menu)
}
