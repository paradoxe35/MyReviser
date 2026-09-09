package stt

// RemotePreset is a hosted transcription endpoint. Only OpenAI-shaped services
// are listed: they all take the same multipart POST, so one client serves them.
type RemotePreset struct {
	ID      string
	Name    string
	BaseURL string
	Models  []string
	KeyHint string
}

var RemotePresets = []RemotePreset{
	{
		ID:      "openai",
		Name:    "OpenAI",
		BaseURL: "https://api.openai.com/v1",
		Models:  []string{"gpt-4o-transcribe", "gpt-4o-mini-transcribe", "whisper-1"},
		KeyHint: "platform.openai.com",
	},
	{
		ID:      "groq",
		Name:    "Groq",
		BaseURL: "https://api.groq.com/openai/v1",
		Models:  []string{"whisper-large-v3-turbo", "whisper-large-v3"},
		KeyHint: "console.groq.com",
	},
	{
		ID:      "custom",
		Name:    "Custom",
		BaseURL: "",
		Models:  nil,
		KeyHint: "any OpenAI-compatible endpoint",
	},
}

func FindPreset(id string) (RemotePreset, bool) {
	for _, preset := range RemotePresets {
		if preset.ID == id {
			return preset, true
		}
	}
	return RemotePreset{}, false
}

func PresetNames() []string {
	names := make([]string, len(RemotePresets))
	for i, preset := range RemotePresets {
		names[i] = preset.Name
	}
	return names
}

func PresetByName(name string) (RemotePreset, bool) {
	for _, preset := range RemotePresets {
		if preset.Name == name {
			return preset, true
		}
	}
	return RemotePreset{}, false
}
