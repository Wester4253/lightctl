"""Interactive TUI. Calls the same commands.py functions the CLI does -
no light-control logic lives here, just rendering and input handling.
"""

from __future__ import annotations

from textual.app import App, ComposeResult
from textual.containers import Vertical
from textual.screen import Screen
from textual.widgets import Footer, Header, Input, Label, ListItem, ListView, Static

import commands
from config import CONFIG_PATH, load_config, save_config
from models import Config, GoveeError, LightState

CLOCK_FORMAT = "%I:%M %p"  # 12-hour, minutes, no seconds

MENU_ITEMS = [
    ("control", "Control Panel"),
    ("night", "Night Mode"),
    ("gaming", "Gaming Mode"),
    ("movie", "Movie Mode"),
    ("work", "Work Mode"),
    ("relax", "Relax Mode"),
    ("effects", "Effects"),
    ("brightness", "Brightness"),
    ("colors", "Colors"),
    ("power", "Power"),
    ("theme", "Theme"),
    ("settings", "Settings"),
]

PROFILE_KEYS = {"night", "gaming", "movie", "work", "relax"}


class StatusPanel(Static):
    def show(
        self,
        states: dict[str, LightState] | None,
        error: str | None = None,
    ) -> None:
        if error:
            self.update(f"[$error]●[/] {error}")
            return
        if states is None:
            self.update("[$foreground 50%]●[/] loading...")
            return
        lines = []
        for device, state in states.items():
            dot = "[$success]●[/]" if state.is_on else "[$foreground 50%]●[/]"
            brightness = (
                f"{state.brightness_pct}%"
                if state.brightness_pct is not None
                else "-"
            )
            lines.append(
                f"{dot} [bold]{device}: {'Online' if state.is_on else 'Off'}[/bold]   "
                f"[$secondary]Effect[/] {state.effect or '-'}   "
                f"[$secondary]Brightness[/] {brightness}"
            )
        self.update("\n".join(lines))


class DismissibleScreen(Screen):
    BINDINGS = [
        ("escape", "dismiss_screen", "Back"),
        ("q", "dismiss_screen", "Back"),
    ]

    def action_dismiss_screen(self) -> None:
        self.dismiss()


class ControlPanelScreen(DismissibleScreen):
    """Live overview of every configured light and its capabilities."""

    def __init__(self, config: Config):
        super().__init__()
        self._config = config

    def compose(self) -> ComposeResult:
        yield Header()
        yield Static("Loading device statistics...", id="control-stats")
        yield Footer()

    def on_mount(self) -> None:
        self._refresh()

    def key_r(self) -> None:
        self._refresh()

    def _refresh(self) -> None:
        try:
            states = commands.get_light_states(self._config)
        except GoveeError as exc:
            self.query_one("#control-stats", Static).update(f"[$error]●[/] {exc}")
            return
        total = len(states)
        on_count = sum(state.is_on for state in states.values())
        lines = [
            f"[bold]Lighting Control Panel[/bold]  •  {on_count}/{total} lights on",
            "",
        ]
        for name, state in states.items():
            entity = self._config.devices.get(name, self._config.entity_id)
            brightness = (
                f"{state.brightness_pct}%"
                if state.brightness_pct is not None
                else "unknown"
            )
            lines.extend(
                [
                    f"[bold]{name.title()}[/bold]  {'ON' if state.is_on else 'OFF'}",
                    f"  Entity: {entity}",
                    f"  Effect: {state.effect or 'none'}",
                    f"  Brightness: {brightness}",
                    f"  Available effects: {len(state.available_effects)}",
                    "",
                ]
            )
        lines.append("Press [bold]r[/bold] to refresh  •  [bold]Esc[/bold]/[bold]q[/bold] to go back")
        self.query_one("#control-stats", Static).update("\n".join(lines))


class EffectScreen(DismissibleScreen):
    """Browse and pick from the effects Home Assistant reports as available."""

    def __init__(self, config: Config, effects_by_device: dict[str, list[str]]):
        super().__init__()
        self._config = config
        self._effects_by_device = effects_by_device

    def compose(self) -> ComposeResult:
        yield Header()
        items = [
            ListItem(Label(f"{device}: {effect}"), name=f"{device}:{effect}")
            for device, effects in self._effects_by_device.items()
            for effect in effects
        ]
        if items:
            yield ListView(*items)
        else:
            yield Label("No effects reported by Home Assistant.")
        yield Footer()

    def on_list_view_selected(self, event: ListView.Selected) -> None:
        device, effect_name = event.item.name.split(":", 1)
        try:
            commands.set_device_effect(self._config, device, effect_name)
        except GoveeError as exc:
            self.notify(str(exc), severity="error")
        else:
            self.notify(f"Effect set to {effect_name}")
        self.dismiss()


class BrightnessScreen(DismissibleScreen):
    def __init__(self, config: Config):
        super().__init__()
        self._config = config

    def compose(self) -> ComposeResult:
        yield Header()
        yield Vertical(
            Label("Brightness, 0-100, then Enter:"),
            Input(placeholder="50"),
        )
        yield Footer()

    def on_input_submitted(self, event: Input.Submitted) -> None:
        try:
            pct = int(event.value)
            commands.set_brightness(self._config, pct)
        except (ValueError, GoveeError) as exc:
            self.notify(str(exc), severity="error")
        else:
            self.notify(f"Brightness set to {pct}%")
        self.dismiss()


class ColorScreen(DismissibleScreen):
    def __init__(self, config: Config):
        super().__init__()
        self._config = config

    def compose(self) -> ComposeResult:
        yield Header()
        yield Vertical(
            Label("RGB, space-separated (e.g. 255 0 255), then Enter:"),
            Input(placeholder="255 0 255"),
        )
        yield Footer()

    def on_input_submitted(self, event: Input.Submitted) -> None:
        try:
            r, g, b = (int(part) for part in event.value.split())
            commands.set_color(self._config, r, g, b)
        except (ValueError, GoveeError) as exc:
            self.notify(str(exc), severity="error")
        else:
            self.notify(f"Color set to ({r}, {g}, {b})")
        self.dismiss()


class ThemeScreen(DismissibleScreen):
    """Pick from every theme Textual ships - Tokyo Night, Catppuccin, Nord,
    Dracula, Gruvbox, Solarized, and the rest - and persist the choice.
    """

    def __init__(self, config: Config):
        super().__init__()
        self._config = config

    def compose(self) -> ComposeResult:
        yield Header()
        yield ListView(
            *(
                ListItem(Label(name), name=name)
                for name in sorted(self.app.available_themes)
            )
        )
        yield Footer()

    def on_list_view_selected(self, event: ListView.Selected) -> None:
        theme_name = event.item.name
        self.app.theme = theme_name
        self._config.theme = theme_name
        save_config(self._config)
        self.notify(f"Theme set to {theme_name}")
        self.dismiss()


class SettingsScreen(DismissibleScreen):
    def __init__(self, config: Config):
        super().__init__()
        self._config = config

    def compose(self) -> ComposeResult:
        yield Header()
        yield Vertical(
            Label(f"Config file: {CONFIG_PATH}"),
            Label(f"Home Assistant: {self._config.ha_base_url}"),
            Label(f"Entity: {self._config.entity_id}"),
            Label(
                "Devices: "
                + ", ".join(
                    f"{name}={entity_id}"
                    for name, entity_id in self._config.devices.items()
                )
            ),
            Label(f"Theme: {self._config.theme}"),
            Label(f"Profiles: {', '.join(sorted(self._config.profiles))}"),
            Label(""),
            Label("Edit config.json directly to change these or add profiles."),
        )
        yield Footer()


class GoveeApp(App):
    CSS = """
    Screen {
        background: $background;
        color: $foreground;
    }

    Header {
        background: $panel;
        color: $primary;
        text-style: bold;
    }

    Footer {
        background: $panel;
        color: $foreground;
    }

    StatusPanel {
        background: $surface;
        border: round $panel;
        padding: 1 2;
        margin: 1 1 0 1;
        height: auto;
    }

    ListView {
        background: $background;
        margin: 1;
        border: round $panel;
        height: auto;
    }

    ListView > ListItem {
        padding: 0 2;
        color: $foreground;
    }

    ListView > ListItem.-hovered {
        background: $surface;
    }

    ListView > ListItem.-highlight {
        background: $primary 30%;
        color: $foreground;
        text-style: bold;
    }

    ListView:focus > ListItem.-highlight {
        background: $primary;
        color: $background;
    }

    Vertical {
        margin: 1 2;
        border: round $panel;
        padding: 1 2;
        height: auto;
    }

    Vertical Label {
        color: $foreground;
        margin-bottom: 1;
    }

    Input {
        border: round $secondary;
        background: $surface;
        color: $foreground;
    }

    Input:focus {
        border: round $primary;
    }
    """
    BINDINGS = [("q", "quit", "Quit")]
    TITLE = "lightctl"
    SUB_TITLE = "Home Assistant lighting control"

    def __init__(self):
        super().__init__()
        self._config: Config | None = None

    def compose(self) -> ComposeResult:
        yield Header(show_clock=True, time_format=CLOCK_FORMAT)
        yield StatusPanel(id="status")
        yield ListView(*(ListItem(Label(label), name=key) for key, label in MENU_ITEMS))
        yield Footer()

    def on_mount(self) -> None:
        try:
            self._config = load_config()
        except GoveeError as exc:
            self.query_one(StatusPanel).show(None, error=str(exc))
            return

        if self._config.theme in self.available_themes:
            self.theme = self._config.theme
        self._refresh_status()

    def _refresh_status(self) -> None:
        panel = self.query_one(StatusPanel)
        try:
            states = commands.get_light_states(self._config)
        except GoveeError as exc:
            panel.show(None, error=str(exc))
        else:
            panel.show(states)

    def on_list_view_selected(self, event: ListView.Selected) -> None:
        if self._config is None:
            return
        key = event.item.name

        if key == "control":
            self.push_screen(
                ControlPanelScreen(self._config),
                callback=lambda _: self._refresh_status(),
            )

        elif key in PROFILE_KEYS:
            try:
                commands.apply_profile(self._config, key)
            except GoveeError as exc:
                self.notify(str(exc), severity="error")
            else:
                self.notify(f"Activated {key}")
                self._refresh_status()

        elif key == "effects":
            try:
                states = commands.get_light_states(self._config)
            except GoveeError as exc:
                self.notify(str(exc), severity="error")
            else:
                self.push_screen(
                    EffectScreen(
                        self._config,
                        {
                            device: state.available_effects
                            for device, state in states.items()
                        },
                    ),
                    callback=lambda _: self._refresh_status(),
                )

        elif key == "brightness":
            self.push_screen(
                BrightnessScreen(self._config),
                callback=lambda _: self._refresh_status(),
            )

        elif key == "colors":
            self.push_screen(
                ColorScreen(self._config),
                callback=lambda _: self._refresh_status(),
            )

        elif key == "power":
            try:
                state = commands.get_light_state(self._config)
                if state.is_on:
                    commands.turn_off(self._config)
                else:
                    commands.turn_on(self._config)
            except GoveeError as exc:
                self.notify(str(exc), severity="error")
            else:
                self._refresh_status()

        elif key == "theme":
            self.push_screen(ThemeScreen(self._config))

        elif key == "settings":
            self.push_screen(SettingsScreen(self._config))


if __name__ == "__main__":
    GoveeApp().run()
