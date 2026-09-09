package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (w *MainWindow) createTranslateSection() fyne.CanvasObject {
	w.primaryLanguage = NewLanguagePicker(w.config.Translate.PrimaryLanguage)
	w.secondaryLanguage = NewLanguagePicker(w.config.Translate.SecondaryLanguage)

	swap := widget.NewButtonWithIcon("Swap", theme.MenuDropDownIcon(), func() {
		primary, secondary := w.primaryLanguage.Code(), w.secondaryLanguage.Code()
		w.primaryLanguage.SetCode(secondary)
		w.secondaryLanguage.SetCode(primary)
	})

	explanation := widget.NewLabel(
		"Translate detects the language of your selection. Text in your primary " +
			"language becomes secondary; anything else becomes primary.")
	explanation.Wrapping = fyne.TextWrapWord

	return container.NewScroll(container.NewPadded(container.NewVBox(
		explanation,
		widget.NewSeparator(),
		widget.NewForm(
			widget.NewFormItem("Primary", w.primaryLanguage),
			widget.NewFormItem("Secondary", w.secondaryLanguage),
		),
		container.NewCenter(swap),
		widget.NewSeparator(),
		widget.NewLabel("The shortcut and prompt live under Actions › Translate selection."),
	)))
}
