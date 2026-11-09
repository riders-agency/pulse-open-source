from __future__ import annotations

SECRET_KEYS = {
    "api_key",
    "authorization",
    "cookie",
    "password",
    "secret",
    "token",
    "vless_url",
    "x-pulse-service-token",
}


def redact(value: object) -> object:
    if isinstance(value, dict):
        out: dict[str, object] = {}
        for key, item in value.items():
            if key.lower() in SECRET_KEYS or any(secret in key.lower() for secret in SECRET_KEYS):
                out[key] = "[redacted]"
            else:
                out[key] = redact(item)
        return out
    if isinstance(value, list):
        return [redact(item) for item in value]
    return value


def assert_project_scope(company_id: str, project_id: str, user_id: str) -> None:
    if not company_id or not project_id or not user_id:
        raise PermissionError("company_id, project_id and user_id are required")
