package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (w *MainWindow) translateLanguages() fyne.CanvasObject {
	w.primaryLanguage = NewLanguagePicker(w.config.Translate.PrimaryLanguage)
	w.secondaryLanguage = NewLanguagePicker(w.config.Translate.SecondaryLanguage)

	updateExclusions := func() {
		w.primaryLanguage.SetExcludedCode(w.secondaryLanguage.Code())
		w.secondaryLanguage.SetExcludedCode(w.primaryLanguage.Code())
	}
	w.primaryLanguage.SetOnCodeChanged(func(string) { updateExclusions(); w.markDirty() })
	w.secondaryLanguage.SetOnCodeChanged(func(string) { updateExclusions(); w.markDirty() })
	updateExclusions()

	swap := widget.NewButtonWithIcon("Swap", theme.ViewRefreshIcon(), func() {
		primary, secondary := w.primaryLanguage.Code(), w.secondaryLanguage.Code()
		// Clear the cross-exclusions while assigning the pair. Otherwise the
		// first Select change can reject the second side of the swap.
		w.primaryLanguage.SetExcludedCode("")
		w.secondaryLanguage.SetExcludedCode("")
		w.primaryLanguage.SetCode(secondary)
		w.secondaryLanguage.SetCode(primary)
		updateExclusions()
		w.markDirty()
	})
	swap.Importance = widget.LowImportance

	explanation := widget.NewLabel(
		"Text in your primary language becomes secondary; anything else becomes primary.")
	explanation.Wrapping = fyne.TextWrapWord
	explanation.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Primary", w.primaryLanguage.Select),
			widget.NewFormItem("Secondary", w.secondaryLanguage.Select),
		),
		container.NewBorder(nil, nil, nil, swap, explanation),
	)
}
