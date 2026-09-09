package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	windowWidth    = 480
	windowHeight   = 560
	titleBarHeight = 34
)

// newChromelessWindow prefers a borderless window so Encre draws its own title
// bar. Fyne only exposes that through the desktop driver's splash window, and
// only at creation time, so a driver without it falls back to native chrome.
func newChromelessWindow(app fyne.App, title string) fyne.Window {
	if drv, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
		window := drv.CreateSplashWindow()
		window.SetTitle(title)
		return window
	}
	return app.NewWindow(title)
}

// chrome wraps content in a custom title bar. Closing hides to the tray, which
// is what the close intercept does for a native title bar too.
func (w *MainWindow) chrome(content fyne.CanvasObject) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Encre", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	closeButton := widget.NewButtonWithIcon("", theme.CancelIcon(), w.HideWindow)
	closeButton.Importance = widget.LowImportance

	background := canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground))
	background.SetMinSize(fyne.NewSize(0, titleBarHeight))

	bar := container.NewStack(background, container.NewBorder(
		nil, nil,
		container.NewPadded(title),
		closeButton,
	))

	return container.NewBorder(
		container.NewVBox(bar, widget.NewSeparator()),
		nil, nil, nil,
		content,
	)
}
