# lightctl (Python/Textual)

Modern terminal UI for controlling Govee smart lights through Home Assistant.

**This is the Python/Textual version.** For the faster Go version, see the `go` branch.

## Quick Install

```bash
git clone -b python https://github.com/wester4253/lightctl.git
cd lightctl
pip install textual typer requests
chmod +x lightctl
sudo cp lightctl /usr/local/bin/
```

**Requirements:** Python 3.10+

**For Go version:** See the `go` branch (recommended for speed and portability)

---

## Features

- **Dual interface** — Rich Textual TUI for browsing, or Typer CLI for automation
- **Home Assistant integration** — Works with existing integrations like `govee2mqtt`
- **Profiles** — Apply complex multi-device scenes with one command
- **Granular control** — Brightness, RGB color, and any effect your lights support
- **Multi-device** — Control several lights at once with per-device overrides
- **PC actions** — Optionally trigger `shutdown` or `suspend` after profile activation
- **Themeable TUI** — Nord, Catppuccin, Gruvbox, and more

## Usage

### Interactive TUI

```bash
lightctl
```

### CLI Commands

```bash
lightctl night          # Apply profile
lightctl brightness 50  # Set brightness
lightctl color 255 0 255  # Set RGB color
lightctl effect Rainbow # Apply effect
lightctl status         # Show device states
```

## Configuration

Configuration is stored at `~/.config/govee/config.json` and is **compatible with the Go version**.

On first run, you must add your Home Assistant long-lived access token:

1. Generate token: Home Assistant → Your Profile → Long-Lived Access Tokens
2. Edit `~/.config/govee/config.json` and add your token
3. Run `lightctl status` to verify

## Architecture

Control path:
```
lightctl → Home Assistant REST API → MQTT → govee2mqtt → Govee lights
```

## Default Profiles

- `night` — Dim, calming colors
- `gaming` — Bright, energetic effects
- `movie` — Dim ambient lighting
- `work` — Bright, focused lighting
- `relax` — Comfortable ambiance

## License

MIT
