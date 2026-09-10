package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (w *MainWindow) createStatusBar() fyne.CanvasObject {
	statusLabel := widget.NewLabel("")
	statusLabel.Bind(w.statusBinding)
	statusLabel.Truncation = fyne.TextTruncateEllipsis

	return container.NewBorder(
		nil, nil,
		widget.NewIcon(theme.InfoIcon()),
		nil,
		statusLabel,
	)
}
