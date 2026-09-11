//go:build !darwin

package main

// showInDock is a no-op outside macOS; there's no Dock to show it in.
func showInDock() {
}

// hideFromDock is a no-op outside macOS; there's no Dock to hide it from.
func hideFromDock() {
}
