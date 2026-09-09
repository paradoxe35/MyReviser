//go:build linux || darwin || windows

package input

/*
// Speech pulls in ggml, which is C++, and cpal, which talks to the platform's
// audio API. cgo unions LDFLAGS across the package, so they belong here rather
// than repeated in every file.
#cgo linux LDFLAGS: -lstdc++ -lasound
#cgo darwin LDFLAGS: -lc++ -framework AudioToolbox -framework CoreAudio -framework AudioUnit
#cgo windows LDFLAGS: -lstdc++ -lole32 -lavrt

#include <stdlib.h>
#include "bindings.h"

extern void speechLevelGateway(float rms);
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// FFISpeech records from the microphone and transcribes locally. Every call
// beyond Level runs on the caller's goroutine; Stop blocks for as long as
// inference takes.
type FFISpeech struct {
	mu     sync.Mutex
	handle C.encre_SttHandle
}

var (
	levelMu      sync.RWMutex
	levelHandler func(float32)
)

func NewFFISpeech() (*FFISpeech, error) {
	handle := C.encre_stt_new(C.encre_LevelCallback(C.speechLevelGateway))
	if handle == nil {
		return nil, fmt.Errorf("failed to create speech recogniser: %s", getLastError())
	}
	return &FFISpeech{handle: handle}, nil
}

// OnLevel receives microphone level while recording. It is called from a
// background thread, and replaces any previous handler.
func OnLevel(handler func(float32)) {
	levelMu.Lock()
	defer levelMu.Unlock()
	levelHandler = handler
}

//export speechLevelGateway
func speechLevelGateway(rms C.float) {
	levelMu.RLock()
	handler := levelHandler
	levelMu.RUnlock()

	if handler != nil {
		handler(float32(rms))
	}
}

// Load keeps a model resident. Idempotent for the same path.
func (s *FFISpeech) Load(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	s.mu.Lock()
	result := C.encre_stt_load(s.handle, cPath)
	s.mu.Unlock()

	if result != 0 {
		return fmt.Errorf("%s", getLastError())
	}
	return nil
}

func (s *FFISpeech) Unload() {
	s.mu.Lock()
	defer s.mu.Unlock()
	C.encre_stt_unload(s.handle)
}

// Warm opens the capture device ahead of time so Start costs nothing.
func (s *FFISpeech) Warm() {
	s.mu.Lock()
	defer s.mu.Unlock()
	C.encre_stt_warm(s.handle)
}

func (s *FFISpeech) Start() error {
	s.mu.Lock()
	result := C.encre_stt_start(s.handle)
	s.mu.Unlock()

	if result != 0 {
		return fmt.Errorf("%s", getLastError())
	}
	return nil
}

// Stop ends recording and returns the transcript. It blocks for the length of
// the transcription, so callers should not run it on the UI goroutine.
func (s *FFISpeech) Stop() (string, error) {
	s.mu.Lock()
	text := C.encre_stt_stop(s.handle)
	s.mu.Unlock()

	return takeString(text)
}

func (s *FFISpeech) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	C.encre_stt_cancel(s.handle)
}

// TranscribeFile reads a 16 kHz mono WAV, for verifying a model without a
// microphone.
func (s *FFISpeech) TranscribeFile(path string) (string, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	s.mu.Lock()
	text := C.encre_stt_transcribe_file(s.handle, cPath)
	s.mu.Unlock()

	return takeString(text)
}

func (s *FFISpeech) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handle != nil {
		C.encre_stt_free(s.handle)
		s.handle = nil
	}
}

func takeString(text *C.char) (string, error) {
	if text == nil {
		return "", fmt.Errorf("%s", getLastError())
	}
	defer C.encre_free_string(text)
	return C.GoString(text), nil
}

// SetDevice chooses the capture device by name. An empty name means the system
// default. It applies to the next recording, not one in progress.
func (s *FFISpeech) SetDevice(name string) error {
	var cName *C.char
	if name != "" {
		cName = C.CString(name)
		defer C.free(unsafe.Pointer(cName))
	}

	s.mu.Lock()
	result := C.encre_stt_set_device(s.handle, cName)
	s.mu.Unlock()

	if result != 0 {
		return fmt.Errorf("%s", getLastError())
	}
	return nil
}

// InputDevices lists microphones. An empty result means none were found, not
// that enumeration failed.
func InputDevices() []Device {
	listed := C.encre_stt_devices()
	if listed == nil {
		return nil
	}
	defer C.encre_free_string(listed)

	return parseDevices(C.GoString(listed))
}
