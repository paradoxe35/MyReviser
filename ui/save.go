package ui

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/paradoxe35/encre/internal/config"
	"github.com/paradoxe35/encre/internal/logger"
)

const statusErrorLimit = 60

func (w *MainWindow) saveSettings() {
	if err := w.applyProviderSettings(); err != nil {
		w.reportSaveError("Error", err)
		return
	}

	if err := w.applyActionSettings(); err != nil {
		w.reportSaveError("Error", err)
		return
	}

	w.config.Translate.PrimaryLanguage = w.primaryLanguage.Code()
	w.config.Translate.SecondaryLanguage = w.secondaryLanguage.Code()
	w.config.EnableProviderMentions = w.mentionsCheck.Checked
	w.applySpeechSettings()

	startMinimized, _ := w.startMinimizedBinding.Get()
	startOnLogin, _ := w.startOnLoginBinding.Get()
	themeSetting, _ := w.themeBinding.Get()

	w.config.Appearance.StartMinimized = startMinimized
	w.config.Appearance.StartOnLogin = startOnLogin
	w.config.Appearance.Theme = themeSetting

	w.applyAutoStartSetting(startOnLogin)

	if err := w.config.Save(); err != nil {
		w.reportSaveError("Error saving settings", err)
		return
	}

	w.statusBinding.Set("Settings saved")
	logger.Info("Settings saved")
}

func (w *MainWindow) applyProviderSettings() error {
	provider, _ := w.providerBinding.Get()
	apiKey, _ := w.apiKeyBinding.Get()
	model, _ := w.modelBinding.Get()
	baseURL, _ := w.baseURLBinding.Get()

	existing := w.config.GetProviderSettings(provider)
	isCustom := w.config.IsCustomProvider(provider)

	if isCustom && baseURL == "" {
		return errors.New("base URL is required for custom providers")
	}

	settings := config.ProviderSettings{
		Model:        model,
		Temperature:  existing.Temperature,
		IsCustom:     existing.IsCustom,
		ProviderType: existing.ProviderType,
		NoAPIKey:     existing.NoAPIKey,
		LowReasoning: existing.LowReasoning,
	}
	if isCustom {
		settings.BaseURL = baseURL
	}

	w.config.SetProviderSettings(provider, settings)

	if err := w.config.SaveAPIKey(provider, apiKey); err != nil {
		return err
	}

	w.config.SetCurrentProvider(provider)
	return nil
}

func (w *MainWindow) applyActionSettings() error {
	for _, kind := range config.ActionOrder {
		action := w.config.Action(kind)

		if binding, ok := w.hotkeyBindings[kind]; ok {
			hotkey, _ := binding.Get()
			action.Hotkey = hotkey
		}
		if enable, ok := w.enables[kind]; ok {
			action.Enabled = enable.Checked
		}

		w.config.SetAction(kind, action)
	}

	for op, editor := range w.operationEditors {
		limit, err := strconv.Atoi(editor.limit.Text)
		if err != nil || validateCharacterLimit(editor.limit.Text) != nil {
			return fmt.Errorf("%s: character limit must be between 1 and 100000", op.Label())
		}

		w.config.SetOperation(op, config.OperationConfig{
			SystemPrompt:   editor.prompt.Text,
			CharacterLimit: limit,
			TimeoutSeconds: int(editor.timeout.Value),
			ProviderID:     providerID(editor.provider.Selected),
		})
	}
	return nil
}

func (w *MainWindow) reportSaveError(prefix string, err error) {
	message := err.Error()
	if len(message) > statusErrorLimit {
		message = message[:statusErrorLimit-3] + "..."
	}
	w.statusBinding.Set(prefix + ": " + message)
	logger.Error(prefix, "error", err)
}
