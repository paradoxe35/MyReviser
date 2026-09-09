package ui

import (
	"strings"

	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/encre/internal/language"
)

// LanguagePicker is a text entry that filters the language registry as you
// type. Fyne has no searchable Select, and a flat list of 78 entries in a
// dropdown is still searchable through the entry field.
type LanguagePicker struct {
	*widget.SelectEntry
	code          string
	excludedCode  string
	onCodeChanged func(string)
}

func NewLanguagePicker(code string) *LanguagePicker {
	picker := &LanguagePicker{code: code}
	picker.SelectEntry = widget.NewSelectEntry(labelsFor(language.All()))
	picker.SetText(labelFor(language.Find(code)))

	picker.OnChanged = func(text string) {
		picker.refreshOptions(text)
		if match, ok := matchLanguage(text); ok &&
			!strings.EqualFold(match.Code, picker.excludedCode) {
			picker.code = match.Code
			if picker.onCodeChanged != nil {
				picker.onCodeChanged(match.Code)
			}
		}
	}

	return picker
}

// Code returns the selected language, falling back to the last valid choice so
// a half-typed query never clears the setting.
func (p *LanguagePicker) Code() string {
	if match, ok := matchLanguage(p.Text); ok &&
		!strings.EqualFold(match.Code, p.excludedCode) {
		return match.Code
	}
	return p.code
}

func (p *LanguagePicker) SetCode(code string) {
	p.code = code
	p.SetText(labelFor(language.Find(code)))
	p.refreshOptions(p.Text)
}

// SetExcludedCode removes the other side's selected language from this
// picker's choices, so a translation pair cannot be configured identically.
func (p *LanguagePicker) SetExcludedCode(code string) {
	p.excludedCode = code
	if strings.EqualFold(p.code, code) {
		for _, candidate := range language.All() {
			if !strings.EqualFold(candidate.Code, code) {
				p.SetCode(candidate.Code)
				break
			}
		}
	}
	p.refreshOptions(p.Text)
}

func (p *LanguagePicker) SetOnCodeChanged(callback func(string)) {
	p.onCodeChanged = callback
}

func (p *LanguagePicker) refreshOptions(query string) {
	matches := language.Search(query)
	filtered := make([]language.Language, 0, len(matches))
	for _, candidate := range matches {
		if !strings.EqualFold(candidate.Code, p.excludedCode) {
			filtered = append(filtered, candidate)
		}
	}
	p.SetOptions(labelsFor(filtered))
}

func labelFor(l language.Language) string {
	endonym := l.Endonym()
	if endonym == "" || endonym == l.Name {
		return l.Name + " (" + l.Code + ")"
	}
	return l.Name + " — " + endonym + " (" + l.Code + ")"
}

func labelsFor(languages []language.Language) []string {
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
