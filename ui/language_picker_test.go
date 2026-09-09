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

func TestLanguagePairCanSwapWithExclusions(t *testing.T) {
	primary := NewLanguagePicker("fr")
	secondary := NewLanguagePicker("en")
	updateExclusions := func() {
		primary.SetExcludedCode(secondary.Code())
		secondary.SetExcludedCode(primary.Code())
	}
	updateExclusions()

	first, second := primary.Code(), secondary.Code()
	primary.SetExcludedCode("")
	secondary.SetExcludedCode("")
	primary.SetCode(second)
	secondary.SetCode(first)
	updateExclusions()

	if primary.Code() != "en" || secondary.Code() != "fr" {
		t.Fatalf("swapped pair = %q/%q, want en/fr", primary.Code(), secondary.Code())
	}
}
