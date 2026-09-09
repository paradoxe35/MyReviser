package config

import (
	"runtime"

	"github.com/paradoxe35/encre/internal/prompt"
)

type ActionKind string

const (
	ActionReviseSelection ActionKind = "revise_selection"
	ActionReviseAll       ActionKind = "revise_all"
	ActionTranslate       ActionKind = "translate"
	ActionDictate         ActionKind = "dictate"
)

// ActionOrder is the order actions appear in settings and in hotkey capture.
var ActionOrder = []ActionKind{
	ActionReviseSelection,
	ActionReviseAll,
	ActionTranslate,
	ActionDictate,
}

var actionLabels = map[ActionKind]string{
	ActionReviseSelection: "Revise selection",
	ActionReviseAll:       "Revise everything",
	ActionTranslate:       "Translate selection",
	ActionDictate:         "Dictate",
}

func (k ActionKind) Label() string {
	if label, ok := actionLabels[k]; ok {
		return label
	}
	return string(k)
}

// UsesAI reports whether the action sends text to a provider. Dictate does not:
// it records audio and transcribes it locally.
func (k ActionKind) UsesAI() bool { return k != ActionDictate }

// SelectsAll reports whether the action selects the whole field before copying,
// rather than acting on what the user already highlighted.
func (k ActionKind) SelectsAll() bool { return k == ActionReviseAll }

type ActionConfig struct {
	Enabled bool   `json:"enabled"`
	Hotkey  string `json:"hotkey"`

	SystemPrompt   string `json:"system_prompt,omitempty"`
	CharacterLimit int    `json:"character_limit,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`

	// ProviderID overrides the default provider for this action. Empty means
	// "use the default", so switching the default carries every action with it.
	ProviderID string `json:"provider_id,omitempty"`

	// PushToTalk records while the hotkey is held instead of toggling on press.
	PushToTalk bool `json:"push_to_talk,omitempty"`
}

// PromptOrDefault lets a blank prompt mean "use the built-in", so a Reset in
// settings is just clearing the field and defaults keep improving over time.
func (a ActionConfig) PromptOrDefault(kind ActionKind) string {
	if a.SystemPrompt != "" {
		return a.SystemPrompt
	}
	return defaultPrompt(kind)
}

func defaultPrompt(kind ActionKind) string {
	switch kind {
	case ActionTranslate:
		return prompt.Translate
	case ActionDictate:
		return prompt.Dictate
	default:
		return prompt.Revise
	}
}

func defaultHotkeys() map[ActionKind]string {
	switch runtime.GOOS {
	case "darwin":
		return map[ActionKind]string{
			ActionReviseSelection: "ctrl+cmd",
			ActionReviseAll:       "ctrl+option+space",
			ActionTranslate:       "ctrl+option+g",
			ActionDictate:         "ctrl+option+d",
		}
	case "windows":
		return map[ActionKind]string{
			ActionReviseSelection: "ctrl+win",
			ActionReviseAll:       "ctrl+alt+space",
			ActionTranslate:       "ctrl+alt+g",
			ActionDictate:         "ctrl+alt+d",
		}
	default:
		return map[ActionKind]string{
			ActionReviseSelection: "ctrl+super",
			ActionReviseAll:       "ctrl+alt+space",
			ActionTranslate:       "ctrl+alt+g",
			ActionDictate:         "ctrl+alt+d",
		}
	}
}

func DefaultActions() map[ActionKind]ActionConfig {
	hotkeys := defaultHotkeys()
	actions := make(map[ActionKind]ActionConfig, len(ActionOrder))

	for _, kind := range ActionOrder {
		actions[kind] = ActionConfig{
			Enabled:        kind != ActionDictate,
			Hotkey:         hotkeys[kind],
			CharacterLimit: DefaultCharacterLimit,
			TimeoutSeconds: DefaultTimeoutSeconds,
			PushToTalk:     kind == ActionDictate,
		}
	}
	return actions
}

func (c *Config) Action(kind ActionKind) ActionConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.actionLocked(kind)
}

func (c *Config) actionLocked(kind ActionKind) ActionConfig {
	action, ok := c.Actions[kind]
	if !ok {
		return DefaultActions()[kind]
	}
	if action.CharacterLimit == 0 {
		action.CharacterLimit = DefaultCharacterLimit
	}
	if action.TimeoutSeconds == 0 {
		action.TimeoutSeconds = DefaultTimeoutSeconds
	}
	return action
}

func (c *Config) SetAction(kind ActionKind, action ActionConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Actions == nil {
		c.Actions = make(map[ActionKind]ActionConfig, len(ActionOrder))
	}
	c.Actions[kind] = action
}

// ProviderFor resolves an action's provider, falling back to the default when
// the action has no override or names one that no longer exists.
func (c *Config) ProviderFor(kind ActionKind) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if id := c.actionLocked(kind).ProviderID; id != "" {
		if _, ok := c.AIProvider.Providers[id]; ok {
			return id
		}
	}
	return c.AIProvider.Provider
}
