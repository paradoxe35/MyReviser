package ui

import (
	"fyne.io/fyne/v2/widget"
)

// These constructors wire markDirty into each widget once, so callers don't repeat change callbacks.

func (w *MainWindow) dirtyCheck(label string, checked bool) *widget.Check {
	check := widget.NewCheck(label, func(bool) { w.markDirty() })
	check.SetChecked(checked)
	return check
}

func (w *MainWindow) dirtyEntry() *widget.Entry {
	entry := widget.NewEntry()
	entry.OnChanged = func(string) { w.markDirty() }
	return entry
}

func (w *MainWindow) dirtyPasswordEntry() *widget.Entry {
	entry := widget.NewPasswordEntry()
	entry.OnChanged = func(string) { w.markDirty() }
	return entry
}

func (w *MainWindow) dirtyMultiLineEntry() *widget.Entry {
	entry := widget.NewMultiLineEntry()
	entry.OnChanged = func(string) { w.markDirty() }
	return entry
}

func (w *MainWindow) dirtySelectEntry() *widget.SelectEntry {
	entry := widget.NewSelectEntry(nil)
	entry.OnChanged = func(string) { w.markDirty() }
	return entry
}

// onSelected runs after the dirty mark; it may be nil.
func (w *MainWindow) dirtySelect(options []string, onSelected func(string)) *widget.Select {
	return widget.NewSelect(options, func(value string) {
		w.markDirty()
		if onSelected != nil {
			onSelected(value)
		}
	})
}
