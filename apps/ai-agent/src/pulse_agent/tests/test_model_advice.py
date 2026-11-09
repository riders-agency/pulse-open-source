from typing import Any

from pulse_agent.graph import advice, intent_schema, turn_nodes


def test_browser_context_cannot_override_authenticated_scope() -> None:
    context = turn_nodes.turn_page_context(
        {
            "user_id": "real-user",
            "project_id": "real-project",
            "company_id": "real-company",
            "session_id": "real-session",
            "page_context": {
                "user_id": "forged",
                "project_id": "forged",
                "projectId": "forged",
                "company_id": "forged",
                "session_id": "forged",
            },
        }
    )
    assert context == {
        "user_id": "real-user",
        "project_id": "real-project",
        "projectId": "real-project",
        "company_id": "real-company",
        "session_id": "real-session",
    }


def test_advice_never_runs_for_pending_or_write_operations(monkeypatch: Any) -> None:
    monkeypatch.setenv("PULSE_AI_MODEL_ADVICE", "true")
    for extra in (
        {"active_tool_call_id": "pending"},
        {"intent": {"intent_key": "task.create"}},
        {"intent": {"intent_key": "unsafe"}},
    ):
        assert advice.advice_for_turn({"body": "Помоги составить план", **extra}) is None


def test_advice_is_explicitly_text_only(monkeypatch: Any) -> None:
    monkeypatch.setenv("PULSE_AI_MODEL_ADVICE", "true")

    async def fake_request(state: dict[str, Any], body: str) -> dict[str, Any]:
        return {"content": "Начните с прототипа.", "usage": {"output_tokens": 5}}

    monkeypatch.setattr(advice, "_request_advice", fake_request)
    result = advice.advice_for_turn({"body": "Помоги составить план"})
    assert result is not None
    assert result["final_response"].startswith("Совет ИИ — без изменения данных:")
    assert "tool_calls" not in result


def test_explicit_advice_cannot_be_reinterpreted_as_data_search(monkeypatch: Any) -> None:
    def unexpected_classifier(*args: Any, **kwargs: Any) -> None:
        raise AssertionError("Explicit advice must not invoke intent classification")

    monkeypatch.setattr(intent_schema, "model_fallback_classification", unexpected_classifier)
    result = intent_schema.classify_turn("Помоги составить краткий план запуска лендинга: три шага.")
    assert result.reason == "explicit_advice"
    assert result.policy.allowed_tools == []
    assert result.tool_candidates == []
    # A recognized command still goes through the normal approval path.
    command = intent_schema.classify_turn("Создай задачу Подготовить демонстрацию завтра")
    assert command.intent_key == intent_schema.IntentKey.TASK_CREATE
    assert command.policy.write_serial_required
