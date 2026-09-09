//go:build darwin

package main

/*
#cgo LDFLAGS: -framework Cocoa
#include "reopen_darwin.h"
*/
import "C"

var onReopen func()

//export scribeHandleReopen
func scribeHandleReopen() {
	if onReopen != nil {
		onReopen()
	}
}

// installReopenHandler makes clicking the app's icon bring the window back, which is what macOS
// expects an already-running app to do. Fyne installs no handler for it, so the method is added to
// the delegate it already owns.
func installReopenHandler(show func()) {
	onReopen = show
	C.ScribeInstallReopenHandler()
}
