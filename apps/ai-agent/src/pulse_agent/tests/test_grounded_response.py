from __future__ import annotations

from typing import Any

from pulse_agent.grounded_response import GROUNDED_RESPONSE_SCHEMA, build_grounded_response


def test_grounded_response_uses_observation_evidence_and_entity_refs() -> None:
    response = build_grounded_response(
        {
            "run_id": "run-1",
            "tool_results": [{"ok": True}],
            "tool_observations": [
                observation(
                    "client.search",
                    {"answer": "Нашел клиента Орион", "items": [{"id": "client-1", "name": "Орион"}]},
                    run_id="run-1",
                    audit_ref="audit-client",
                    entity_refs=[
                        {
                            "entity_type": "client",
                            "entity_id": "client-1",
                            "label": "Орион",
                            "query": "Орион",
                            "source": "client.search",
                        },
                        {
                            "entity_type": "client",
                            "entity_id": "client-1",
                            "label": "Орион",
                            "query": "Орион",
                            "source": "client.search",
                        },
                    ],
                )
            ],
        },
        fallback="Нашел 10 клиентов",
    )

    assert response["schema"] == GROUNDED_RESPONSE_SCHEMA
    assert response["response_status"] == "answered"
    assert response["final_response"] == "Нашел клиента Орион"
    assert response["grounding"]["fallback_discarded"] is True
    assert response["evidence_refs"] == [
        {
            "tool_key": "client.search",
            "tool_call_id": "",
            "step_index": 1,
            "status": "succeeded",
            "ok": True,
            "audit_ref": "audit-client",
            "answer": "Нашел клиента Орион",
            "error_code": "",
            "entity_refs": [
                {
                    "entity_type": "client",
                    "entity_id": "client-1",
                    "label": "Орион",
                    "aliases": ["Орион"],
                    "source": "client.search",
                },
                {
                    "entity_type": "client",
                    "entity_id": "client-1",
                    "label": "Орион",
                    "aliases": ["Орион"],
                    "source": "client.search",
                },
            ],
        }
    ]
    assert response["entity_refs"] == [
        {
            "entity_type": "client",
            "entity_id": "client-1",
            "label": "Орион",
            "aliases": ["Орион"],
            "source": "client.search",
        }
    ]


def test_grounded_response_discards_unverified_read_fallback() -> None:
    response = build_grounded_response(
        {
            "run_id": "run-new",
            "active_mode": "project_read",
            "current_step": "run_read_tools",
            "awaiting_tool_observation": True,
            "tool_results": [],
            "tool_observations": [
                observation(
                    "client.search",
                    {"answer": "Старый ответ про клиента"},
                    run_id="run-old",
                    audit_ref="audit-old",
                )
            ],
        },
        fallback="Нашел 5 клиентов",
    )

    assert response["response_status"] == "no_result"
    assert "Нашел 5" not in response["final_response"]
    assert "Старый ответ" not in response["final_response"]
    assert response["evidence_refs"] == []
    assert response["grounding"]["fallback_discarded"] is True


def test_grounded_response_partial_and_tool_error_templates() -> None:
    partial = build_grounded_response(
        {
            "run_id": "run-1",
            "tool_results": [{"ok": True}, {"ok": False}],
            "tool_observations": [
                observation("task.search", {"answer": "Нашел 2 задачи"}, run_id="run-1"),
                observation(
                    "payment.lookup_due",
                    {},
                    ok=False,
                    status="failed",
                    run_id="run-1",
                    error_code="tool_gateway_error",
                    user_message="Оплаты временно недоступны",
                ),
            ],
        },
        fallback="fallback",
    )

    assert partial["response_status"] == "partial"
    assert "Нашел 2 задачи" in partial["final_response"]
    assert "Часть данных не получена: Оплаты временно недоступны" in partial["final_response"]

    failed = build_grounded_response(
        {
            "run_id": "run-1",
            "tool_results": [{"ok": False}],
            "tool_observations": [
                observation(
                    "task.search",
                    {},
                    ok=False,
                    status="failed",
                    run_id="run-1",
                    error_code="tool_gateway_error",
                    user_message="Tool Gateway в ограниченном режиме",
                )
            ],
        },
        fallback="Нашел задачу",
    )

    assert failed["response_status"] == "tool_error"
    assert failed["final_response"] == "Не удалось получить подтвержденные данные: Tool Gateway в ограниченном режиме"
    assert "Нашел задачу" not in failed["final_response"]


def test_grounded_response_no_result_template_for_empty_success() -> None:
    response = build_grounded_response(
        {
            "run_id": "run-1",
            "tool_results": [{"ok": True}],
            "tool_observations": [observation("task.search", {"items": []}, run_id="run-1")],
        },
        fallback="Нашел задачи",
    )

    assert response["response_status"] == "no_result"
    assert response["final_response"] == "По подтвержденным данным результата нет."


def observation(
    tool_key: str,
    normalized_output: dict[str, Any],
    *,
    ok: bool = True,
    status: str = "succeeded",
    run_id: str = "",
    audit_ref: str = "",
    entity_refs: list[dict[str, Any]] | None = None,
    error_code: str = "",
    user_message: str = "",
) -> dict[str, Any]:
    return {
        "schema": "riders-pulse-tool-observation-v1",
        "tool_key": tool_key,
        "step_index": 1,
        "run_id": run_id,
        "turn_id": "",
        "status": status,
        "ok": ok,
        "normalized_output": normalized_output,
        "entity_refs": entity_refs or [],
        "audit_ref": audit_ref,
        "error_code": error_code,
        "error_message": "",
        "user_message": user_message,
    }
