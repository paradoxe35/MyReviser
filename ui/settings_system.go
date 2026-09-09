package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/scribe/internal/logger"
	"github.com/paradoxe35/scribe/internal/platform"
	"github.com/paradoxe35/scribe/internal/version"
)

func (w *MainWindow) createSystemSection() fyne.CanvasObject {
	// Theme selection
	themeLabel := widget.NewLabel("Theme:")
	themeLabel.TextStyle.Bold = true

	themeSelect := widget.NewSelect(
		[]string{"auto", "light", "dark"},
		func(value string) {
			w.themeBinding.Set(value)
			w.applyTheme(value)
		},
	)

	// Set initial selection
	currentTheme, _ := w.themeBinding.Get()
	themeSelect.SetSelected(currentTheme)

	// Theme description
	themeDesc := widget.NewLabel("Auto: Follow system theme\nLight: Always use light theme\nDark: Always use dark theme")
	themeDesc.Wrapping = fyne.TextWrapWord

	// Start Minimized checkbox
	startMinimizedCheck := widget.NewCheck("Start minimized to system tray", func(checked bool) {
		w.startMinimizedBinding.Set(checked)
	})
	startMinimizedCheck.Bind(w.startMinimizedBinding)

	// Start on Login checkbox
	startOnLoginCheck := widget.NewCheck("Start on login", func(checked bool) {
		w.startOnLoginBinding.Set(checked)
	})
	startOnLoginCheck.Bind(w.startOnLoginBinding)

	// Version display (only show for production builds)
	var versionContainer *fyne.Container
	if version.IsProduction(w.app) {
		versionLabel := widget.NewLabel(fmt.Sprintf("Version: %s", version.GetVersion(w.app)))
		versionLabel.TextStyle.Italic = true
		// Use a subtle gray color for the version
		versionLabel.Importance = widget.LowImportance

		versionContainer = container.NewVBox(
			widget.NewSeparator(),
			container.NewPadded(versionLabel),
		)
	}

	formItems := []fyne.CanvasObject{
		container.NewPadded(container.NewVBox(themeLabel, themeSelect, themeDesc)),
		widget.NewSeparator(),
		container.NewPadded(container.NewVBox(startMinimizedCheck, startOnLoginCheck)),
	}

	// Add version if production build
	if versionContainer != nil {
		formItems = append(formItems, versionContainer)
	}

	form := container.NewVBox(formItems...)

	return container.NewScroll(form)
}
func (w *MainWindow) applyTheme(themeName string) {
	switch themeName {
	case "light":
		w.app.Settings().SetTheme(&forceLight{})
	case "dark":
		w.app.Settings().SetTheme(&forceDark{})
	case "auto":
		w.app.Settings().SetTheme(theme.DefaultTheme())
	}
}
func (w *MainWindow) applyAutoStartSetting(enabled bool) {
	autoStart := platform.GetAutoStart()

	if enabled {
		if err := autoStart.Enable(); err != nil {
			logger.Error("Failed to enable auto-start", "error", err)
			w.statusBinding.Set("✗ Failed to enable auto-start")
		} else {
			logger.Info("Auto-start enabled")
		}
	} else {
		if err := autoStart.Disable(); err != nil {
			logger.Error("Failed to disable auto-start", "error", err)
			w.statusBinding.Set("✗ Failed to disable auto-start")
		} else {
			logger.Info("Auto-start disabled")
		}
	}
}

// restartApplication restarts the application
func (w *MainWindow) restartApplication() {
	logger.Info("User requested application restart")

	go func() {
		time.Sleep(200 * time.Millisecond)

		err := platform.RestartApplication()
		if err != nil {
			logger.Error("Failed to restart application", "error", err)
			return
		}

		logger.Info("New instance started, quitting current instance")

		fyne.Do(func() {
			w.app.Quit()
		})
	}()
}
