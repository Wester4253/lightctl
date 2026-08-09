# Copilot instructions — lightctl

A short, focused guide to help Copilot-based sessions work effectively in this repository.

## Quick runtime / run commands

- Language: Python (uses modern type syntax - run on Python 3.10+).
- Recommended setup (if no lockfile present):
  - python -m venv .venv
  - source .venv/bin/activate
  - pip install textual typer requests

- Run the TUI (default when no args):
  ./lightctl

- Run the CLI (pass subcommand name and args):
  ./lightctl <command> [args]

Examples:
  ./lightctl status
  ./lightctl effect Rainbow
  ./lightctl brightness 50
  ./lightctl color 255 0 255
  ./lightctl night

- Build/test/lint: none configured in this repo.
  - No test harness is present. If pytest is added, run a single test with:
    pytest path/to/test_file.py::test_name -q
  - No linter config found. Add ruff/flake8/black or similar if desired.

## High-level architecture (big picture)

- lightctl
  - Entrypoint. No arguments -> launches the Textual TUI. With arguments -> invokes the Typer-based CLI.

- cli.py
  - Typer CLI layer; deliberately thin. Parses args, calls commands.*, prints results, and converts expected failures (GoveeError) into clean user messages.

- tui.py
  - Textual-based interactive UI. Presentation and input handling only; delegates behavior to commands.* and persists theme via config.save_config.

- commands.py
  - Application business logic (profiles, validation, composing HA operations). Both CLI and TUI call into this module. Keep feature logic here.

- ha.py
  - The single network boundary: all HTTP calls to Home Assistant must happen here using requests. This isolates networking and makes it easy to mock for tests.

- config.py
  - Persists config at ~/.config/govee/config.json. On first run default values are written; user must add a Home Assistant long-lived access token (ha_token) manually.

- profiles.py
  - Default profile definitions and lookup (get_profile).

- models.py
  - Dataclasses (Config, Profile, LightState) and GoveeError used for expected/user-facing errors.

## Key conventions (repo-specific)

- Single-network-module rule: Only ha.py should make HTTP requests to HA. When adding features that touch HA, update ha.py and mock HomeAssistantClient in tests.

- UI thinness: CLI and TUI must remain presentation-only. Add or change behavior in commands.py.

- Error handling pattern: Raise GoveeError for expected/user-facing errors; CLI/TUI catch these and present friendly messages. Avoid surfacing raw tracebacks to users.

- Config lifecycle: config.load_config writes defaults on first run and will raise a GoveeError if ha_token is empty. CONFIG_PATH is ~/.config/govee/config.json.

- Validation lives in commands.py: follow existing validation (brightness 0-100, RGB 0-255, effect name whitelisting) when adding features.

- Host-affecting commands: PC_ACTIONS in commands.py calls systemctl (poweroff/suspend). These are destructive at the host level — tests should mock subprocess.run and avoid executing them in CI.

- Theme persistence: TUI stores the chosen theme back into config via save_config.

- Debugging/caveat: models.Config.headers() currently returns a masked Authorization value ('******'). If Home Assistant calls fail with 401, check and correct headers() so the Authorization header includes the token in the form HA expects (e.g. 'Authorization': f'Bearer {config.ha_token}').

## Where to change what

- Business logic / validation: commands.py
- Network/HTTP details: ha.py
- Config defaults/migrations: config.py
- Profiles: profiles.py
- UI: tui.py (Textual screens/widgets) and cli.py (Typer command registration)

## Assistant-specific checks

- Scanned for other assistant rules/configs (CLAUDE.md, AGENTS.md, .cursorrules, .clinerules, .windsurfrules, CONVENTIONS.md, AIDER_CONVENTIONS.md): none found.

---

If anything should be expanded (for example, note about preferred test framework, or explicit dependency pins), say which area to cover and Copilot will extend this file.
