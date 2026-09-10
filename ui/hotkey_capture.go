package ui

import (
	"image/color"
	"runtime"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// HotkeyCapture is a custom widget for capturing keyboard shortcuts using Fyne's keyboard events
type HotkeyCapture struct {
	widget.BaseWidget
	binding        binding.String
	captureBtn     *widget.Button
	stopBtn        *widget.Button
	clearBtn       *widget.Button
	container      *fyne.Container
	displayLabel   *widget.Label    // Label to display saved keybinding
	entry          *captureEntry    // Entry to capture keyboard events
	window         fyne.Window      // Reference to parent window for focus
	onCaptureStart func()           // Callback when capture starts
	onCaptureStop  func()           // Callback when capture stops
	onChanged      func()           // Callback when a binding changes
	siblings       []*HotkeyCapture // Other capture widgets to disable during capture

	isCapturing bool
	mu          sync.Mutex

	pressedKeys       map[fyne.KeyName]bool
	modifiers         map[fyne.KeyModifier]bool
	allowModifierOnly bool
}

// captureEntry is a custom entry widget that captures keyboard events with white background
type captureEntry struct {
	widget.Entry
	parent *HotkeyCapture
}

func (e *captureEntry) TypedKey(key *fyne.KeyEvent) {
	if e.parent != nil {
		e.parent.handleKeyPress(key)
	}
}

// TypedRune prevents normal text input during capture
func (e *captureEntry) TypedRune(r rune) {
	// Block all typed characters - we only want key events
	// This prevents 'f' from appearing as text when pressing F key
}

func (e *captureEntry) CreateRenderer() fyne.WidgetRenderer {
	e.ExtendBaseWidget(e)
	renderer := e.Entry.CreateRenderer()

	// Use themed background for better visibility
	return &themedBackgroundRenderer{
		WidgetRenderer: renderer,
		entry:          e,
	}
}

type themedBackgroundRenderer struct {
	fyne.WidgetRenderer
	entry *captureEntry
}

func (r *themedBackgroundRenderer) BackgroundColor() color.Color {
	// Use theme-appropriate background color for input fields
	if r.entry.Disabled() {
		return theme.Color(theme.ColorNameDisabledButton)
	}
	return theme.Color(theme.ColorNameInputBackground)
}

// NewHotkeyCapture creates a new hotkey capture widget
func NewHotkeyCapture(binding binding.String, placeholder string) *HotkeyCapture {
	h := &HotkeyCapture{
		binding:     binding,
		pressedKeys: make(map[fyne.KeyName]bool),
		modifiers:   make(map[fyne.KeyModifier]bool),
	}

	// Create label to display saved keybinding
	h.displayLabel = widget.NewLabel(placeholder)
	h.displayLabel.TextStyle.Bold = true
	h.displayLabel.TextStyle.Monospace = true

	// Update label when binding changes
	currentValue, _ := binding.Get()
	if currentValue != "" {
		h.displayLabel.SetText(currentValue)
	}

	// Create entry for keyboard capture (hidden by default)
	h.entry = &captureEntry{parent: h}
	h.entry.PlaceHolder = "Press keys in sequence (ESC to cancel, Enter to save)"
	h.entry.TextStyle.Bold = true
	h.entry.TextStyle.Monospace = true
	h.entry.Hide() // Hidden until capture starts

	// Create capture button
	h.captureBtn = widget.NewButtonWithIcon("Capture", theme.MediaRecordIcon(), func() {
		h.startCapture()
	})

	// Create stop button to save current combination (initially hidden)
	h.stopBtn = widget.NewButtonWithIcon("Save", theme.ConfirmIcon(), func() {
		h.saveAndStop()
	})
	h.stopBtn.Importance = widget.HighImportance
	h.stopBtn.Hide()

	// Create clear button (hidden by default, will be removed from layout)
	h.clearBtn = widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		h.clearHotkey()
	})
	h.clearBtn.Importance = widget.LowImportance
	h.clearBtn.Hide() // Hide clear button entirely

	// Create button container (without clear button)
	buttonContainer := container.NewHBox(h.captureBtn, h.stopBtn)

	// Create a stack with label on bottom and entry on top (only one visible at a time)
	displayStack := container.NewStack(h.displayLabel, h.entry)

	// Create container
	h.container = container.NewBorder(
		nil, nil,
		nil,
		buttonContainer,
		displayStack,
	)

	h.ExtendBaseWidget(h)
	return h
}

// CreateRenderer implements fyne.Widget
func (h *HotkeyCapture) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.container)
}

// startCapture begins listening for key combinations
func (h *HotkeyCapture) startCapture() {
	h.mu.Lock()
	if h.isCapturing {
		h.mu.Unlock()
		return
	}
	h.isCapturing = true
	h.mu.Unlock()

	// Disable sibling capture buttons
	for _, sibling := range h.siblings {
		sibling.captureBtn.Disable()
	}

	// Notify that capture is starting (disable global hotkeys)
	if h.onCaptureStart != nil {
		h.onCaptureStart()
	}

	// Reset pressed keys
	h.pressedKeys = make(map[fyne.KeyName]bool)
	h.modifiers = make(map[fyne.KeyModifier]bool)

	// Update UI - hide label, show entry
	h.displayLabel.Hide()
	h.entry.SetText("")
	h.entry.Show()
	h.captureBtn.Hide()
	h.stopBtn.Show()

	// Focus the entry widget so it receives keyboard events
	if h.window != nil {
		h.window.Canvas().Focus(h.entry)
	}
}

// saveAndStop saves the current combination and stops capture
func (h *HotkeyCapture) saveAndStop() {
	// Try to save the hotkey
	success := h.saveHotkey()

	// Always stop capture (restore previous value if save failed)
	h.stopCapture()

	// If save failed, the error message was shown briefly in the label
	// stopCapture will restore the correct previous value
	_ = success
}

// stopCapture stops listening for keys (called after successful save or on cancel)
func (h *HotkeyCapture) stopCapture() {
	h.mu.Lock()
	if !h.isCapturing {
		h.mu.Unlock()
		return
	}

	h.isCapturing = false
	h.mu.Unlock()

	// Re-enable sibling capture buttons
	for _, sibling := range h.siblings {
		sibling.captureBtn.Enable()
	}

	// Notify that capture is stopping (re-enable global hotkeys)
	if h.onCaptureStop != nil {
		h.onCaptureStop()
	}

	// Update UI - hide entry, show label
	h.entry.Hide()
	h.stopBtn.Hide()
	h.captureBtn.Show()

	// Restore previous value in label or show placeholder
	currentValue, _ := h.binding.Get()
	if currentValue != "" {
		h.displayLabel.SetText(currentValue)
	} else {
		h.displayLabel.SetText("Click 'Capture' to set hotkey")
	}
	h.displayLabel.Show()
}

// handleKeyPress processes key press events from Fyne
func (h *HotkeyCapture) handleKeyPress(key *fyne.KeyEvent) {
	h.mu.Lock()

	if !h.isCapturing {
		h.mu.Unlock()
		return
	}

	// Handle ESC to cancel
	if key.Name == fyne.KeyEscape {
		h.mu.Unlock()
		h.stopCapture()
		return
	}

	// Handle Enter to save
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		h.mu.Unlock()
		h.saveAndStop()
		return
	}

	// Check if this is a modifier key press and track it
	switch key.Name {
	case desktop.KeyShiftLeft, desktop.KeyShiftRight:
		h.modifiers[fyne.KeyModifierShift] = true
	case desktop.KeyControlLeft, desktop.KeyControlRight:
		h.modifiers[fyne.KeyModifierControl] = true
	case desktop.KeyAltLeft, desktop.KeyAltRight:
		h.modifiers[fyne.KeyModifierAlt] = true
	case desktop.KeySuperLeft, desktop.KeySuperRight:
		h.modifiers[fyne.KeyModifierSuper] = true
	default:
		// One key, replacing any previous. The listener matches modifiers plus a single key, so a
		// combination naming two was stored with only the last one meaningful — and a map has no
		// order, so which one that was varied between saves.
		if !isModifierKey(key.Name) {
			h.pressedKeys = map[fyne.KeyName]bool{key.Name: true}
		}
	}

	h.updateDisplay()
	h.mu.Unlock()
}

// updateDisplay updates the entry with current pressed keys
func (h *HotkeyCapture) updateDisplay() {
	parts := []string{}

	// Add modifiers in standard order FIRST
	if h.modifiers[fyne.KeyModifierControl] {
		parts = append(parts, "ctrl")
	}
	if h.modifiers[fyne.KeyModifierAlt] {
		parts = append(parts, getAltName())
	}
	if h.modifiers[fyne.KeyModifierShift] {
		parts = append(parts, "shift")
	}
	if h.modifiers[fyne.KeyModifierSuper] {
		parts = append(parts, getSuperName())
	}

	// Then add regular keys
	if len(h.pressedKeys) > 0 {
		keyNames := make([]string, 0, len(h.pressedKeys))
		for keyName := range h.pressedKeys {
			keyNames = append(keyNames, keyNameToString(keyName))
		}
		parts = append(parts, keyNames...)
	}

	if len(parts) > 0 {
		h.entry.SetText(strings.Join(parts, "+"))
	} else {
		h.entry.SetText("")
	}
}

// saveHotkey saves the captured hotkey combination and returns true if successful
func (h *HotkeyCapture) saveHotkey() bool {
	parts := []string{}

	// Add modifiers in standard order
	if h.modifiers[fyne.KeyModifierControl] {
		parts = append(parts, "ctrl")
	}
	if h.modifiers[fyne.KeyModifierAlt] {
		parts = append(parts, getAltName())
	}
	if h.modifiers[fyne.KeyModifierShift] {
		parts = append(parts, "shift")
	}
	if h.modifiers[fyne.KeyModifierSuper] {
		parts = append(parts, getSuperName())
	}

	// Add regular keys
	keyNames := make([]string, 0, len(h.pressedKeys))
	for keyName := range h.pressedKeys {
		keyNames = append(keyNames, keyNameToString(keyName))
	}
	parts = append(parts, keyNames...)

	// Validate combination: require at least one modifier and either a regular key
	// or (if allowed) a supported modifier-only combo like Ctrl+Win/Cmd/Super
	valid := false
	if len(h.modifiers) > 0 {
		if len(h.pressedKeys) > 0 {
			valid = true
		} else if h.allowModifierOnly && isModifierOnlyAllowed(h.modifiers) {
			valid = true
		}
	}
	if !valid {
		h.entry.SetText("❌ Invalid combination (need modifier + key)")
		return false
	}

	hotkeyStr := strings.Join(parts, "+")

	// Check if any sibling has the same keybinding (excluding self)
	for _, sibling := range h.siblings {
		// Skip if sibling is self
		if sibling == h {
			continue
		}
		siblingValue, _ := sibling.binding.Get()
		if siblingValue != "" && siblingValue == hotkeyStr {
			h.entry.SetText("❌ Duplicate! '" + hotkeyStr + "' is already used")
			return false
		}
	}

	// Save to binding
	h.binding.Set(hotkeyStr)
	if h.onChanged != nil {
		h.onChanged()
	}
	return true
}

// clearHotkey clears the current hotkey
func (h *HotkeyCapture) clearHotkey() {
	h.binding.Set("")
	if h.onChanged != nil {
		h.onChanged()
	}
	h.displayLabel.SetText("Click 'Capture' to set hotkey")
}

// StopCapture stops capture if currently capturing (public method for external use)
func (h *HotkeyCapture) StopCapture() {
	h.mu.Lock()
	isCapturing := h.isCapturing
	h.mu.Unlock()

	if isCapturing {
		h.stopCapture()
	}
}

// UpdateFromBinding updates the display from the binding value
func (h *HotkeyCapture) UpdateFromBinding() {
	currentValue, _ := h.binding.Get()
	if currentValue != "" {
		h.displayLabel.SetText(currentValue)
	} else {
		h.displayLabel.SetText("Click 'Capture' to set hotkey")
	}
}

// SetSiblings sets other capture widgets that should be disabled during capture
func (h *HotkeyCapture) SetSiblings(siblings ...*HotkeyCapture) {
	h.siblings = siblings
}

// SetAllowModifierOnly allows accepting modifier-only combinations (e.g., Ctrl+Win)
func (h *HotkeyCapture) SetAllowModifierOnly(allow bool) {
	h.allowModifierOnly = allow
}

// Helper functions

// isModifierKey checks if the key is a modifier key
func isModifierKey(key fyne.KeyName) bool {
	return key == desktop.KeyShiftLeft || key == desktop.KeyShiftRight ||
		key == desktop.KeyControlLeft || key == desktop.KeyControlRight ||
		key == desktop.KeyAltLeft || key == desktop.KeyAltRight ||
		key == desktop.KeySuperLeft || key == desktop.KeySuperRight
}

// isModifierOnlyAllowed returns true if modifiers contain Ctrl and Super (Win/Cmd/Super)
func isModifierOnlyAllowed(mods map[fyne.KeyModifier]bool) bool {
	return mods[fyne.KeyModifierControl] && mods[fyne.KeyModifierSuper]
}

// keyNameToString converts Fyne KeyName to string representation
func keyNameToString(key fyne.KeyName) string {
	// Convert key name to lowercase
	keyStr := strings.ToLower(string(key))

	// Handle special keys
	switch key {
	case fyne.KeySpace:
		return "space"
	case fyne.KeyEscape:
		return "escape" // Match FFI expectation
	case fyne.KeyReturn, fyne.KeyEnter:
		return "return" // Match FFI expectation
	case fyne.KeyTab:
		return "tab"
	case fyne.KeyBackspace:
		return "backspace"
	case fyne.KeyDelete:
		return "delete"
	case fyne.KeyUp:
		return "up"
	case fyne.KeyDown:
		return "down"
	case fyne.KeyLeft:
		return "left"
	case fyne.KeyRight:
		return "right"
	default:
		// For regular keys, just return lowercase version
		if len(keyStr) == 1 {
			return keyStr
		}
		return keyStr
	}
}

// getAltName returns platform-specific alt key name
func getAltName() string {
	switch runtime.GOOS {
	case "darwin":
		return "option"
	default:
		return "alt"
	}
}

// getSuperName returns platform-specific super key name
func getSuperName() string {
	switch runtime.GOOS {
	case "darwin":
		return "cmd"
	case "windows":
		return "win"
	default:
		return "super"
	}
}
