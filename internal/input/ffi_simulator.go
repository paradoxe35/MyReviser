//go:build linux || darwin || windows

package input

/*
#cgo CFLAGS: -I${SRCDIR}/../../rust-ffi

// Linux (includes X11 and Wayland dependencies)
#cgo linux LDFLAGS: ${SRCDIR}/../../lib/libencre_ffi.a -lpthread -ldl -lm -lxdo -lX11 -lXtst -lxkbcommon

// macOS
#cgo darwin LDFLAGS: ${SRCDIR}/../../lib/libencre_ffi.a
#cgo darwin LDFLAGS: -framework CoreFoundation -framework AppKit -framework ApplicationServices -framework Carbon

// Windows
#cgo windows LDFLAGS: ${SRCDIR}/../../lib/libencre_ffi.a
#cgo windows LDFLAGS: -lws2_32 -luserenv -lbcrypt -lntdll -static

#include <stdlib.h>
#include "bindings.h"
*/
import "C"
import (
	"fmt"

	"github.com/paradoxe35/encre/internal/logger"
)

type FFIKeySimulator struct {
	handle C.encre_SimulatorHandle
}

func NewFFIKeySimulator() (*FFIKeySimulator, error) {
	handle := C.encre_simulator_new()
	if handle == nil {
		return nil, fmt.Errorf("failed to create key simulator: %s", getLastError())
	}

	return &FFIKeySimulator{handle: handle}, nil
}

// SelectAll simulates Ctrl+A / Cmd+A.
func (s *FFIKeySimulator) SelectAll() error {
	if s.handle == nil {
		return fmt.Errorf("key simulator not initialized")
	}

	logger.Debug("FFI: Simulating Select All")
	result := C.encre_simulate_select_all(s.handle)
	if result != 0 {
		return fmt.Errorf("failed to simulate select all: %s", getLastError())
	}

	return nil
}

// Copy simulates Ctrl+C / Cmd+C.
func (s *FFIKeySimulator) Copy() error {
	if s.handle == nil {
		return fmt.Errorf("key simulator not initialized")
	}

	logger.Debug("FFI: Simulating Copy")
	result := C.encre_simulate_copy(s.handle)
	if result != 0 {
		return fmt.Errorf("failed to simulate copy: %s", getLastError())
	}

	return nil
}

// Paste simulates Ctrl+V / Cmd+V.
func (s *FFIKeySimulator) Paste() error {
	if s.handle == nil {
		return fmt.Errorf("key simulator not initialized")
	}

	logger.Debug("FFI: Simulating Paste")
	result := C.encre_simulate_paste(s.handle)
	if result != 0 {
		return fmt.Errorf("failed to simulate paste: %s", getLastError())
	}

	return nil
}

// ReleaseModifiers drops modifiers still held from the triggering hotkey; otherwise simulating
// Ctrl+A while Alt is down sends Ctrl+Alt+A. On macOS it also waits for the keyboard to confirm
// the keys are up, since a posted event merges with modifiers still held.
func (s *FFIKeySimulator) ReleaseModifiers() error {
	if s.handle == nil {
		return fmt.Errorf("key simulator not initialized")
	}

	result := C.encre_simulate_release_modifiers(s.handle)
	if result != 0 {
		return fmt.Errorf("failed to release modifiers: %s", getLastError())
	}

	return nil
}

func (s *FFIKeySimulator) Close() {
	if s.handle != nil {
		C.encre_simulator_free(s.handle)
		s.handle = nil
	}
}

func FFISimulateSelectAll() error {
	sim, err := NewFFIKeySimulator()
	if err != nil {
		return err
	}
	defer sim.Close()

	return sim.SelectAll()
}

func FFISimulateCopy() error {
	sim, err := NewFFIKeySimulator()
	if err != nil {
		return err
	}
	defer sim.Close()

	return sim.Copy()
}

func FFISimulatePaste() error {
	sim, err := NewFFIKeySimulator()
	if err != nil {
		return err
	}
	defer sim.Close()

	return sim.Paste()
}
