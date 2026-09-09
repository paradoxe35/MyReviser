# Scribe

[![Build and Release](https://github.com/paradoxe35/scribe/actions/workflows/build.yml/badge.svg)](https://github.com/paradoxe35/scribe/actions/workflows/build.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple tool that fixes your text with AI, anywhere on your computer.

## Why I Built This

English isn't my first language. Every time I write an email, a message, or some documentation, I end up with grammar mistakes or typos. I got tired of copying text to ChatGPT, waiting for a response, then copying it back.

I wanted something simpler: press a hotkey, and my text gets fixed. No switching apps, no copy-paste dance. Just write, press a key, done.

That's Scribe. It sits in your system tray, listens for a hotkey, grabs your text, sends it to an AI, and replaces it with the corrected version. All in a few seconds.

You can also customize the prompt to do other things - translate, summarize, change tone, whatever you need. But for me, it's mostly about fixing my terrible grammar.

## Demo

![Scribe Demo](assets/demo.gif)

## How It Works

1. You're writing somewhere (email, browser, notes, anywhere)
2. Select your text (or use the "select all" hotkey)
3. Press the hotkey
4. Your text gets replaced with the AI-improved version

That's it.

## Installation

Download the [latest release](https://github.com/paradoxe35/scribe/releases/latest) for your system:

- **Windows**: Run the installer or extract the portable ZIP
- **macOS**: Open the DMG, drag to Applications. First launch: right-click > Open
- **Linux**: Use the .deb package, AppImage, or portable archive

On first launch, grant accessibility/input permissions when prompted - this is needed for global hotkeys to work.

### Linux Users

**Wayland users only**: You need to be in the `input` group for hotkeys to work:

```bash
sudo usermod -aG input $USER
# Then log out and log back in
```

X11 users don't need this - hotkeys work out of the box.

**Required packages** (Debian/Ubuntu - the .deb installs these automatically):

```bash
sudo apt install libgl1 libx11-6 libxext6 libxcb1 libxinerama1 libxtst6 libxdo3 libxkbcommon0 libxi6 libxcursor1 libxrandr2 libxrender1 libxfixes3 libxxf86vm1
```

## Quick Start

1. Launch Scribe (it appears in your system tray)
2. Right-click the tray icon > Settings
3. Add your API key for OpenAI, Claude, or Gemini
4. Start writing somewhere, select text, press the hotkey

## Default Hotkeys

| Action              | Linux/Windows    | macOS               |
| ------------------- | ---------------- | ------------------- |
| Select All & Revise | `Ctrl+Alt+Space` | `Ctrl+Option+Space` |
| Revise Selection    | `Ctrl+Super`     | `Ctrl+Cmd`          |

You can change these in Settings > Hotkeys.

## Supported AI Providers

| Provider | Example Models                          |
| -------- | --------------------------------------- |
| OpenAI   | gpt-4o, gpt-4o-mini                     |
| Claude   | claude-3-5-haiku, claude-3-5-sonnet     |
| Gemini   | gemini-2.5-flash, gemini-2.5-flash-lite |

You can also add custom OpenAI-compatible providers (like local LLMs, OpenRouter, Together AI) with their own base URL.

## Configuration

Config files are created on first run:

| Platform      | Path                                   |
| ------------- | -------------------------------------- |
| Linux / macOS | `~/.scribe/config.json`             |
| Windows       | `%USERPROFILE%\.scribe\config.json` |

## Building From Source

You'll need:

- Go 1.24+ with `CGO_ENABLED=1`
- Rust toolchain (stable)
- Platform dependencies (see `rust-ffi/README.md`)

```bash
make build    # Build for current platform
make run      # Run locally
make test     # Run tests
```

The app uses a Go frontend (Fyne UI) with a Rust backend for system input handling (hotkeys, clipboard, key simulation). This hybrid approach gives us reliable cross-platform support.

## Troubleshooting

**Hotkeys not working?**

- Check that only one instance is running (look in system tray)
- On macOS: grant Accessibility permissions in System Settings
- On Linux Wayland: make sure you're in the `input` group (X11 doesn't need this)

**Revisions failing?**

- Check your API key is valid
- Look at logs in `~/.scribe/logs/`

## License

MIT - see [LICENSE](LICENSE)
