"""Load and save ~/.config/govee/config.json.

On first run, writes out defaults (your existing HA URL/entity plus the
five default profiles) so there's something sensible to edit. A Home
Assistant long-lived access token is the one thing you have to add by
hand - there's no way to mint that non-interactively.
"""

import json
from pathlib import Path

from models import Config, GoveeError, Profile
from profiles import DEFAULT_PROFILES

CONFIG_DIR = Path.home() / ".config" / "govee"
CONFIG_PATH = CONFIG_DIR / "config.json"

DEFAULT_HA_BASE_URL = "http://100.88.207.105:8123"
DEFAULT_ENTITY_ID = "light.rgbic_strip_lights"
DEFAULT_DEVICES = {"govee": DEFAULT_ENTITY_ID, "nanoleaf": "light.light_panels_55_e4_27"}


def _default_config() -> Config:
    return Config(
        ha_base_url=DEFAULT_HA_BASE_URL,
        ha_token="",
        entity_id=DEFAULT_ENTITY_ID,
        theme="tokyo-night",
        profiles={
            name: Profile(
                name=p.name,
                effect=p.effect,
                brightness_pct=p.brightness_pct,
                rgb_color=p.rgb_color,
                pc_action_prompt=p.pc_action_prompt,
                device_overrides=p.device_overrides.copy(),
            )
            for name, p in DEFAULT_PROFILES.items()
        },
        devices=DEFAULT_DEVICES.copy(),
    )


def _read_config() -> Config:
    try:
        raw = json.loads(CONFIG_PATH.read_text())
    except (json.JSONDecodeError, OSError) as exc:
        raise GoveeError(f"Could not read config at {CONFIG_PATH}: {exc}") from exc

    profiles = {}
    for name, data in raw.get("profiles", {}).items():
        default_profile = DEFAULT_PROFILES.get(name)
        rgb = data.get("rgb_color")
        profiles[name] = Profile(
            name=data["name"],
            effect=data.get("effect"),
            brightness_pct=data.get("brightness_pct"),
            rgb_color=tuple(rgb) if rgb else None,
            pc_action_prompt=data.get("pc_action_prompt", False),
            device_overrides={
                device: Profile(
                    name=name,
                    effect=override.get("effect"),
                    brightness_pct=override.get("brightness_pct"),
                    rgb_color=tuple(override["rgb_color"]) if override.get("rgb_color") else None,
                )
                for device, override in (
                    data.get("devices")
                    or (
                        {
                            device: {
                                "effect": override.effect,
                                "brightness_pct": override.brightness_pct,
                                "rgb_color": list(override.rgb_color)
                                if override.rgb_color
                                else None,
                            }
                            for device, override in default_profile.device_overrides.items()
                        }
                        if default_profile
                        else {}
                    )
                ).items()
            },
        )

    return Config(
        ha_base_url=raw["ha_base_url"],
        ha_token=raw.get("ha_token", ""),
        entity_id=raw["entity_id"],
        profiles=profiles,
        theme=raw.get("theme", "tokyo-night"),
        devices={
            "govee": raw["entity_id"],
            "nanoleaf": DEFAULT_DEVICES["nanoleaf"],
            **(raw.get("devices") or {}),
        },
    )


def load_config() -> Config:
    if not CONFIG_PATH.exists():
        cfg = _default_config()
        save_config(cfg)
    else:
        cfg = _read_config()
        save_config(cfg)

    if not cfg.ha_token:
        raise GoveeError(
            f"No Home Assistant token set. Generate a long-lived access "
            f"token from your HA profile page (bottom of the page, "
            f"'Long-Lived Access Tokens') and add it as ha_token in "
            f"{CONFIG_PATH}, then run this again."
        )
    return cfg


def save_config(config: Config) -> None:
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    payload = {
        "ha_base_url": config.ha_base_url,
        "ha_token": config.ha_token,
        "entity_id": config.entity_id,
        "devices": config.devices,
        "theme": config.theme,
        "profiles": {
            name: {
                "name": p.name,
                "effect": p.effect,
                "brightness_pct": p.brightness_pct,
                "rgb_color": list(p.rgb_color) if p.rgb_color else None,
                "pc_action_prompt": p.pc_action_prompt,
                "devices": {
                    device: {
                        "effect": override.effect,
                        "brightness_pct": override.brightness_pct,
                        "rgb_color": list(override.rgb_color) if override.rgb_color else None,
                    }
                    for device, override in p.device_overrides.items()
                },
            }
            for name, p in config.profiles.items()
        },
    }
    CONFIG_PATH.write_text(json.dumps(payload, indent=2) + "\n")
