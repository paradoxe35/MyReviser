package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/paradoxe35/encre/internal/config"
	"github.com/paradoxe35/encre/internal/stt"
)

const remoteEngineLabel = "Hosted service"

func (w *MainWindow) createSpeechSection() fyne.CanvasObject {
	w.speechEngine = widget.NewSelect(
		[]string{"On this computer", remoteEngineLabel}, nil)
	w.speechEngine.SetSelected(engineLabel(w.config.Speech.Engine))

	local := w.localSpeechPane()
	remote := w.remoteSpeechPane()

	show := func(label string) {
		if label == remoteEngineLabel {
			local.Hide()
			remote.Show()
		} else {
			remote.Hide()
			local.Show()
		}
	}
	w.speechEngine.OnChanged = show
	show(w.speechEngine.Selected)

	// Options are set once and rarely revisited; the model list is what the
	// screen is for. A dialog keeps the list full height.
	options := widget.NewButtonWithIcon("", theme.SettingsIcon(), w.showSpeechOptions)
	options.Importance = widget.LowImportance

	w.buildSpeechOptions()

	return container.NewBorder(
		container.NewPadded(container.NewVBox(
			w.speechHeader(),
			container.NewBorder(nil, nil, widget.NewLabel("Transcribe"), options, w.speechEngine),
		)),
		nil, nil, nil,
		container.NewStack(local, remote),
	)
}

func engineLabel(engine config.SpeechEngine) string {
	if engine == config.SpeechRemote {
		return remoteEngineLabel
	}
	return "On this computer"
}

func (w *MainWindow) speechHeader() fyne.CanvasObject {
	hint := widget.NewLabel("Hold the dictate shortcut, speak, release.")
	hint.Wrapping = fyne.TextWrapWord
	return hint
}

func (w *MainWindow) localSpeechPane() *fyne.Container {
	w.speechModels = NewModelList(w.speechStore(), w.Window, w.config.Speech.ModelID,
		func(model stt.Model) {
			w.config.Speech.ModelID = model.ID
			w.statusBinding.Set("Speech model set to " + model.Name)
		})

	catalog := stt.Models()
	summary := widget.NewLabel(fmt.Sprintf("%d models", len(catalog.Models)))
	summary.TextStyle.Italic = true

	refresh := widget.NewButton("Check for new", w.refreshCatalog)

	return container.NewBorder(
		container.NewPadded(container.NewBorder(nil, nil, summary, refresh)),
		nil, nil, nil,
		w.speechModels,
	)
}

func (w *MainWindow) refreshCatalog() {
	w.statusBinding.Set("Checking for new models…")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := stt.Refresh(ctx)
		fyne.Do(func() {
			if err != nil {
				w.statusBinding.Set("Could not reach the model list")
				return
			}
			w.statusBinding.Set("Model list updated")
			w.speechModels.apply()
		})
	}()
}

func (w *MainWindow) remoteSpeechPane() *fyne.Container {
	speech := w.config.Speech

	w.speechRemoteModel = widget.NewSelectEntry(nil)
	w.speechRemoteModel.SetText(speech.RemoteModel)

	w.speechRemoteURL = widget.NewEntry()
	w.speechRemoteURL.SetPlaceHolder("https://api.example.com/v1")
	w.speechRemoteURL.SetText(speech.RemoteBaseURL)

	w.speechRemoteKey = widget.NewPasswordEntry()
	w.speechRemoteKey.SetPlaceHolder("API key")
	w.speechRemoteKey.SetText(speech.RemoteAPIKey)

	w.speechRemote = widget.NewSelect(stt.PresetNames(), w.applyPreset)
	if preset, ok := stt.FindPreset(speech.RemoteProvider); ok {
		w.speechRemote.SetSelected(preset.Name)
	} else {
		w.speechRemote.SetSelected(stt.RemotePresets[0].Name)
	}

	note := widget.NewLabel("Audio is sent to this service. Nothing is downloaded.")
	note.Wrapping = fyne.TextWrapWord
	note.TextStyle.Italic = true

	return container.NewPadded(container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Service", w.speechRemote),
			widget.NewFormItem("Model", w.speechRemoteModel),
			widget.NewFormItem("Endpoint", w.speechRemoteURL),
			widget.NewFormItem("API key", w.speechRemoteKey),
		),
		note,
	))
}

// applyPreset fills the endpoint for known services and leaves it editable only
// for a custom one, so a typo cannot silently break a working provider.
func (w *MainWindow) applyPreset(name string) {
	preset, ok := stt.PresetByName(name)
	if !ok {
		return
	}

	w.speechRemoteModel.SetOptions(preset.Models)

	if preset.ID == "custom" {
		w.speechRemoteURL.Enable()
		return
	}

	w.speechRemoteURL.SetText(preset.BaseURL)
	w.speechRemoteURL.Disable()
	if len(preset.Models) > 0 && w.speechRemoteModel.Text == "" {
		w.speechRemoteModel.SetText(preset.Models[0])
	}
}

// Built once so Save reads the same widgets whether or not the dialog was
// ever opened.
func (w *MainWindow) buildSpeechOptions() {
	speech := w.config.Speech

	w.microphone = NewMicrophonePicker(speech.InputDevice)

	w.speechKeepLoaded = widget.NewCheck("Keep the model in memory", nil)
	w.speechKeepLoaded.SetChecked(speech.KeepModelLoaded)

	w.speechCleanUp = widget.NewCheck("Tidy the transcript with AI", nil)
	w.speechCleanUp.SetChecked(speech.CleanUp)
}

func (w *MainWindow) showSpeechOptions() {
	w.microphone.Refresh()

	content := container.NewVBox(
		widget.NewLabel("Microphone"),
		w.microphone,
		widget.NewSeparator(),
		w.speechKeepLoaded,
		w.speechCleanUp,
	)

	options := dialog.NewCustom("Speech options", "Done", content, w.Window)
	options.Resize(fyne.NewSize(430, 330))
	options.Show()
}

func (w *MainWindow) speechStore() *stt.Store {
	if w.speechStoreRef == nil {
		w.speechStoreRef = stt.NewStore()
	}
	return w.speechStoreRef
}

func (w *MainWindow) applySpeechSettings() {
	speech := &w.config.Speech

	speech.Engine = config.SpeechLocal
	if w.speechEngine.Selected == remoteEngineLabel {
		speech.Engine = config.SpeechRemote
	}

	speech.InputDevice = w.microphone.Device()
	speech.KeepModelLoaded = w.speechKeepLoaded.Checked
	speech.CleanUp = w.speechCleanUp.Checked

	if preset, ok := stt.PresetByName(w.speechRemote.Selected); ok {
		speech.RemoteProvider = preset.ID
	}
	speech.RemoteModel = w.speechRemoteModel.Text
	speech.RemoteBaseURL = w.speechRemoteURL.Text
	speech.RemoteAPIKey = w.speechRemoteKey.Text
}
