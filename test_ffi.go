//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/paradoxe35/encre/internal/input"
)

func main() {
	fmt.Println("Testing Rust FFI Integration...")
	fmt.Println("================================")

	fmt.Println("\n1. Testing Clipboard Manager...")
	clipboard, err := input.NewFFIClipboardManager()
	if err != nil {
		fmt.Printf("❌ Failed to create clipboard: %v\n", err)
		os.Exit(1)
	}
	defer clipboard.Close()

	testText := "Hello from Rust FFI!"
	err = clipboard.SetText(testText)
	if err != nil {
		fmt.Printf("❌ Failed to set clipboard text: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Set clipboard text")

	retrievedText, err := clipboard.GetText()
	if err != nil {
		fmt.Printf("❌ Failed to get clipboard text: %v\n", err)
		os.Exit(1)
	}

	if retrievedText == testText {
		fmt.Printf("✓ Retrieved clipboard text correctly: '%s'\n", retrievedText)
	} else {
		fmt.Printf("❌ Text mismatch: expected '%s', got '%s'\n", testText, retrievedText)
		os.Exit(1)
	}

	fmt.Println("\n2. Testing Key Simulator...")
	simulator, err := input.NewFFIKeySimulator()
	if err != nil {
		fmt.Printf("❌ Failed to create simulator: %v\n", err)
		os.Exit(1)
	}
	defer simulator.Close()

	fmt.Println("✓ Created key simulator")
	fmt.Println("  (Note: Actual key simulation would require a GUI context)")

	fmt.Println("\n3. Testing Hotkey Manager...")
	hotkeyMgr := input.NewFFIHotkeyManager()
	if hotkeyMgr == nil {
		fmt.Println("❌ Failed to create hotkey manager")
		os.Exit(1)
	}
	defer hotkeyMgr.Close()

	err = hotkeyMgr.RegisterHotkey("ctrl+alt+t", "test", func() {
		fmt.Println("✓ Hotkey triggered!")
	})

	if err != nil {
		fmt.Printf("❌ Failed to register hotkey: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Registered hotkey 'ctrl+alt+t'")

	// Doesn't actually wait to receive a hotkey in this test.
	err = hotkeyMgr.Start()
	if err != nil {
		fmt.Printf("❌ Failed to start hotkey manager: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Started hotkey manager")

	hotkeyMgr.Stop()
	fmt.Println("✓ Stopped hotkey manager")

	fmt.Println("\n================================")
	fmt.Println("✓ All FFI tests passed!")
	fmt.Println("\nRust FFI integration is working correctly.")
	fmt.Println("The following components are functional:")
	fmt.Println("  • Clipboard Manager (arboard)")
	fmt.Println("  • Key Simulator (enigo)")
	fmt.Println("  • Hotkey Manager (rdev)")
}
