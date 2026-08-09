# lightctl

[![Python 3.10+](https://img.shields.io/badge/python-3.10%2B-blue.svg)](https://www.python.org/downloads/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Built with Textual](https://img.shields.io/badge/TUI-Textual-purple.svg)](https://github.com/Textualize/textual)
[![Ask DeepWiki](https://devin.ai/assets/askdeepwiki.png)](https://deepwiki.com/Wester4253/lightctl)

A fast terminal tool for controlling Govee lights (and other Home Assistant-connected devices) — without touching the Home Assistant UI.

Switch scenes, tweak colors, or fire off a "night mode" command straight from your shell, whether you want a full interactive TUI or a one-line CLI call for your scripts.

```bash
lightctl gaming        # apply a full lighting scene in one command
lightctl brightness 50 # or make quick one-off adjustments
lightctl                # or launch the interactive TUI
```

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Dual interface** — a rich Bubbletea/Lipgloss TUI (Go) or Textual TUI (Python) for browsing and clicking around, or a scriptable CLI for quick actions and automation.
- **Fast startup** — Go binary with instant cold start, no Python interpreter overhead.
- **Beautiful UI** — Rounded borders, Dracula-inspired colors, smooth progress bars, and visual status indicators.
- **Home Assistant integration** — talks directly to the Home Assistant REST API, working with existing integrations like `govee2mqtt`.
- **Profiles** — apply complex, multi-device scenes with one command (`lightctl gaming`). Ships with `night`, `gaming`, `movie`, `work`, and `relax`, and supports fully custom profiles.
- **Granular control** — brightness, RGB color, and any effect your lights support.
- **Multi-device** — control several lights at once, with per-device overrides inside a single profile.
- **PC actions** — optionally trigger `shutdown` or `suspend` after a profile runs (e.g. shut down when `night` mode kicks in).
- **Themeable TUI** — Tokyo Night, Nord, Catppuccin, Dracula, Gruvbox, Solarized, and more.
- **Single binary** — No dependencies, just copy and run (Go version).

## Installation

### Quick Install (Go Version - Recommended)

The fastest way to get lightctl up and running with the modern Go/Bubbletea TUI:

```bash
curl -fsSL https://raw.githubusercontent.com/Wester4253/lightctl/main/install.sh | bash
```

Or with wget:
```bash
wget -qO- https://raw.githubusercontent.com/Wester4253/lightctl/main/install.sh | bash
```

This will:
- Check for Go 1.22+ (required)
- Clone the repository
- Build the optimized binary
- Install to `/usr/local/bin/lightctl`

**Manual Installation (Go):**

```bash
git clone https://github.com/wester4253/lightctl.git
cd lightctl/go-lightctl
go build -o lightctl .
sudo cp lightctl /usr/local/bin/
```

### Python Version (Legacy)

**Requirements:** Python 3.10+

```bash
git clone https://github.com/wester4253/lightctl.git
cd lightctl
pip install textual typer requests
chmod +x lightctl
sudo cp lightctl /usr/bin/lightctl
```

Once installed, `lightctl` is available as a regular command from anywhere on your system.

## Configuration

On first run, `lightctl` creates a config file at `~/.config/govee/config.json`. You'll need to edit it before it can talk to Home Assistant:

1. Generate a **Long-Lived Access Token** from your Home Assistant profile page (scroll to the bottom).
2. Open `~/.config/govee/config.json`.
3. Paste the token into `ha_token`.
4. Set `ha_base_url` to your Home Assistant instance, and confirm `entity_id` / `devices` match your actual light entities.

<details>
<summary>Example <code>config.json</code></summary>

```json
{
  "ha_base_url": "http://homeassistant.local:8123",
  "ha_token": "PASTE_YOUR_LONG_LIVED_ACCESS_TOKEN_HERE",
  "entity_id": "light.govee_strip",
  "devices": {
    "govee": "light.govee_strip",
    "nanoleaf": "light.nanoleaf_panels"
  },
  "theme": "tokyo-night",
  "profiles": {
    "night": {
      "name": "night",
      "effect": "Sleep",
      "brightness_pct": 15,
      "pc_action_prompt": true,
      "devices": {
        "nanoleaf": {
          "effect": "moonlight",
          "brightness_pct": 15,
          "rgb_color": null
        }
      }
    }
  }
}
```

</details>

## Usage

### Interactive TUI

Run with no arguments to launch the full Textual interface — navigate with your keyboard to apply profiles, change effects, adjust brightness, and see live device status.

```bash
lightctl
```

### CLI

**Power**
```bash
lightctl on
lightctl off
```

**State & effects**
```bash
lightctl status              # current status
lightctl brightness 50       # set brightness (0-100)
lightctl color 255 0 255     # set a solid RGB color
lightctl effects             # list available effects
lightctl effect Rainbow      # apply a specific effect
```

**Profiles**
```bash
lightctl night
lightctl gaming
lightctl movie
lightctl work
lightctl relax

lightctl profile my_custom_scene   # any custom profile from config.json
```

## Architecture

`lightctl` keeps a clean separation between presentation and logic — both the TUI and CLI delegate to the same core:

| Module | Responsibility |
|---|---|
| `lightctl_app.py` | Entry point; launches TUI or CLI based on args |
| `tui.py` | Presentation-only interactive interface (Textual) |
| `cli.py` | Presentation-only command-line interface (Typer) |
| `commands.py` | Core application logic and business rules |
| `ha.py` | Client for the Home Assistant REST API |
| `config.py` | Loads/saves `~/.config/govee/config.json` |
| `profiles.py` | Default lighting profiles |
| `models.py` | Data structures (`Config`, `Profile`, `LightState`) |

## Contributing

Issues and pull requests are welcome. If you're adding a new profile, effect, or theme, keep it in line with the existing structure in `profiles.py` / `config.py` where possible.

## License

[MIT](LICENSE)
