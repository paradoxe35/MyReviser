//go:build darwin

package main

/*
#cgo LDFLAGS: -framework Cocoa
#include "reopen_darwin.h"
*/
import "C"

var onReopen func()

//export encreHandleReopen
func encreHandleReopen() {
	if onReopen != nil {
		onReopen()
	}
}

// installReopenHandler makes clicking the Dock icon reopen the window. Fyne doesn't wire this up,
// so the handler is added to the delegate Fyne already owns.
func installReopenHandler(show func()) {
	onReopen = show
	C.EncreInstallReopenHandler()
}
