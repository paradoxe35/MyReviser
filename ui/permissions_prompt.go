package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/scribe/internal/permissions"
)

type permissionPrompt struct {
	root                   *fyne.Container
	infoLabel              *widget.Label
	accessibilitySection   fyne.CanvasObject
	inputMonitoringSection fyne.CanvasObject
	dividerAboveList       *widget.Separator
	dividerBelowList       *widget.Separator
	restartRow             fyne.CanvasObject
	restartButton          *widget.Button
}

func newPermissionPrompt() *permissionPrompt {
	title := widget.NewLabelWithStyle("Permissions Required", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	info := widget.NewLabel("Scribe needs the following permissions to function properly.")
	info.Wrapping = fyne.TextWrapWord

	accessibilityButton := newPermissionButton("Grant Access", func() {
		permissions.OpenPreference(permissions.Accessibility)
	})

	accessibilitySection := buildPermissionRow(
		"Accessibility",
		"Required to automate keyboard input and clipboard operations. If Scribe is not in the list, add it with +.",
		accessibilityButton,
	)

	inputMonitoringButton := newPermissionButton("Grant Access", func() {
		permissions.OpenPreference(permissions.InputMonitoring)
	})

	inputMonitoringSection := buildPermissionRow(
		"Input Monitoring",
		"Required to detect global keyboard shortcuts in the background.",
		inputMonitoringButton,
	)

	restartButton := widget.NewButtonWithIcon("Restart Application", theme.MediaReplayIcon(), func() {
		// Will be set by the window
	})
	restartButton.Importance = widget.HighImportance

	// Create restart row with centered button (not full width)
	restartRow := container.NewPadded(
		container.NewCenter(restartButton),
	)
	restartRow.Hide()

	dividerAbove := widget.NewSeparator()
	dividerBelow := widget.NewSeparator()
	dividerBelow.Hide()

	body := container.NewVBox(
		title,
		info,
		widget.NewLabel(""), // spacing
		dividerAbove,
		accessibilitySection,
		inputMonitoringSection,
		dividerBelow,
		restartRow,
	)

	root := container.NewPadded(body)

	return &permissionPrompt{
		root:                   root,
		infoLabel:              info,
		accessibilitySection:   accessibilitySection,
		inputMonitoringSection: inputMonitoringSection,
		dividerAboveList:       dividerAbove,
		dividerBelowList:       dividerBelow,
		restartRow:             restartRow,
		restartButton:          restartButton,
	}
}

func (p *permissionPrompt) canvasObject() fyne.CanvasObject {
	return p.root
}

func (p *permissionPrompt) update(state permissions.State, showRestart bool) {
	pendingAccessibility := !state.AccessibilityGranted
	pendingInput := !state.InputMonitoringGranted
	hasPending := pendingAccessibility || pendingInput

	if pendingAccessibility {
		p.accessibilitySection.Show()
	} else {
		p.accessibilitySection.Hide()
	}

	if pendingInput {
		p.inputMonitoringSection.Show()
	} else {
		p.inputMonitoringSection.Hide()
	}

	switch {
	case hasPending:
		p.infoLabel.SetText("Scribe needs the following permissions to function properly.")
		p.dividerAboveList.Show()
		p.dividerBelowList.Hide()
		p.restartRow.Hide()
		p.root.Show()
	case showRestart:
		p.infoLabel.SetText("All permissions granted! Please restart the application.")
		p.dividerAboveList.Hide()
		p.dividerBelowList.Show()
		p.restartRow.Show()
		p.root.Show()
	default:
		p.restartRow.Hide()
		p.root.Hide()
	}

	p.root.Refresh()
}

func buildPermissionRow(title, description string, button *widget.Button) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	descriptionLabel := widget.NewLabel(description)
	descriptionLabel.Wrapping = fyne.TextWrapWord

	textColumn := container.NewVBox(titleLabel, descriptionLabel)

	// Use HBox with layout that allows text to expand
	row := container.NewBorder(nil, nil, nil, button, textColumn)
	return container.NewPadded(row)
}

func newPermissionButton(label string, tapped func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, theme.SettingsIcon(), tapped)
	btn.Importance = widget.MediumImportance
	btn.IconPlacement = widget.ButtonIconLeadingText
	return btn
}
