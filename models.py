"""Shared data shapes for the govee app. No behavior lives here."""

from dataclasses import dataclass, field
from typing import Optional


class GoveeError(Exception):
    """Raised for expected user-facing failures."""


@dataclass
class Profile:
    """A named lighting scene, optionally overridden per device."""

    name: str
    effect: Optional[str] = None
    brightness_pct: Optional[int] = None
    rgb_color: Optional[tuple[int, int, int]] = None
    pc_action_prompt: bool = False
    device_overrides: dict[str, "Profile"] = field(default_factory=dict)


@dataclass
class LightState:
    is_on: bool
    brightness_pct: Optional[int]
    effect: Optional[str]
    available_effects: list[str]


@dataclass
class Config:
    ha_base_url: str
    ha_token: str
    entity_id: str
    profiles: dict[str, Profile] = field(default_factory=dict)
    theme: str = "tokyo-night"
    devices: dict[str, str] = field(default_factory=dict)

    def headers(self) -> dict[str, str]:
        return {
            "Authorization": "Bearer " + self.ha_token,
            "Content-Type": "application/json",
        }
