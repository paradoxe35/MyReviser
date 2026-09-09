package language

import (
	"sort"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// Language pairs an IETF tag with its English name. Name is what prompts use:
// the model should see a stable label regardless of the user's UI locale.
type Language struct {
	Code string
	Name string
}

var all = []Language{
	{Code: "fr", Name: "French"},
	{Code: "en", Name: "English"},
	{Code: "es", Name: "Spanish"},
	{Code: "de", Name: "German"},
	{Code: "it", Name: "Italian"},
	{Code: "pt", Name: "Portuguese"},
	{Code: "pt-BR", Name: "Portuguese (Brazil)"},
	{Code: "nl", Name: "Dutch"},
	{Code: "pl", Name: "Polish"},
	{Code: "ro", Name: "Romanian"},
	{Code: "sv", Name: "Swedish"},
	{Code: "da", Name: "Danish"},
	{Code: "nb", Name: "Norwegian"},
	{Code: "fi", Name: "Finnish"},
	{Code: "is", Name: "Icelandic"},
	{Code: "cs", Name: "Czech"},
	{Code: "sk", Name: "Slovak"},
	{Code: "hu", Name: "Hungarian"},
	{Code: "el", Name: "Greek"},
	{Code: "bg", Name: "Bulgarian"},
	{Code: "uk", Name: "Ukrainian"},
	{Code: "ru", Name: "Russian"},
	{Code: "sr", Name: "Serbian"},
	{Code: "hr", Name: "Croatian"},
	{Code: "sl", Name: "Slovenian"},
	{Code: "lt", Name: "Lithuanian"},
	{Code: "lv", Name: "Latvian"},
	{Code: "et", Name: "Estonian"},
	{Code: "tr", Name: "Turkish"},
	{Code: "ar", Name: "Arabic"},
	{Code: "he", Name: "Hebrew"},
	{Code: "fa", Name: "Persian"},
	{Code: "ur", Name: "Urdu"},
	{Code: "hi", Name: "Hindi"},
	{Code: "bn", Name: "Bengali"},
	{Code: "ta", Name: "Tamil"},
	{Code: "te", Name: "Telugu"},
	{Code: "mr", Name: "Marathi"},
	{Code: "gu", Name: "Gujarati"},
	{Code: "pa", Name: "Punjabi"},
	{Code: "th", Name: "Thai"},
	{Code: "vi", Name: "Vietnamese"},
	{Code: "id", Name: "Indonesian"},
	{Code: "ms", Name: "Malay"},
	{Code: "tl", Name: "Filipino"},
	{Code: "zh-Hans", Name: "Chinese (Simplified)"},
	{Code: "zh-Hant", Name: "Chinese (Traditional)"},
	{Code: "ja", Name: "Japanese"},
	{Code: "ko", Name: "Korean"},
	{Code: "sw", Name: "Swahili"},
	{Code: "am", Name: "Amharic"},
	{Code: "ha", Name: "Hausa"},
	{Code: "yo", Name: "Yoruba"},
	{Code: "ig", Name: "Igbo"},
	{Code: "zu", Name: "Zulu"},
	{Code: "af", Name: "Afrikaans"},
	{Code: "mg", Name: "Malagasy"},
	{Code: "rw", Name: "Kinyarwanda"},
	{Code: "ln", Name: "Lingala"},
	{Code: "ca", Name: "Catalan"},
	{Code: "eu", Name: "Basque"},
	{Code: "gl", Name: "Galician"},
	{Code: "cy", Name: "Welsh"},
	{Code: "ga", Name: "Irish"},
	{Code: "sq", Name: "Albanian"},
	{Code: "hy", Name: "Armenian"},
	{Code: "ka", Name: "Georgian"},
	{Code: "az", Name: "Azerbaijani"},
	{Code: "kk", Name: "Kazakh"},
	{Code: "uz", Name: "Uzbek"},
	{Code: "mn", Name: "Mongolian"},
	{Code: "ne", Name: "Nepali"},
	{Code: "si", Name: "Sinhala"},
	{Code: "km", Name: "Khmer"},
	{Code: "lo", Name: "Lao"},
	{Code: "my", Name: "Burmese"},
	{Code: "eo", Name: "Esperanto"},
	{Code: "la", Name: "Latin"},
}

var byCode = func() map[string]Language {
	index := make(map[string]Language, len(all))
	for _, l := range all {
		index[strings.ToLower(l.Code)] = l
	}
	return index
}()

func All() []Language { return all }

// Find returns the registry entry for a tag, or a synthesised one so an
// unknown code still renders instead of vanishing.
func Find(code string) Language {
	if l, ok := byCode[strings.ToLower(strings.TrimSpace(code))]; ok {
		return l
	}
	trimmed := strings.TrimSpace(code)
	return Language{Code: trimmed, Name: trimmed}
}

func IsKnown(code string) bool {
	_, ok := byCode[strings.ToLower(strings.TrimSpace(code))]
	return ok
}

// Endonym returns the language's name in its own language, falling back to
// the English name when the tag is not one x/text can parse.
func (l Language) Endonym() string {
	tag, err := language.Parse(l.Code)
	if err != nil {
		return l.Name
	}
	if name := display.Self.Name(tag); name != "" {
		return name
	}
	return l.Name
}

// Search filters by code, English name or endonym, preferring prefix matches.
func Search(query string) []Language {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return all
	}

	var prefix, contains []Language
	for _, l := range all {
		code, name, endonym := strings.ToLower(l.Code), strings.ToLower(l.Name), strings.ToLower(l.Endonym())
		switch {
		case strings.HasPrefix(code, q), strings.HasPrefix(name, q), strings.HasPrefix(endonym, q):
			prefix = append(prefix, l)
		case strings.Contains(name, q), strings.Contains(endonym, q):
			contains = append(contains, l)
		}
	}
	return append(prefix, contains...)
}

// Sorted returns the registry ordered by English name, for pickers that want
// alphabetical rather than the curated popularity order of All.
func Sorted() []Language {
	out := append([]Language(nil), all...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Defaults seeds a primary/secondary pair from the OS locale. The secondary is
// whichever of English or French the primary is not, so the pair is never
// degenerate.
func Defaults(localeTag string) (primary, secondary string) {
	primary = "en"
	if IsKnown(localeTag) {
		primary = Find(localeTag).Code
	} else if base, _, found := strings.Cut(localeTag, "-"); found && IsKnown(base) {
		primary = Find(base).Code
	}

	if strings.EqualFold(primary, "en") {
		return primary, "fr"
	}
	return primary, "en"
}
