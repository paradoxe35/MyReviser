package stt

import (
	"fmt"
	"sort"
	"strings"
)

// LanguageName resolves a catalog language code to its English name, falling
// back to the code itself when unknown.
func LanguageName(code string) string {
	if name, ok := languageNames[strings.ToLower(code)]; ok {
		return name
	}
	return code
}

// languageNames covers the codes the model catalog uses. Not exhaustive: an
// unknown code displays as-is rather than hiding the model's capability.
var languageNames = map[string]string{
	"en": "English", "fr": "French", "es": "Spanish", "de": "German",
	"it": "Italian", "pt": "Portuguese", "nl": "Dutch", "pl": "Polish",
	"ru": "Russian", "uk": "Ukrainian", "cs": "Czech", "sk": "Slovak",
	"hu": "Hungarian", "ro": "Romanian", "bg": "Bulgarian", "el": "Greek",
	"tr": "Turkish", "ar": "Arabic", "he": "Hebrew", "fa": "Persian",
	"ur": "Urdu", "hi": "Hindi", "bn": "Bengali", "ta": "Tamil",
	"te": "Telugu", "mr": "Marathi", "gu": "Gujarati", "pa": "Punjabi",
	"th": "Thai", "vi": "Vietnamese", "id": "Indonesian", "ms": "Malay",
	"tl": "Filipino", "zh": "Chinese", "ja": "Japanese", "ko": "Korean",
	"sv": "Swedish", "da": "Danish", "nb": "Norwegian", "fi": "Finnish",
	"is": "Icelandic", "ca": "Catalan", "eu": "Basque", "gl": "Galician",
	"cy": "Welsh", "ga": "Irish", "af": "Afrikaans", "sw": "Swahili",
	"am": "Amharic", "ha": "Hausa", "yo": "Yoruba", "ig": "Igbo",
	"zu": "Zulu", "rw": "Kinyarwanda", "ln": "Lingala", "mg": "Malagasy",
	"hy": "Armenian", "ka": "Georgian", "az": "Azerbaijani", "kk": "Kazakh",
	"uz": "Uzbek", "mn": "Mongolian", "ne": "Nepali", "si": "Sinhala",
	"km": "Khmer", "lo": "Lao", "my": "Burmese", "sq": "Albanian",
	"sr": "Serbian", "hr": "Croatian", "sl": "Slovenian", "mk": "Macedonian",
	"bs": "Bosnian", "et": "Estonian", "lv": "Latvian", "lt": "Lithuanian",
	"ml": "Malayalam", "kn": "Kannada", "or": "Odia", "so": "Somali",
	"om": "Oromo", "jv": "Javanese", "su": "Sundanese", "eo": "Esperanto",
}

// SpeedLabel describes how a model is expected to keep up on this machine.
// The thresholds match the row summaries: comfortable runs are "fast", ten
// times real-time or better is "very fast".
func SpeedLabel(model Model, host Machine) string {
	switch model.Fit(host) {
	case FitTooLarge:
		return "too large for this machine"
	case FitSlow:
		return "slow here"
	default:
		if model.EstimatedRealtime(host) >= 10 {
			return "very fast here"
		}
		return "fast here"
	}
}

// ModelDetails renders the human-facing facts about a model: what it speaks,
// what it costs to run, and how it is expected to behave on this machine.
// Deliberately omits the catalog description and raw benchmark numbers.
func ModelDetails(model Model, host Machine, downloaded bool) string {
	var lines []string

	if len(model.Languages) == 0 {
		lines = append(lines, "Languages: unknown")
	} else {
		names := make([]string, 0, len(model.Languages))
		for _, code := range model.Languages {
			names = append(names, LanguageName(code))
		}
		sort.Strings(names)
		lines = append(lines, fmt.Sprintf("Languages (%d): %s",
			len(names), strings.Join(names, ", ")))
	}

	lines = append(lines, fmt.Sprintf("Size: %.0f MB", model.SizeMB()))
	lines = append(lines, fmt.Sprintf("Speed: %s", SpeedLabel(model, host)))
	lines = append(lines, streamingLine(model))
	if model.License != "" {
		lines = append(lines, "License: "+model.License)
	}
	if model.WordErrorRate > 0 {
		lines = append(lines, fmt.Sprintf("Word error rate: %.1f%%", model.WordErrorRate*100))
	}

	state := "Not downloaded"
	switch {
	case downloaded:
		state = "Downloaded"
	case model.SizeBytes > 0:
		state = fmt.Sprintf("Not downloaded (%.0f MB)", model.SizeMB())
	}
	lines = append(lines, "State: "+state)

	return strings.Join(lines, "\n")
}

func streamingLine(model Model) string {
	if model.Streaming {
		return "Streaming: supported — text appears while you speak"
	}
	return "Streaming: not supported — text appears when you stop"
}
