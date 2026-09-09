package input

import "strings"

// Device is a microphone the host can record from.
type Device struct {
	Name      string
	IsDefault bool
}

// parseDevices reads the newline-separated list the FFI returns, where the
// default carries a leading '*'.
func parseDevices(listed string) []Device {
	if strings.TrimSpace(listed) == "" {
		return nil
	}

	lines := strings.Split(listed, "\n")
	devices := make([]Device, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		name, isDefault := strings.CutPrefix(line, "*")
		devices = append(devices, Device{Name: name, IsDefault: isDefault})
	}
	return devices
}
