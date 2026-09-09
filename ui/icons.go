package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed assets/keyboard.svg
var keyboardSVG []byte

// Fyne ships no keyboard icon, and every stand-in read as something else.
// ThemedResource recolours the fills to match light or dark.
var keyboardIcon = theme.NewThemedResource(
	&fyne.StaticResource{StaticName: "keyboard.svg", StaticContent: keyboardSVG},
)
