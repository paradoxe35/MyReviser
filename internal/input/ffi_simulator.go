//go:build linux || darwin || windows

package input

/*
#cgo CFLAGS: -I${SRCDIR}/../../rust-ffi

// Linux (includes X11 and Wayland dependencies)
#cgo linux LDFLAGS: ${SRCDIR}/../../lib/libscribe_ffi.a -lpthread -ldl -lm -lxdo -lX11 -lXtst -lxkbcommon

// macOS
#cgo darwin LDFLAGS: ${SRCDIR}/../../lib/libscribe_ffi.a
#cgo darwin LDFLAGS: -framework CoreFoundation -framework AppKit -framework ApplicationServices -framework Carbon

// Windows
#cgo windows LDFLAGS: ${SRCDIR}/../../lib/libscribe_ffi.a
#cgo windows LDFLAGS: -lws2_32 -luserenv -lbcrypt -lntdll -static

#include <stdlib.h>
#include "bindings.h"
*/
import "C"
import (
	"fmt"

	"github.com/paradoxe35/scribe/internal/logger"
)

// FFIKeySimulator wraps the Rust FFI key simulator
type FFIKeySimulator struct {
	handle C.scribe_SimulatorHandle
}

// NewFFIKeySimulator creates a new FFI-based key simulator
func NewFFIKeySimulator() (*FFIKeySimulator, error) {
	handle := C.scribe_simulator_new()
	if handle == nil {
		return nil, fmt.Errorf("failed to create key simulator: %s", getLastError())
	}

	return &FFIKeySimulator{handle: handle}, nil
}

// SelectAll simulates the select all keyboard shortcut (Ctrl+A / Cmd+A)
func (s *FFIKeySimulator) SelectAll() error {
	if s.handle == nil {
		return fmt.Errorf("key simulator not initialized")
	}

	logger.Debug("FFI: Simulating Select All")
	result := C.scribe_simulate_select_all(s.handle)
	if result != 0 {
		return fmt.Errorf("failed to simulate select all: %s", getLastError())
	}

	return nil
}

// Copy simulates the copy keyboard shortcut (Ctrl+C / Cmd+C)
func (s *FFIKeySimulator) Copy() error {
	if s.handle == nil {
		return fmt.Errorf("key simulator not initialized")
	}

	logger.Debug("FFI: Simulating Copy")
	result := C.scribe_simulate_copy(s.handle)
	if result != 0 {
		return fmt.Errorf("failed to simulate copy: %s", getLastError())
	}

	return nil
}

// Paste simulates the paste keyboard shortcut (Ctrl+V / Cmd+V)
func (s *FFIKeySimulator) Paste() error {
	if s.handle == nil {
		return fmt.Errorf("key simulator not initialized")
	}

	logger.Debug("FFI: Simulating Paste")
	result := C.scribe_simulate_paste(s.handle)
	if result != 0 {
		return fmt.Errorf("failed to simulate paste: %s", getLastError())
	}

	return nil
}

// ReleaseModifiers drops modifiers still held from the triggering hotkey.
//
// Without it, simulating Ctrl+A while Alt is still down sends Ctrl+Alt+A and selects nothing.
// On macOS this also waits for the keyboard to agree the keys are up, because a posted event is
// merged with the modifiers still being held.
func (s *FFIKeySimulator) ReleaseModifiers() error {
	if s.handle == nil {
		return fmt.Errorf("key simulator not initialized")
	}

	result := C.scribe_simulate_release_modifiers(s.handle)
	if result != 0 {
		return fmt.Errorf("failed to release modifiers: %s", getLastError())
	}

	return nil
}

// Close frees the key simulator resources
func (s *FFIKeySimulator) Close() {
	if s.handle != nil {
		C.scribe_simulator_free(s.handle)
		s.handle = nil
	}
}

// FFI-based implementations of simulator functions
// These will replace the robotgo-based functions

// FFISimulateSelectAll simulates the select all keyboard shortcut
func FFISimulateSelectAll() error {
	sim, err := NewFFIKeySimulator()
	if err != nil {
		return err
	}
	defer sim.Close()

	return sim.SelectAll()
}

// FFISimulateCopy simulates the copy keyboard shortcut
func FFISimulateCopy() error {
	sim, err := NewFFIKeySimulator()
	if err != nil {
		return err
	}
	defer sim.Close()

	return sim.Copy()
}

// FFISimulatePaste simulates the paste keyboard shortcut
func FFISimulatePaste() error {
	sim, err := NewFFIKeySimulator()
	if err != nil {
		return err
	}
	defer sim.Close()

	return sim.Paste()
}
