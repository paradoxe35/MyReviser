package ui

import (
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/encre/internal/language"
)

const languageSuggestionLimit = 12

// LanguagePicker is a text entry that filters the language registry as you
// type. Fyne has no searchable Select, and a flat list of 78 entries in a
// dropdown is unusable.
type LanguagePicker struct {
	*widget.SelectEntry
	code string
}

func NewLanguagePicker(code string) *LanguagePicker {
	picker := &LanguagePicker{code: code}
	picker.SelectEntry = widget.NewSelectEntry(labelsFor(language.All()))
	picker.SetText(labelFor(language.Find(code)))

	picker.OnChanged = func(text string) {
		picker.SetOptions(labelsFor(language.Search(text)))
		if match, ok := matchLanguage(text); ok {
			picker.code = match.Code
		}
	}

	return picker
}

// Code returns the selected language, falling back to the last valid choice so
// a half-typed query never clears the setting.
func (p *LanguagePicker) Code() string {
	if match, ok := matchLanguage(p.Text); ok {
		return match.Code
	}
	return p.code
}

func (p *LanguagePicker) SetCode(code string) {
	p.code = code
	p.SetText(labelFor(language.Find(code)))
}

func labelFor(l language.Language) string {
	endonym := l.Endonym()
	if endonym == "" || endonym == l.Name {
		return l.Name + " (" + l.Code + ")"
	}
	return l.Name + " — " + endonym + " (" + l.Code + ")"
}

func labelsFor(languages []language.Language) []string {
	if len(languages) > languageSuggestionLimit {
		languages = languages[:languageSuggestionLimit]
	}

	labels := make([]string, len(languages))
	for i, l := range languages {
		labels[i] = labelFor(l)
	}
	return labels
}

// matchLanguage resolves what the user typed, accepting either a full label
// from the dropdown or a bare code or name typed by hand.
func matchLanguage(text string) (language.Language, bool) {
	for _, l := range language.All() {
		if labelFor(l) == text {
			return l, true
		}
	}

	if language.IsKnown(text) {
		return language.Find(text), true
	}

	if matches := language.Search(text); len(matches) == 1 {
		return matches[0], true
	}
	return language.Language{}, false
}
