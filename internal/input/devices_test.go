package input

import "testing"

func TestParseDevices(t *testing.T) {
	devices := parseDevices("Built-in Microphone\n*USB Headset\nWebcam")

	if len(devices) != 3 {
		t.Fatalf("got %d devices, want 3", len(devices))
	}
	if devices[1].Name != "USB Headset" || !devices[1].IsDefault {
		t.Errorf("the starred entry should be the default, got %+v", devices[1])
	}
	if devices[0].IsDefault || devices[2].IsDefault {
		t.Error("only one device should be marked default")
	}
	// The marker must not survive into the name, or selection would never match.
	for _, device := range devices {
		if device.Name == "" || device.Name[0] == '*' {
			t.Errorf("name not cleaned: %q", device.Name)
		}
	}
}

func TestParseDevicesHandlesNoMicrophone(t *testing.T) {
	for _, listed := range []string{"", "   ", "\n"} {
		if got := parseDevices(listed); len(got) != 0 {
			t.Errorf("parseDevices(%q) = %v, want empty", listed, got)
		}
	}
}

func TestParseDevicesRemovesDuplicateNames(t *testing.T) {
	devices := parseDevices("Built-in Microphone\nBuilt-in Microphone\n*Built-in Microphone\nUSB Headset")

	if len(devices) != 2 {
		t.Fatalf("got %d devices, want 2: %+v", len(devices), devices)
	}
	if devices[0].Name != "Built-in Microphone" || !devices[0].IsDefault {
		t.Errorf("duplicate default marker was not preserved: %+v", devices[0])
	}
}
