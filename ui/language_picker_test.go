package ui

import (
	"testing"

	"github.com/paradoxe35/encre/internal/language"
)

func TestLanguagePickerTracksValidSelection(t *testing.T) {
	picker := NewLanguagePicker("en")
	var changed string
	picker.SetOnCodeChanged(func(code string) { changed = code })

	picker.SetSelected(labelFor(language.Find("fr")))
	picker.OnChanged(picker.Selected)

	if got := picker.Code(); got != "fr" {
		t.Fatalf("Code() = %q, want fr", got)
	}
	if changed != "fr" {
		t.Fatalf("onCodeChanged = %q, want fr", changed)
	}
}

func TestLanguagePickerExcludesOtherLanguage(t *testing.T) {
	picker := NewLanguagePicker("en")
	picker.SetExcludedCode("fr")
	picker.SetSelected(labelFor(language.Find("fr")))
	picker.OnChanged(picker.Selected)

	if got := picker.Code(); got == language.Find("fr").Code {
		t.Fatalf("excluded language was accepted: %q", got)
	}
}

func TestLanguagePickerChangesDuplicateSelection(t *testing.T) {
	picker := NewLanguagePicker("en")
	picker.SetExcludedCode("en")

	if got := picker.Code(); got == "en" {
		t.Fatal("picker kept a language that was excluded")
	}
}
