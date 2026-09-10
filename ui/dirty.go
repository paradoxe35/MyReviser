package ui

import (
	"fyne.io/fyne/v2/widget"
)

// Widgets that edit settings report through markDirty. These constructors wire
// that in once, so each section declares what it edits rather than repeating
// change callbacks.

// dirtyCheck builds a checkbox that marks settings unsaved when toggled.
func (w *MainWindow) dirtyCheck(label string, checked bool) *widget.Check {
	check := widget.NewCheck(label, func(bool) { w.markDirty() })
	check.SetChecked(checked)
	return check
}

// dirtyEntry builds a text entry that marks settings unsaved when edited.
func (w *MainWindow) dirtyEntry() *widget.Entry {
	entry := widget.NewEntry()
	entry.OnChanged = func(string) { w.markDirty() }
	return entry
}

// dirtyPasswordEntry builds a password entry that marks settings unsaved when edited.
func (w *MainWindow) dirtyPasswordEntry() *widget.Entry {
	entry := widget.NewPasswordEntry()
	entry.OnChanged = func(string) { w.markDirty() }
	return entry
}

// dirtyMultiLineEntry builds a multi-line entry that marks settings unsaved when edited.
func (w *MainWindow) dirtyMultiLineEntry() *widget.Entry {
	entry := widget.NewMultiLineEntry()
	entry.OnChanged = func(string) { w.markDirty() }
	return entry
}

// dirtySelectEntry builds a combo entry that marks settings unsaved when edited.
func (w *MainWindow) dirtySelectEntry() *widget.SelectEntry {
	entry := widget.NewSelectEntry(nil)
	entry.OnChanged = func(string) { w.markDirty() }
	return entry
}

// dirtySelect builds a dropdown that marks settings unsaved when the user picks
// an option. onSelected runs after the dirty mark; it may be nil.
func (w *MainWindow) dirtySelect(options []string, onSelected func(string)) *widget.Select {
	return widget.NewSelect(options, func(value string) {
		w.markDirty()
		if onSelected != nil {
			onSelected(value)
		}
	})
}
