from __future__ import annotations

from pulse_agent.security import redact


def test_redact_covers_required_secret_keys() -> None:
    payload = {
        "api_key": "secret-api-key",
        "authorization": "Bearer secret",
        "cookie": "rp_access=secret",
        "password": "secret-password",
        "secret": "secret",
        "token": "secret-token",
        "vless_url": "vless://secret",
        "x-pulse-service-token": "service-secret",
        "nested": {
            "telegram_bot_token": "nested-secret",
            "public_value": "visible",
        },
        "items": [{"client_secret": "item-secret"}, {"name": "visible"}],
    }

    redacted = redact(payload)

    for key in (
        "api_key",
        "authorization",
        "cookie",
        "password",
        "secret",
        "token",
        "vless_url",
        "x-pulse-service-token",
    ):
        assert redacted[key] == "[redacted]"
    assert redacted["nested"]["telegram_bot_token"] == "[redacted]"
    assert redacted["nested"]["public_value"] == "visible"
    assert redacted["items"][0]["client_secret"] == "[redacted]"
    assert redacted["items"][1]["name"] == "visible"
