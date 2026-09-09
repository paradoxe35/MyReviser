package ui

import (
	"errors"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/scribe/internal/config"
)

const providerDefaultOption = "Default"

// actionEditor holds the widgets for one action so save and reset can reach
// them without the window carrying a field per action.
type actionEditor struct {
	kind     config.ActionKind
	enabled  *widget.Check
	capture  *HotkeyCapture
	prompt   *widget.Entry
	limit    *widget.Entry
	timeout  *widget.Slider
	provider *widget.Select
}

func (w *MainWindow) createActionsSection() fyne.CanvasObject {
	w.actionEditors = make(map[config.ActionKind]*actionEditor, len(config.ActionOrder))

	items := make([]*widget.AccordionItem, 0, len(config.ActionOrder))
	captures := make([]*HotkeyCapture, 0, len(config.ActionOrder))

	for _, kind := range config.ActionOrder {
		editor := w.newActionEditor(kind)
		w.actionEditors[kind] = editor
		captures = append(captures, editor.capture)
		items = append(items, widget.NewAccordionItem(kind.Label(), editor.content(kind)))
	}

	// Every capture must know every other, so only one records at a time and
	// duplicate bindings are rejected.
	for _, capture := range captures {
		capture.SetSiblings(others(captures, capture)...)
	}

	accordion := widget.NewAccordion(items...)
	accordion.Open(0)

	w.mentionsCheck = widget.NewCheck("Enable @provider mentions", nil)
	w.mentionsCheck.SetChecked(w.config.EnableProviderMentions)

	mentionsHelp := widget.NewLabel(
		"Start a selection with @provider to run that one action on it, e.g. \"@claude Fix this\".")
	mentionsHelp.Wrapping = fyne.TextWrapWord
	mentionsHelp.TextStyle = fyne.TextStyle{Italic: true}

	help := widget.NewLabel(
		"• Click 'Capture', press keys one at a time, then Enter to save\n" +
			"• Requires at least one modifier (Ctrl/Alt/Shift/Super)\n" +
			"• Press ESC to cancel")
	help.Wrapping = fyne.TextWrapWord

	reset := widget.NewButton("Reset shortcuts to defaults", func() {
		defaults := config.DefaultActions()
		for kind, editor := range w.actionEditors {
			editor.capture.StopCapture()
			w.hotkeyBindings[kind].Set(defaults[kind].Hotkey)
			editor.capture.UpdateFromBinding()
		}
	})

	return container.NewScroll(container.NewVBox(
		accordion,
		widget.NewSeparator(),
		container.NewPadded(container.NewVBox(
			w.mentionsCheck,
			mentionsHelp,
			widget.NewAccordion(widget.NewAccordionItem("How to capture shortcuts", help)),
			reset,
		)),
	))
}

func others(all []*HotkeyCapture, self *HotkeyCapture) []*HotkeyCapture {
	rest := make([]*HotkeyCapture, 0, len(all)-1)
	for _, capture := range all {
		if capture != self {
			rest = append(rest, capture)
		}
	}
	return rest
}

func (w *MainWindow) newActionEditor(kind config.ActionKind) *actionEditor {
	action := w.config.Action(kind)

	capture := NewHotkeyCapture(w.hotkeyBindings[kind], "Click 'Capture' to set shortcut")
	capture.window = w.Window
	capture.SetAllowModifierOnly(true)
	capture.onCaptureStart = func() {
		if w.hotkeyManager != nil {
			w.hotkeyManager.Disable()
		}
	}
	capture.onCaptureStop = func() {
		if w.hotkeyManager != nil {
			w.hotkeyManager.Enable()
		}
	}

	editor := &actionEditor{
		kind:    kind,
		enabled: widget.NewCheck("Enabled", nil),
		capture: capture,
	}
	editor.enabled.SetChecked(action.Enabled)

	if !kind.UsesAI() {
		return editor
	}

	editor.prompt = widget.NewMultiLineEntry()
	editor.prompt.SetText(action.SystemPrompt)
	editor.prompt.SetPlaceHolder("Leave empty to use the built-in prompt")
	editor.prompt.Wrapping = fyne.TextWrapWord
	editor.prompt.SetMinRowsVisible(6)

	editor.limit = widget.NewEntry()
	editor.limit.SetText(strconv.Itoa(action.CharacterLimit))
	editor.limit.Validator = func(value string) error {
		return validateCharacterLimit(value)
	}

	editor.timeout = widget.NewSlider(15, 300)
	editor.timeout.Step = 5
	editor.timeout.SetValue(float64(action.TimeoutSeconds))

	editor.provider = widget.NewSelect(w.providerOptions(), nil)
	editor.provider.SetSelected(providerLabel(action.ProviderID))

	return editor
}

func (e *actionEditor) content(kind config.ActionKind) fyne.CanvasObject {
	rows := []fyne.CanvasObject{e.enabled, e.capture}

	if e.prompt != nil {
		timeoutValue := widget.NewLabel("")
		syncTimeoutLabel(timeoutValue, e.timeout.Value)
		e.timeout.OnChanged = func(value float64) { syncTimeoutLabel(timeoutValue, value) }

		reset := widget.NewButton("Reset prompt", func() { e.prompt.SetText("") })

		rows = append(rows,
			widget.NewSeparator(),
			widget.NewForm(
				widget.NewFormItem("Provider", e.provider),
				widget.NewFormItem("Character limit", e.limit),
				widget.NewFormItem("Timeout", container.NewBorder(nil, nil, nil, timeoutValue, e.timeout)),
			),
			widget.NewLabel("System prompt"),
			e.prompt,
			reset,
		)
	}

	if kind == config.ActionDictate {
		rows = append(rows, widget.NewLabel("Speech-to-text is not available yet."))
	}

	return container.NewPadded(container.NewVBox(rows...))
}

func syncTimeoutLabel(label *widget.Label, value float64) {
	label.SetText(strconv.Itoa(int(value)) + "s")
}

func (w *MainWindow) providerOptions() []string {
	names := w.config.GetAllProviderNames()
	return append([]string{providerDefaultOption}, names...)
}

func providerLabel(id string) string {
	if id == "" {
		return providerDefaultOption
	}
	return id
}

func providerID(label string) string {
	if label == providerDefaultOption {
		return ""
	}
	return label
}

func validateCharacterLimit(value string) error {
	if value == "" {
		return nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil {
		return errors.New("must be a number")
	}
	if limit < 100 || limit > 20000 {
		return errors.New("must be between 100 and 20000")
	}
	return nil
}
