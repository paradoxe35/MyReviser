package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/encre/internal/input"
)

const systemDefaultDevice = "System default"

// MicrophonePicker chooses the capture device. A device saved earlier but not
// present now stays listed and selected, so unplugging a headset does not
// silently reassign the setting.
type MicrophonePicker struct {
	widget.BaseWidget

	selector  *widget.Select
	status    *widget.Label
	saved     string
	onChanged func()

	// Injected so the empty and unplugged cases can be tested without hardware.
	list func() []input.Device
}

func NewMicrophonePicker(saved string) *MicrophonePicker {
	return newMicrophonePicker(saved, input.InputDevices)
}

func newMicrophonePicker(saved string, list func() []input.Device) *MicrophonePicker {
	p := &MicrophonePicker{}
	p.selector = widget.NewSelect(nil, func(string) {
		if p.onChanged != nil {
			p.onChanged()
		}
	})
	*p = MicrophonePicker{
		selector: p.selector,
		status:   widget.NewLabel(""),
		saved:    saved,
		list:     list,
	}
	p.status.TextStyle.Italic = true

	p.ExtendBaseWidget(p)
	p.Refresh()
	return p
}

// Device returns the chosen name, empty for the system default.
func (p *MicrophonePicker) Device() string {
	if p.selector.Selected == systemDefaultDevice {
		return ""
	}
	return p.selector.Selected
}

// Refresh re-reads the device list, keeping the current choice if it survives.
func (p *MicrophonePicker) Refresh() {
	devices := p.list()

	options := []string{systemDefaultDevice}
	present := false
	for _, device := range devices {
		options = append(options, device.Name)
		if device.Name == p.saved {
			present = true
		}
	}

	// Keep a missing device visible rather than dropping the user's choice.
	if p.saved != "" && !present {
		options = append(options, p.saved)
	}

	p.selector.Options = options
	p.selector.SetSelected(p.selection())
	p.describe(devices, present)

	p.BaseWidget.Refresh()
}

func (p *MicrophonePicker) selection() string {
	if p.saved == "" {
		return systemDefaultDevice
	}
	return p.saved
}

func (p *MicrophonePicker) describe(devices []input.Device, present bool) {
	switch {
	case len(devices) == 0:
		p.status.SetText("No microphone found. Connect one and press Rescan.")
	case p.saved != "" && !present:
		p.status.SetText(p.saved + " is not connected. The default will be used until it returns.")
	default:
		p.status.SetText(defaultName(devices))
	}
}

func defaultName(devices []input.Device) string {
	for _, device := range devices {
		if device.IsDefault {
			return "System default is " + device.Name
		}
	}
	return ""
}

func (p *MicrophonePicker) CreateRenderer() fyne.WidgetRenderer {
	rescan := widget.NewButton("Rescan", func() {
		p.saved = p.Device()
		p.Refresh()
	})

	return widget.NewSimpleRenderer(container.NewVBox(
		container.NewBorder(nil, nil, nil, rescan, p.selector),
		p.status,
	))
}
