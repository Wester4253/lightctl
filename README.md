# lightctl (Go/Bubbletea)

Modern terminal UI for controlling Govee smart lights through Home Assistant.

**This is the Go/Bubbletea version.** For the Python version, see the `python` branch.

## Why Go?

- **⚡ Fast** — Instant startup with compiled binary
- **📦 Single file** — No dependencies, just download and run
- **✨ Beautiful** — Modern UI with Dracula colors, rounded borders, smooth animations
- **🖱️ Interactive** — Full mouse and keyboard support

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/Wester4253/lightctl/go/install.sh | bash
```

**Requirements:** Go 1.22+

**Manual Installation:**

```bash
git clone -b go https://github.com/wester4253/lightctl.git
cd lightctl
go build -ldflags="-s -w" -o lightctl .
sudo cp lightctl /usr/local/bin/
```

**For Python version:** See the `python` branch

---

## Features

- **Dual interface** — Rich Bubbletea TUI for browsing, or scriptable CLI for automation
- **Fast startup** — Compiled binary with instant cold start
- **Beautiful UI** — Rounded borders, Dracula-inspired colors, smooth progress bars
- **Home Assistant integration** — Works with existing integrations like `govee2mqtt`
- **Profiles** — Apply complex multi-device scenes with one command
- **Granular control** — Brightness, RGB color, and any effect your lights support
- **Multi-device** — Control several lights at once with per-device overrides
- **PC actions** — Optionally trigger `shutdown` or `suspend` after profile activation
- **Themeable TUI** — Tokyo Night, Nord, Catppuccin, Dracula, Gruvbox, Solarized, and more
- **Single binary** — No dependencies, just copy and run

## Usage

### Interactive TUI

Run with no arguments to launch the full-screen interactive UI:

```bash
lightctl
```

**Navigation:**
- `↑`/`↓` or `j`/`k` — Navigate lists
- `Enter` or `Space` — Select item
- `r` — Refresh device states/effects
- `q` or `Ctrl+C` — Quit

### CLI Commands

```bash
# Profiles
lightctl night          # Apply night profile
lightctl gaming         # Apply gaming profile
lightctl movie          # Apply movie profile
lightctl work           # Apply work profile
lightctl relax          # Apply relax profile

# Control
lightctl on             # Turn lights on
lightctl off            # Turn lights off
lightctl brightness 50  # Set brightness (0-100)
lightctl color 255 0 255  # Set RGB color

# Effects
lightctl effects        # List available effects
lightctl effect Rainbow # Apply specific effect

# Status
lightctl status         # Show current device states
```

## Configuration

Configuration is stored at `~/.config/govee/config.json` and is **compatible with the Python version**.

On first run, a default config is created. You must add your Home Assistant long-lived access token:

1. Generate token: Home Assistant → Your Profile → Long-Lived Access Tokens
2. Edit `~/.config/govee/config.json` and add your token to `ha_token`
3. Verify with `lightctl status`

**Example config:**

```json
{
  "ha_base_url": "http://localhost:8123",
  "ha_token": "your_token_here",
  "profiles": {
    "night": {
      "govee": {"brightness": 5, "effect": "Music"},
      "nanoleaf": {"brightness": 5}
    },
    "gaming": {
      "govee": {"brightness": 75, "effect": "Rainbow"},
      "nanoleaf": {"brightness": 75}
    }
  },
  "theme": "dracula"
}
```

## Architecture

This controls lights through Home Assistant only — it never talks to lights, MQTT, or servers directly.

**Control path:**
```
lightctl → Home Assistant REST API → Home Assistant → MQTT → govee2mqtt → Govee lights
```

All HTTP calls go through the Home Assistant REST API using Bearer token authentication.

**Project structure:**
```
main.go              Entry point with CLI dispatch
internal/
  app/tui.go        Bubbletea TUI implementation
  ha/client.go      Home Assistant REST client
  config/config.go  Config management
  models/models.go  Data types
```

## Profiles

Profiles apply coordinated settings across multiple devices. Each profile can set brightness, color, and effect per device.

**Default profiles:**
- `night` — Dim, calming colors for winding down
- `gaming` — Bright, energetic effects
- `movie` — Dim ambient lighting
- `work` — Bright, focused lighting
- `relax` — Comfortable, low-key ambiance

**PC Actions:** After certain profiles (like `night`), you can optionally shutdown or suspend your PC.

## Themes

Available TUI themes:
- Tokyo Night
- Nord
- Catppuccin Mocha
- Dracula (default)
- Gruvbox Dark
- Solarized Dark
- One Dark
- Monokai Pro

Change theme in the TUI Settings menu or edit `config.json`.

## Build from Source

```bash
git clone -b go https://github.com/wester4253/lightctl.git
cd lightctl
go mod download
go build -ldflags="-s -w" -o lightctl .
```

The `-ldflags="-s -w"` flag strips debug info for smaller binary size.

## Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/Wester4253/lightctl/go/uninstall.sh | bash
```

Or manually:
```bash
sudo rm /usr/local/bin/lightctl
rm -rf ~/.config/govee  # Optional: removes config
```

## License

MIT

## Contributing

Issues and pull requests welcome! This is a personal project but contributions are appreciated.
