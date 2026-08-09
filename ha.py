"""Home Assistant client - the only module allowed to touch the network.

This calls HA's native REST API (the light.turn_on/turn_off services
and the /states endpoint) using a long-lived access token, instead of
one fixed webhook per action. It still goes entirely through Home
Assistant -> govee2mqtt -> the lights; the PC never talks to the
lights directly, and no HA automations/scripts need to change.
"""

import requests

from models import Config, GoveeError, LightState


class HomeAssistantClient:
    def __init__(self, config: Config, timeout: float = 5.0):
        self._config = config
        self._timeout = timeout

    def _request(self, method: str, path: str, json_body: dict | None = None):
        url = f"{self._config.ha_base_url.rstrip('/')}{path}"
        try:
            response = requests.request(
                method,
                url,
                headers=self._config.headers(),
                json=json_body,
                timeout=self._timeout,
            )
        except requests.RequestException as exc:
            raise GoveeError(f"Could not reach Home Assistant at {url}: {exc}") from exc

        if response.status_code == 401:
            raise GoveeError("Home Assistant rejected the request - check your ha_token.")
        if not response.ok:
            raise GoveeError(f"Home Assistant returned HTTP {response.status_code} for {path}")

        return response

    def turn_on(
        self,
        entity_id: str | None = None,
        effect: str | None = None,
        brightness_pct: int | None = None,
        rgb_color: tuple[int, int, int] | None = None,
    ) -> None:
        data: dict = {"entity_id": entity_id or self._config.entity_id}
        if brightness_pct is not None:
            data["brightness_pct"] = brightness_pct
        if effect is not None:
            data["effect"] = effect
        if rgb_color is not None:
            data["rgb_color"] = list(rgb_color)
        self._request("POST", "/api/services/light/turn_on", data)

    def turn_off(self, entity_id: str | None = None) -> None:
        self._request(
            "POST",
            "/api/services/light/turn_off",
            {"entity_id": entity_id or self._config.entity_id},
        )

    def get_state(self, entity_id: str | None = None) -> LightState:
        target = entity_id or self._config.entity_id
        response = self._request("GET", f"/api/states/{target}")
        payload = response.json()
        attributes = payload.get("attributes", {})
        brightness = attributes.get("brightness")
        return LightState(
            is_on=payload.get("state") == "on",
            brightness_pct=round(brightness / 255 * 100) if brightness else None,
            effect=attributes.get("effect"),
            available_effects=attributes.get("effect_list", []),
        )
