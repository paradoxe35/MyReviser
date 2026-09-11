package ui

import (
	"fyne.io/fyne/v2"
)

const (
	windowWidth    = 600
	windowHeight   = 600
	titleBarHeight = 34
)

// newChromelessWindow uses native chrome, not actually chromeless, so the title bar stays draggable on every platform.
func newChromelessWindow(app fyne.App, title string) fyne.Window {
	return app.NewWindow(title)
}
