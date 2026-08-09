"""Default profiles and profile lookup."""

from models import Config, GoveeError, Profile

DEFAULT_PROFILES: dict[str, Profile] = {
    "night": Profile(name="night", effect="Sleep", brightness_pct=15, pc_action_prompt=True,
                     device_overrides={"nanoleaf": Profile(name="night", effect="moonlight", brightness_pct=15)}),
    "gaming": Profile(name="gaming", effect="Rainbow", brightness_pct=100,
                      device_overrides={"nanoleaf": Profile(name="gaming", effect="Color Burst", brightness_pct=100)}),
    "movie": Profile(name="movie", effect="Candle", brightness_pct=20,
                     device_overrides={"nanoleaf": Profile(name="movie", effect="Romantic", brightness_pct=20)}),
    "work": Profile(name="work", effect=None, brightness_pct=100, rgb_color=(255, 255, 255),
                    device_overrides={"nanoleaf": Profile(name="work", effect="Triluminox Energy Crystal", brightness_pct=100)}),
    "relax": Profile(name="relax", effect="Ocean", brightness_pct=40,
                     device_overrides={"nanoleaf": Profile(name="relax", effect="Forest", brightness_pct=40)}),
}


def get_profile(config: Config, name: str) -> Profile:
    profile = config.profiles.get(name)
    if profile is None:
        available = ", ".join(sorted(config.profiles)) or "(none configured)"
        raise GoveeError(f"No profile named '{name}'. Available: {available}")
    return profile
