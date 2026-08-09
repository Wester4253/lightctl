"""The actual actions govee can take. CLI and TUI both call into here -
neither should contain business logic of its own.
"""

import subprocess

from ha import HomeAssistantClient
from models import Config, GoveeError, LightState, Profile
from profiles import get_profile

PC_ACTIONS = {
    "1": ("Shutting down...", ["systemctl", "poweroff"]),
    "2": ("Sleeping...", ["systemctl", "suspend"]),
    "3": ("Ignoring PC action.", None),
}

def _devices(config: Config) -> dict[str, str]:
    return config.devices or {"govee": config.entity_id}


def apply_profile(config: Config, name: str) -> Profile:
    """Apply a profile's effect/brightness/color in one call. Returns
    the profile so the caller (CLI/TUI) can decide whether to show the
    PC action prompt.
    """
    profile = get_profile(config, name)
    client = HomeAssistantClient(config)
    for device_name, entity_id in _devices(config).items():
        settings = profile.device_overrides.get(device_name, profile)
        client.turn_on(
            entity_id=entity_id,
            effect=settings.effect,
            brightness_pct=settings.brightness_pct,
            rgb_color=settings.rgb_color,
        )
    return profile


def set_effect(config: Config, effect: str) -> None:
    client = HomeAssistantClient(config)
    matched = False
    for entity_id in _devices(config).values():
        state = client.get_state(entity_id)
        if state.available_effects and effect not in state.available_effects:
            continue
        client.turn_on(entity_id=entity_id, effect=effect)
        matched = True
    if not matched:
        raise GoveeError(f"Unknown effect '{effect}' on configured lights.")


def get_light_states(config: Config) -> dict[str, LightState]:
    client = HomeAssistantClient(config)
    return {
        name: client.get_state(entity_id)
        for name, entity_id in _devices(config).items()
    }


def set_device_effect(config: Config, device_name: str, effect: str) -> None:
    devices = _devices(config)
    if device_name not in devices:
        raise GoveeError(f"Unknown device '{device_name}'.")
    client = HomeAssistantClient(config)
    state = client.get_state(devices[device_name])
    if state.available_effects and effect not in state.available_effects:
        available = ", ".join(state.available_effects)
        raise GoveeError(
            f"Unknown effect '{effect}' for {device_name}. Available: {available}"
        )
    client.turn_on(entity_id=devices[device_name], effect=effect)


def set_brightness(config: Config, pct: int) -> None:
    if not 0 <= pct <= 100:
        raise GoveeError(f"Brightness must be 0-100, got {pct}")
    client = HomeAssistantClient(config)
    for entity_id in _devices(config).values():
        client.turn_on(entity_id=entity_id, brightness_pct=pct)


def set_color(config: Config, r: int, g: int, b: int) -> None:
    for value, label in ((r, "red"), (g, "green"), (b, "blue")):
        if not 0 <= value <= 255:
            raise GoveeError(f"{label} must be 0-255, got {value}")
    client = HomeAssistantClient(config)
    for entity_id in _devices(config).values():
        client.turn_on(entity_id=entity_id, rgb_color=(r, g, b))


def turn_on(config: Config) -> None:
    client = HomeAssistantClient(config)
    for entity_id in _devices(config).values():
        client.turn_on(entity_id=entity_id)


def turn_off(config: Config) -> None:
    client = HomeAssistantClient(config)
    for entity_id in _devices(config).values():
        client.turn_off(entity_id=entity_id)


def get_light_state(config: Config) -> LightState:
    return HomeAssistantClient(config).get_state()


def run_pc_action(choice: str) -> str:
    """Execute the shutdown/sleep/ignore choice. Returns a message to print."""
    if choice not in PC_ACTIONS:
        raise GoveeError(f"Invalid option '{choice}'. Choose 1, 2, or 3.")

    message, cmd = PC_ACTIONS[choice]
    if cmd is not None:
        subprocess.run(cmd, check=True)
    return message
