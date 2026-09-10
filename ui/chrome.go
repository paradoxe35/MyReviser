package ui

import (
	"fyne.io/fyne/v2"
)

const (
	windowWidth    = 600
	windowHeight   = 600
	titleBarHeight = 34
)

// newChromelessWindow is retained as the local construction point for the
// settings window, but uses native window chrome so the title bar is draggable
// on every supported desktop platform.
func newChromelessWindow(app fyne.App, title string) fyne.Window {
	return app.NewWindow(title)
}
