package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	locale "github.com/jeandeaual/go-locale"
	"github.com/paradoxe35/scribe/internal/language"
	"github.com/paradoxe35/scribe/internal/utils"
)

const (
	ProviderTypeOpenAICompatible = "openai-compatible"

	BuiltInOpenAI = "openai"
	BuiltInClaude = "claude"
	BuiltInGemini = "gemini"

	DefaultCharacterLimit = 1000
	DefaultTimeoutSeconds = 30
)

type Config struct {
	mu         sync.RWMutex
	AIProvider AIProviderConfig            `json:"ai_provider"`
	Actions    map[ActionKind]ActionConfig `json:"actions"`
	Translate  TranslateConfig             `json:"translate"`
	Appearance AppearanceConfig            `json:"appearance"`
	Meta       MetaConfig                  `json:"meta"`

	// EnableProviderMentions lets a selection opt into a provider by starting
	// with "@name". Applies to every AI-backed action.
	EnableProviderMentions bool `json:"enable_provider_mentions"`
}

type ProviderSettings struct {
	APIKey       string  `json:"api_key"`
	BaseURL      string  `json:"base_url,omitempty"`
	Model        string  `json:"model,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	IsCustom     bool    `json:"is_custom,omitempty"`
	ProviderType string  `json:"provider_type,omitempty"`
	// NoAPIKey suits a model running on this machine. Absent means a key is required, so every
	// configuration written before this existed keeps its meaning.
	NoAPIKey bool `json:"no_api_key,omitempty"`
	// LowReasoning asks a reasoning model to think less. Off by default: a model that does not
	// reason rejects the parameter, and the retry that recovers from it costs a round trip.
	LowReasoning bool `json:"low_reasoning,omitempty"`
}

func (s ProviderSettings) RequiresAPIKey() bool {
	return !s.NoAPIKey
}

type AIProviderConfig struct {
	Provider  string                      `json:"provider"` // "openai" | "claude" | "gemini"
	Providers map[string]ProviderSettings `json:"providers"`
}

type TranslateConfig struct {
	PrimaryLanguage   string `json:"primary_language"`
	SecondaryLanguage string `json:"secondary_language"`
}

type AppearanceConfig struct {
	Theme          string `json:"theme"` // "auto" | "light" | "dark"
	StartMinimized bool   `json:"start_minimized"`
	StartOnLogin   bool   `json:"start_on_login"`
}

type MetaConfig struct {
	FirstRun bool `json:"first_run"`
}

var (
	currentConfig *Config
	configMutex   sync.RWMutex
	listeners     []func(*Config)
	listenerMutex sync.RWMutex
)

const APP_ID = "me.pngwasi.scribe"

// ConfigPath returns the path to the configuration file
func ConfigPath() string {
	return utils.AppHomeDir("config.json")
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		AIProvider: AIProviderConfig{
			Provider: "openai",
			Providers: map[string]ProviderSettings{
				"openai": {
					BaseURL:     "https://api.openai.com/v1",
					Model:       "gpt-4o",
					Temperature: 1.0,
				},
				"claude": {
					BaseURL:     "https://api.anthropic.com",
					Model:       "claude-3-5-haiku-20241022",
					Temperature: 1.0,
				},
				"gemini": {
					BaseURL:     "https://generativelanguage.googleapis.com",
					Model:       "gemini-2.5-flash",
					Temperature: 1.0,
				},
			},
		},
		Actions:                DefaultActions(),
		Translate:              defaultTranslate(),
		Appearance:             defaultAppearance(),
		Meta:                   MetaConfig{FirstRun: true},
		EnableProviderMentions: true,
	}
}

func defaultAppearance() AppearanceConfig {
	return AppearanceConfig{
		Theme:          "auto",
		StartMinimized: false,
		StartOnLogin:   false,
	}
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Ensure config directory exists
	configDir := filepath.Dir(ConfigPath())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(ConfigPath()); os.IsNotExist(err) {
		// Create default config
		cfg := Default()
		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		currentConfig = cfg
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.applyDefaults()

	if cfg.AIProvider.Providers != nil {
		for name, settings := range cfg.AIProvider.Providers {
			if settings.Temperature == 0 {
				settings.Temperature = 1.0
				cfg.AIProvider.Providers[name] = settings
			}
		}
	}

	currentConfig = cfg
	return cfg, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	c.mu.Lock()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(ConfigPath(), data, 0644); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to write config file: %w", err)
	}
	c.mu.Unlock()

	// Update global config and notify listeners
	configMutex.Lock()
	currentConfig = c
	configMutex.Unlock()

	notifyListeners(c)

	return nil
}

// Get returns the current configuration
func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return currentConfig
}

// Update mutates the live configuration and persists it. Save takes the same
// package mutex, so the lock is released before calling it.
func Update(fn func(*Config)) error {
	configMutex.RLock()
	cfg := currentConfig
	configMutex.RUnlock()

	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	fn(cfg)
	return cfg.Save()
}

func RegisterListener(listener func(*Config)) {
	listenerMutex.Lock()
	defer listenerMutex.Unlock()
	listeners = append(listeners, listener)
}

func notifyListeners(cfg *Config) {
	listenerMutex.RLock()
	defer listenerMutex.RUnlock()

	for _, listener := range listeners {
		go listener(cfg)
	}
}

// GetProviderSettings returns settings for a specific provider
func (c *Config) GetProviderSettings(provider string) ProviderSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.AIProvider.Providers == nil {
		return ProviderSettings{}
	}

	settings, ok := c.AIProvider.Providers[provider]
	if !ok {
		// Return defaults for this provider
		defaults := Default()
		if defaultSettings, ok := defaults.AIProvider.Providers[provider]; ok {
			return defaultSettings
		}
		return ProviderSettings{}
	}

	if settings.Temperature == 0 {
		settings.Temperature = 1.0
	}

	return settings
}

// SetProviderSettings updates settings for a specific provider
func (c *Config) SetProviderSettings(provider string, settings ProviderSettings) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.AIProvider.Providers == nil {
		c.AIProvider.Providers = make(map[string]ProviderSettings)
	}

	if settings.Temperature == 0 {
		settings.Temperature = 1.0
	}

	c.AIProvider.Providers[provider] = settings
}

// GetCurrentProvider returns the currently selected provider name
func (c *Config) GetCurrentProvider() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AIProvider.Provider
}

// SetCurrentProvider sets the currently selected provider
func (c *Config) SetCurrentProvider(provider string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AIProvider.Provider = provider
}

func BuiltInProviders() []string {
	return []string{BuiltInOpenAI, BuiltInClaude, BuiltInGemini}
}

func IsBuiltInProvider(name string) bool {
	nameLower := strings.ToLower(name)
	for _, p := range BuiltInProviders() {
		if strings.ToLower(p) == nameLower {
			return true
		}
	}
	return false
}

func (c *Config) GetAllProviderNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0)
	for _, name := range BuiltInProviders() {
		names = append(names, name)
	}

	customNames := make([]string, 0)
	for name, settings := range c.AIProvider.Providers {
		if settings.IsCustom {
			customNames = append(customNames, name)
		}
	}
	sort.Strings(customNames)
	names = append(names, customNames...)
	return names
}

func isValidProviderName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func (c *Config) AddCustomProvider(name string, settings ProviderSettings) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if !isValidProviderName(name) {
		return fmt.Errorf("provider name can only contain letters, numbers, hyphens, and underscores")
	}
	if IsBuiltInProvider(name) {
		return fmt.Errorf("cannot use built-in provider name: %s", name)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.AIProvider.Providers == nil {
		c.AIProvider.Providers = make(map[string]ProviderSettings)
	}

	nameLower := strings.ToLower(name)
	for existingName := range c.AIProvider.Providers {
		if strings.ToLower(existingName) == nameLower {
			return fmt.Errorf("provider with name '%s' already exists", existingName)
		}
	}

	settings.IsCustom = true
	c.AIProvider.Providers[name] = settings
	return nil
}

func (c *Config) DeleteCustomProvider(name string) error {
	if IsBuiltInProvider(name) {
		return fmt.Errorf("cannot delete built-in provider: %s", name)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	settings, exists := c.AIProvider.Providers[name]
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}
	if !settings.IsCustom {
		return fmt.Errorf("cannot delete non-custom provider: %s", name)
	}

	delete(c.AIProvider.Providers, name)

	if c.AIProvider.Provider == name {
		c.AIProvider.Provider = BuiltInOpenAI
	}

	return nil
}

func (c *Config) IsCustomProvider(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if settings, exists := c.AIProvider.Providers[name]; exists {
		return settings.IsCustom
	}
	return false
}

func defaultTranslate() TranslateConfig {
	primary, secondary := language.Defaults(osLocale())
	return TranslateConfig{PrimaryLanguage: primary, SecondaryLanguage: secondary}
}

func osLocale() string {
	tag, err := locale.GetLocale()
	if err != nil {
		return "en"
	}
	return strings.ReplaceAll(tag, "_", "-")
}

// applyDefaults fills anything a hand-edited config left out, so a partial file
// still starts rather than booting with zero-valued hotkeys and limits.
func (c *Config) applyDefaults() {
	if c.Actions == nil {
		c.Actions = DefaultActions()
	} else {
		defaults := DefaultActions()
		for _, kind := range ActionOrder {
			action, ok := c.Actions[kind]
			if !ok {
				c.Actions[kind] = defaults[kind]
				continue
			}
			if action.CharacterLimit == 0 {
				action.CharacterLimit = DefaultCharacterLimit
			}
			if action.TimeoutSeconds == 0 {
				action.TimeoutSeconds = DefaultTimeoutSeconds
			}
			if action.Hotkey == "" {
				action.Hotkey = defaults[kind].Hotkey
			}
			c.Actions[kind] = action
		}
	}

	if c.Translate.PrimaryLanguage == "" || c.Translate.SecondaryLanguage == "" {
		c.Translate = defaultTranslate()
	}
	if c.Appearance.Theme == "" {
		c.Appearance.Theme = defaultAppearance().Theme
	}
}
