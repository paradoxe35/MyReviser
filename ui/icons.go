package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed assets/keyboard.svg
var keyboardSVG []byte

// Fyne ships no keyboard icon; ThemedResource recolors it to match light/dark theme.
var keyboardIcon = theme.NewThemedResource(
	&fyne.StaticResource{StaticName: "keyboard.svg", StaticContent: keyboardSVG},
)
