package config

import "testing"

func TestSpeechCleanupDefaultsOff(t *testing.T) {
	if defaultSpeech().CleanUp {
		t.Fatal("transcript cleanup should be opt-in")
	}
}

func TestAppearanceStartsVisibleByDefault(t *testing.T) {
	if defaultAppearance().StartMinimized {
		t.Fatal("the application should start visible by default")
	}
}
