//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

// AppKit is main-thread only, and these are reached from tray callbacks and from the instance
// handover, neither of which is that thread.
void SetActivationPolicyRegular(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

        // Activation on the next run-loop turn. Switching policy does not bring the app forward,
        // and doing both in one turn leaves the Dock icon there with the window behind whatever
        // was in front — visible in the menu bar, invisible on screen.
        dispatch_async(dispatch_get_main_queue(), ^{
            [NSApp activateIgnoringOtherApps:YES];
        });
    });
}

void SetActivationPolicyAccessory(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    });
}
*/
import "C"

import "github.com/paradoxe35/scribe/internal/logger"

// showInDock puts Scribe in the Dock and brings its window forward.
func showInDock() {
	logger.Info("Setting macOS activation policy to Regular (show in Dock)")
	C.SetActivationPolicyRegular()
}

// hideFromDock returns Scribe to a menu-bar-only app.
func hideFromDock() {
	logger.Info("Setting macOS activation policy to Accessory (hide from Dock)")
	C.SetActivationPolicyAccessory()
}
