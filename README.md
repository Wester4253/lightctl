# lightctl (Go)

This directory contains a complete Go rewrite of `lightctl` using Bubbletea/Bubbles/Lipgloss with improved visual polish. The Python implementation in the parent directory is preserved.

## Features (Full Parity with Python Version)

**TUI (Interactive Mode):**
- Live status panel showing all devices with on/off state, effect, and brightness
- Control Panel with detailed device statistics and capabilities  
- Effect browser with per-device effect lists
- Brightness control with visual slider and arrow key adjustment
- RGB color picker
- Theme selector with persistence to config file
- Settings viewer showing all configuration
- PC action prompts (shutdown/suspend/ignore) after profile activation
- Keyboard navigation with intuitive bindings
- Mouse support for list selection
- Rounded borders and polished Lipgloss styling with Dracula-inspired colors

**CLI (Headless Mode):**
- All profile commands: `night`, `gaming`, `movie`, `work`, `relax`, `profile NAME`
- Effect commands: `effects` (list), `effect NAME` (apply)
- Control commands: `brightness 0-100`, `color R G B`, `on`, `off`
- Status command: `status` (show all device states)
- Profile listing: `profiles`, `profile names`

## Requirements and Build

Go 1.22 or newer is required.

```bash
cd go-lightctl
go mod tidy
go build -o lightctl .
```

Run the resulting binary with no arguments for the interactive TUI, or pass a CLI command:

```bash
# Interactive TUI (default)
./lightctl

# CLI examples
./lightctl status
./lightctl gaming
./lightctl profile my_custom_scene
./lightctl effects
./lightctl effect Rainbow
./lightctl brightness 50
./lightctl color 255 0 255
./lightctl on
./lightctl off
```

## Configuration

Reads and writes the same file as the Python application: `~/.config/govee/config.json`. On first run it creates compatible defaults including the five built-in profiles (night, gaming, movie, work, relax), Govee and Nanoleaf devices, per-device profile overrides, and the tokyo-night theme. Add a Home Assistant long-lived access token to `ha_token` before use.

The `devices` map may contain any number of Home Assistant light entity IDs. Each profile can specify `effect`, `brightness_pct`, `rgb_color`, `pc_action_prompt`, and a `devices` map with per-device overrides. TUI theme changes are saved back to this file.

## Architecture

The CLI and TUI use only Home Assistant's native REST API with an authenticated Bearer token. No MQTT, NetBird, or direct light/server access is used. HTTP calls are isolated in `internal/ha`; JSON/configuration in `internal/config`; shared data types in `internal/models`; Bubbletea/Bubbles UI in `internal/app`. Profiles prompt safely for shutdown, suspend, or ignore when `pc_action_prompt` is enabled.

The TUI supports:
- Keyboard navigation with arrow keys, j/k, Enter, Esc, q
- Live device status refresh with 'r' key
- Mouse events (clicking list items)
- Multi-screen navigation (main menu, control panel, effects, inputs, settings)
- Visual feedback with color-coded status indicators
- Progress bars and input fields with validation

## Visual Improvements Over Python Version

- Rounded borders using Lipgloss
- Color-coded status indicators (green for on, red for off)
- Improved spacing and padding throughout
- Better contrast with Dracula-inspired color palette
- Visual progress bar for brightness control
- Cleaner layout with logical grouping
- Real-time status panel always visible on main screen
