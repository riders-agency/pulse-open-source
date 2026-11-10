from __future__ import annotations

import asyncio
import os
import re
from typing import Any

from pulse_agent.clients.pulse_model_gateway import PulseModelGatewayClient, PulseModelGatewayError
from pulse_agent.config import Settings


def is_advice_request(body: str) -> bool:
    return bool(re.match(r"^(помоги|предложи|объясни|расскажи|как\b|привет|здравствуй)", body.strip(), re.IGNORECASE))


def advice_for_turn(state: dict[str, Any]) -> dict[str, Any] | None:
    if os.environ.get("PULSE_AI_MODEL_ADVICE") != "true" or state.get("active_tool_call_id"):
        return None
    body = str(state.get("body") or state.get("current_body") or "").strip()
    intent = state.get("intent") or {}
    key = str(intent.get("intent_key") or "unknown")
    if key not in {"conversation.smalltalk", "unknown"}:
        return None
    if not is_advice_request(body):
        return None
    try:
        result = asyncio.run(_request_advice(state, body[:6000]))
    except (PulseModelGatewayError, RuntimeError, ValueError):
        return {
            "final_response": ("Ответ модели сейчас недоступен или достигнут лимит демо. "
                               "Задачи можно создавать через карточки.")
        }
    content = str(result.get("content") or "").strip()
    if not content:
        return None
    return {"final_response": "Совет ИИ — без изменения данных:\n\n" + content, "usage": result.get("usage") or {}}


async def _request_advice(state: dict[str, Any], body: str) -> dict[str, Any]:
    settings = Settings()
    client = PulseModelGatewayClient(settings.pulse_model_grpc_addr, settings.service_token)
    try:
        return await client.final_response(
            [
                {
                    "role": "system",
                    "content": (
                        "Ты помощник команды Pulse. Ответь по-русски кратко и по существу. "
                        "В этом режиме ты только предлагаешь идеи и объясняешь. У тебя нет инструментов, "
                        "ключей, текущих данных проекта или результатов операций. Не утверждай, что создал, "
                        "изменил, отправил или проверил что-либо. Для изменения задач предложи запросить "
                        "карточку и подтвердить её в интерфейсе. Не выдумывай сведения о проекте."
                    ),
                },
                {"role": "user", "content": body},
            ],
            company_id=str(state.get("company_id") or ""),
            project_id=str(state.get("project_id") or ""),
            user_id=str(state.get("user_id") or ""),
            session_id=str(state.get("session_id") or ""),
            max_tokens=650,
            temperature=0.3,
        )
    finally:
        await client.close()
