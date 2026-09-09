package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/encre/internal/config"
	"github.com/paradoxe35/encre/internal/input"
	"github.com/paradoxe35/encre/internal/logger"
	"github.com/paradoxe35/encre/internal/permissions"
	"github.com/paradoxe35/encre/internal/platform"
	"github.com/paradoxe35/encre/internal/stt"
)

const (
	// UI spacing constants
	PaddingSmall  = 5
	PaddingMedium = 10
	PaddingLarge  = 20
)

type MainWindow struct {
	fyne.Window
	app                 fyne.App
	config              *config.Config
	hotkeyManager       *input.FFIHotkeyManager
	permissionPrompt    *permissionPrompt
	rootContainer       *fyne.Container
	mainContent         fyne.CanvasObject
	permissionContainer fyne.CanvasObject

	// Data bindings
	providerBinding       binding.String
	apiKeyBinding         binding.String
	modelBinding          binding.String
	baseURLBinding        binding.String
	statusBinding         binding.String
	startMinimizedBinding binding.Bool
	startOnLoginBinding   binding.Bool
	themeBinding          binding.String

	hotkeyBindings   map[config.ActionKind]binding.String
	operationEditors map[config.Operation]*operationEditor
	captures         map[config.ActionKind]*HotkeyCapture
	enables          map[config.ActionKind]*widget.Check

	primaryLanguage   *LanguagePicker
	secondaryLanguage *LanguagePicker
	mentionsCheck     *widget.Check

	speechModels      *ModelList
	speechStoreRef    *stt.Store
	speechEngine      *widget.Select
	speechKeepLoaded  *widget.Check
	speechCleanUp     *widget.Check
	microphone        *MicrophonePicker
	speechRemote      *widget.Select
	speechRemoteModel *widget.SelectEntry
	speechRemoteURL   *widget.Entry
	speechRemoteKey   *widget.Entry

	// UI containers for dynamic visibility
	baseURLContainer *fyne.Container
	baseURLEntry     *widget.Entry

	// Custom provider UI components
	providerSelect       *widget.Select
	deleteProviderButton *widget.Button

	// Platform-specific callbacks
	onShowCallback func()
	onHideCallback func()
}

func NewMainWindow(app fyne.App, cfg *config.Config, hotkeyManager *input.FFIHotkeyManager) *MainWindow {
	window := newChromelessWindow(app, "Encre")
	window.Resize(fyne.NewSize(windowWidth, windowHeight))
	window.SetFixedSize(true)
	window.CenterOnScreen()
	window.SetIcon(app.Icon())

	prompt := newPermissionPrompt()

	mw := &MainWindow{
		Window:           window,
		app:              app,
		config:           cfg,
		hotkeyManager:    hotkeyManager,
		permissionPrompt: prompt,
	}

	// Set restart button callback
	prompt.restartButton.OnTapped = mw.restartApplication

	// Initialize data bindings
	mw.initBindings()

	// Apply theme from config
	themeName := cfg.Appearance.Theme
	if themeName == "" {
		themeName = "auto"
	}
	mw.applyTheme(themeName)

	// Create and set content containers
	mw.mainContent = mw.createContent()
	mw.permissionContainer = mw.permissionPrompt.canvasObject()
	mw.rootContainer = container.NewStack(mw.mainContent, mw.permissionContainer)
	window.SetContent(mw.rootContainer)
	mw.showMainContent()

	return mw
}

func (w *MainWindow) initBindings() {
	w.providerBinding = binding.NewString()
	w.apiKeyBinding = binding.NewString()
	w.modelBinding = binding.NewString()
	w.baseURLBinding = binding.NewString()
	w.statusBinding = binding.NewString()
	w.startMinimizedBinding = binding.NewBool()
	w.startOnLoginBinding = binding.NewBool()
	w.themeBinding = binding.NewString()

	// Set initial values from config
	currentProvider := w.config.GetCurrentProvider()
	w.providerBinding.Set(currentProvider)

	// Load settings for current provider
	w.loadProviderSettings(currentProvider)

	w.hotkeyBindings = make(map[config.ActionKind]binding.String, len(config.ActionOrder))
	for _, kind := range config.ActionOrder {
		value := binding.NewString()
		value.Set(w.config.Action(kind).Hotkey)
		w.hotkeyBindings[kind] = value
	}

	w.statusBinding.Set("Ready")
	w.startMinimizedBinding.Set(w.config.Appearance.StartMinimized)

	// Sync StartOnLogin with actual system state
	// The user might have removed the login item outside the app
	autoStart := platform.GetAutoStart()
	actualStartOnLogin := autoStart.IsEnabled()
	w.startOnLoginBinding.Set(actualStartOnLogin)

	// Update config if out of sync
	if w.config.Appearance.StartOnLogin != actualStartOnLogin {
		w.config.Appearance.StartOnLogin = actualStartOnLogin
		w.config.Save()
		logger.Info("Synced StartOnLogin with system state", "enabled", actualStartOnLogin)
	}

	// Set theme, default to "auto" if empty
	theme := w.config.Appearance.Theme
	if theme == "" {
		theme = "auto"
	}
	w.themeBinding.Set(theme)
}

// SetPermissionState updates the permission prompt visibility and messaging.
func (w *MainWindow) SetPermissionState(state permissions.State, showRestart bool) {
	if w.permissionPrompt == nil {
		return
	}

	w.permissionPrompt.update(state, showRestart)

	if !state.AllGranted() || showRestart {
		w.showPermissionContent()
	} else {
		w.showMainContent()
	}
}

func (w *MainWindow) showPermissionContent() {
	if w.permissionContainer != nil {
		w.permissionContainer.Show()
	}
	if w.mainContent != nil {
		w.mainContent.Hide()
	}
	if w.rootContainer != nil {
		w.rootContainer.Objects = []fyne.CanvasObject{w.mainContent, w.permissionContainer}
		w.rootContainer.Refresh()
	}
}

func (w *MainWindow) showMainContent() {
	if w.mainContent != nil {
		w.mainContent.Show()
	}
	if w.permissionContainer != nil {
		w.permissionContainer.Hide()
	}
	if w.rootContainer != nil {
		w.rootContainer.Objects = []fyne.CanvasObject{w.permissionContainer, w.mainContent}
		w.rootContainer.Refresh()
	}
}

func (w *MainWindow) createContent() fyne.CanvasObject {
	statusBar := w.createStatusBar()

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("AI", theme.ComputerIcon(), w.createProviderSection()),
		container.NewTabItemWithIcon("Hotkeys", keyboardIcon, w.createHotkeysSection()),
		container.NewTabItemWithIcon("Actions", theme.DocumentIcon(), w.createActionsSection()),
		container.NewTabItemWithIcon("Speech", theme.MediaRecordIcon(), w.createSpeechSection()),
		container.NewTabItemWithIcon("System", theme.SettingsIcon(), w.createSystemSection()),
	)

	// Fix for AppTabs layout width issue on Windows
	// See: https://github.com/fyne-io/fyne/issues/5338
	tabs.OnSelected = func(tab *container.TabItem) {
		go func() {
			time.Sleep(50 * time.Millisecond)
			fyne.Do(func() {
				tab.Content.Refresh()
			})
		}()
	}

	// Save button
	saveBtn := widget.NewButtonWithIcon("Save Settings", theme.DocumentSaveIcon(), w.saveSettings)
	saveBtn.Importance = widget.HighImportance

	// Main layout
	content := container.NewBorder(
		nil, // top
		container.NewBorder(nil, nil, nil, saveBtn, statusBar), // bottom
		nil,  // left
		nil,  // right
		tabs, // center
	)

	return container.NewPadded(content)
}

func (w *MainWindow) ShowWindow() {
	w.Show()
	// After the window exists: the Dock entry activates the app, and activating with nothing on
	// screen is what leaves it frontmost and empty.
	if w.onShowCallback != nil {
		w.onShowCallback()
	}
	w.RequestFocus()

	// Force layout refresh after show to fix Windows minimize/restore sizing issue
	// See: https://github.com/fyne-io/fyne/issues/300
	w.Resize(w.Canvas().Size())
	w.Content().Refresh()
}

// HideWindow hides the window and handles platform-specific behavior (e.g., macOS Dock)
func (w *MainWindow) HideWindow() {
	w.Hide()
	// Call the platform-specific hide handler if set
	if w.onHideCallback != nil {
		w.onHideCallback()
	}
}

// SetShowHideCallbacks sets the callbacks for platform-specific show/hide behavior
func (w *MainWindow) SetShowHideCallbacks(onShow func(), onHide func()) {
	w.onShowCallback = onShow
	w.onHideCallback = onHide
}

// applyAutoStartSetting enables or disables auto-start based on the setting
