package config

import (
	"testing"

	"github.com/paradoxe35/encre/internal/prompt"
)

func TestDefaultActionsCoverEveryKind(t *testing.T) {
	actions := DefaultActions()

	for _, kind := range ActionOrder {
		action, ok := actions[kind]
		if !ok {
			t.Fatalf("%s missing from DefaultActions", kind)
		}
		if action.Hotkey == "" {
			t.Errorf("%s has no default hotkey", kind)
		}
		if action.CharacterLimit == 0 || action.TimeoutSeconds == 0 {
			t.Errorf("%s has zero limit or timeout", kind)
		}
	}
}

func TestDefaultHotkeysAreUnique(t *testing.T) {
	seen := make(map[string]ActionKind)
	for kind, action := range DefaultActions() {
		if other, clash := seen[action.Hotkey]; clash {
			t.Errorf("%s and %s share the default hotkey %q", kind, other, action.Hotkey)
		}
		seen[action.Hotkey] = kind
	}
}

func TestDictateDefaultsToPushToTalkAndDisabled(t *testing.T) {
	dictate := DefaultActions()[ActionDictate]

	if dictate.Enabled {
		t.Error("dictate should be off until speech-to-text ships")
	}
	if !dictate.PushToTalk {
		t.Error("dictate should default to push-to-talk")
	}
}

func TestPromptOrDefaultFallsBackPerKind(t *testing.T) {
	blank := ActionConfig{}

	if got := blank.PromptOrDefault(ActionTranslate); got != prompt.Translate {
		t.Error("translate did not fall back to the translate prompt")
	}
	if got := blank.PromptOrDefault(ActionReviseSelection); got != prompt.Revise {
		t.Error("revise did not fall back to the revise prompt")
	}

	custom := ActionConfig{SystemPrompt: "mine"}
	if got := custom.PromptOrDefault(ActionTranslate); got != "mine" {
		t.Errorf("PromptOrDefault overrode a custom prompt with %q", got)
	}
}

func TestActionFillsZeroedLimits(t *testing.T) {
	cfg := &Config{Actions: map[ActionKind]ActionConfig{
		ActionTranslate: {Hotkey: "ctrl+alt+g"},
	}}

	action := cfg.Action(ActionTranslate)
	if action.CharacterLimit != DefaultCharacterLimit {
		t.Errorf("CharacterLimit = %d, want %d", action.CharacterLimit, DefaultCharacterLimit)
	}
	if action.TimeoutSeconds != DefaultTimeoutSeconds {
		t.Errorf("TimeoutSeconds = %d, want %d", action.TimeoutSeconds, DefaultTimeoutSeconds)
	}
}

func TestActionFallsBackForUnknownKind(t *testing.T) {
	cfg := &Config{Actions: map[ActionKind]ActionConfig{}}

	if got := cfg.Action(ActionReviseAll).Hotkey; got == "" {
		t.Error("a missing action should fall back to its default")
	}
}

func TestProviderForPrefersOverrideThenDefault(t *testing.T) {
	cfg := &Config{
		AIProvider: AIProviderConfig{
			Provider: BuiltInOpenAI,
			Providers: map[string]ProviderSettings{
				BuiltInOpenAI: {},
				BuiltInClaude: {},
			},
		},
		Actions: map[ActionKind]ActionConfig{
			ActionTranslate:       {ProviderID: BuiltInClaude},
			ActionReviseSelection: {},
			ActionReviseAll:       {ProviderID: "deleted-provider"},
		},
	}

	if got := cfg.ProviderFor(ActionTranslate); got != BuiltInClaude {
		t.Errorf("override ignored: got %q", got)
	}
	if got := cfg.ProviderFor(ActionReviseSelection); got != BuiltInOpenAI {
		t.Errorf("empty override should use the default: got %q", got)
	}
	if got := cfg.ProviderFor(ActionReviseAll); got != BuiltInOpenAI {
		t.Errorf("override naming a deleted provider should fall back: got %q", got)
	}
}

func TestApplyDefaultsRepairsPartialConfig(t *testing.T) {
	cfg := &Config{}
	cfg.applyDefaults()

	if len(cfg.Actions) != len(ActionOrder) {
		t.Errorf("Actions has %d entries, want %d", len(cfg.Actions), len(ActionOrder))
	}
	if cfg.Translate.PrimaryLanguage == cfg.Translate.SecondaryLanguage {
		t.Error("language pair must differ")
	}
	if cfg.Appearance.Theme == "" {
		t.Error("theme was left empty")
	}
}

func TestActionKindClassification(t *testing.T) {
	if !ActionReviseAll.SelectsAll() {
		t.Error("revise_all should select the whole field")
	}
	if ActionReviseSelection.SelectsAll() || ActionTranslate.SelectsAll() {
		t.Error("selection-scoped actions must not select all")
	}
	if ActionDictate.UsesAI() {
		t.Error("dictate does not call a text provider")
	}
	if !ActionTranslate.UsesAI() {
		t.Error("translate calls a text provider")
	}
}
