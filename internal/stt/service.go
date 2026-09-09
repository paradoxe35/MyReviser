package stt

import (
	"errors"
	"fmt"
	"sync"

	"github.com/paradoxe35/encre/internal/config"
	"github.com/paradoxe35/encre/internal/input"
	"github.com/paradoxe35/encre/internal/logger"
)

var ErrNoModel = errors.New("no speech model selected - choose one in Settings")

// Service turns hotkey edges into transcripts. The engine is created lazily so
// an app that never dictates never opens an audio device.
type Service struct {
	store *Store

	mu        sync.Mutex
	speech    *input.FFISpeech
	recording bool
	loaded    string
	device    string
}

func NewService() *Service {
	return &Service{store: NewStore()}
}

func (s *Service) Store() *Store { return s.store }

// OnLevel receives microphone level while recording, from a background thread.
func (s *Service) OnLevel(handler func(float32)) { input.OnLevel(handler) }

func (s *Service) engine() (*input.FFISpeech, error) {
	if s.speech != nil {
		return s.speech, nil
	}

	speech, err := input.NewFFISpeech()
	if err != nil {
		return nil, err
	}

	s.speech = speech
	s.loaded = ""
	s.device = ""
	return speech, nil
}

// Prepare loads the model and opens the microphone ahead of the first
// dictation. Every step is idempotent, so it is cheap to call repeatedly.
func (s *Service) Prepare(cfg config.SpeechConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	speech, err := s.engine()
	if err != nil {
		return err
	}
	if err := s.load(speech, cfg); err != nil {
		return err
	}

	s.applyDevice(speech, cfg)
	if cfg.WarmMicrophone {
		speech.Warm()
	}
	return nil
}

// applyDevice is a no-op when nothing changed, so a warmed stream is not torn
// down and reopened on every dictation.
func (s *Service) applyDevice(speech *input.FFISpeech, cfg config.SpeechConfig) {
	if s.device == cfg.InputDevice {
		return
	}
	if err := speech.SetDevice(cfg.InputDevice); err != nil {
		logger.Warn("Could not select the microphone", "device", cfg.InputDevice, "error", err)
		return
	}
	s.device = cfg.InputDevice
}

func (s *Service) load(speech *input.FFISpeech, cfg config.SpeechConfig) error {
	if cfg.ModelID == "" {
		return ErrNoModel
	}

	model, ok := FindModel(cfg.ModelID)
	if !ok {
		return fmt.Errorf("unknown model %q", cfg.ModelID)
	}
	if !s.store.Downloaded(model) {
		return fmt.Errorf("%s is not downloaded yet", model.Name)
	}

	path := s.store.Path(model)
	if s.loaded == path {
		return nil
	}
	if err := speech.Load(path); err != nil {
		return err
	}

	s.loaded = path
	logger.Info("Speech model loaded", "model", model.Name)
	return nil
}

// StartRecording loads the model too, so the first dictation needs no setup.
func (s *Service) StartRecording(cfg config.SpeechConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.recording {
		return errors.New("already recording")
	}

	speech, err := s.engine()
	if err != nil {
		return err
	}
	if err := s.load(speech, cfg); err != nil {
		return err
	}

	s.applyDevice(speech, cfg)
	if err := speech.Start(); err != nil {
		return err
	}

	s.recording = true
	return nil
}

// StopRecording blocks for as long as transcription takes.
func (s *Service) StopRecording() (string, error) {
	s.mu.Lock()
	speech, recording := s.speech, s.recording
	s.recording = false
	s.mu.Unlock()

	if speech == nil || !recording {
		return "", errors.New("not recording")
	}

	text, err := speech.Stop()
	if err != nil {
		return "", err
	}

	logger.Info("Dictation transcribed", "characters", len(text))
	return text, nil
}

func (s *Service) Cancel() {
	s.mu.Lock()
	speech := s.speech
	s.recording = false
	s.mu.Unlock()

	if speech != nil {
		speech.Cancel()
	}
}

// TranscribeFile verifies a model against a 16 kHz mono WAV.
func (s *Service) TranscribeFile(cfg config.SpeechConfig, path string) (string, error) {
	s.mu.Lock()
	speech, err := s.engine()
	if err == nil {
		err = s.load(speech, cfg)
	}
	s.mu.Unlock()

	if err != nil {
		return "", err
	}
	return speech.TranscribeFile(path)
}

// Devices lists microphones. Creating the engine is deferred until something
// needs it, so this is the first call that may open an audio device.
func (s *Service) Devices() []input.Device {
	return input.InputDevices()
}

func (s *Service) Close() {
	s.mu.Lock()
	speech := s.speech
	s.speech = nil
	s.recording = false
	s.mu.Unlock()

	if speech != nil {
		speech.Close()
	}
}
