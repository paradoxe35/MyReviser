package prompt

import (
	"strings"
	"testing"
)

func TestRenderTranslateSubstitutesBothLanguages(t *testing.T) {
	got := RenderTranslate(Translate, "English", "French")

	if strings.Contains(got, PlaceholderPrimary) || strings.Contains(got, PlaceholderSecondary) {
		t.Error("placeholders survived rendering")
	}
	if !strings.Contains(got, "English") || !strings.Contains(got, "French") {
		t.Error("rendered prompt is missing a language name")
	}
}

func TestRenderTranslateAppendsWhenTemplateHasNoPlaceholders(t *testing.T) {
	got := RenderTranslate("Translate this.", "English", "French")

	if !strings.HasPrefix(got, "Translate this.") {
		t.Error("custom template was not preserved")
	}
	if !strings.Contains(got, "English") || !strings.Contains(got, "French") {
		t.Errorf("language pair was not appended: %q", got)
	}
}

func TestRenderTranslateHandlesOneSidedTemplate(t *testing.T) {
	got := RenderTranslate("Always use "+PlaceholderPrimary+".", "English", "French")

	if strings.Contains(got, PlaceholderPrimary) {
		t.Error("primary placeholder survived")
	}
	if strings.Contains(got, "Otherwise, translate it into") {
		t.Error("should substitute, not append, when a placeholder is present")
	}
}

func TestBuiltInPromptsForbidPreamble(t *testing.T) {
	for name, text := range map[string]string{
		"Revise":    Revise,
		"Translate": Translate,
		"Dictate":   Dictate,
	} {
		if !strings.Contains(text, "No preamble") {
			t.Errorf("%s prompt does not forbid a preamble", name)
		}
	}
}
