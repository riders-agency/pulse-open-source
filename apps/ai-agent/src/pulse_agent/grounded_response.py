from __future__ import annotations

from typing import Any

from pulse_agent.graph.observations import OBSERVATION_SCHEMA

GROUNDED_RESPONSE_SCHEMA = "riders-pulse-grounded-final-response-v1"

READ_MODES = {"project_read"}
READ_INTENTS = {
    "client.count",
    "client.search",
    "payment.lookup_due",
    "question.answer",
    "task.search",
    "workflow.answer_question",
}


def build_grounded_response(state: dict[str, Any], *, fallback: str = "") -> dict[str, Any]:
    observations = current_turn_observations(state)
    evidence_refs = evidence_refs_for(observations)
    entity_refs = entity_refs_for(observations)
    answers = answers_from_observations(observations)
    failures = [item for item in observations if not bool(item.get("ok"))]
    fallback_text = fallback.strip()

    if answers:
        response_status = "partial" if failures else "answered"
        final_response = "\n".join(f"- {answer}" for answer in answers) if len(answers) > 1 else answers[0]
        if failures:
            final_response = f"{final_response}\n\nЧасть данных не получена: {failure_summary(failures[-1])}"
    elif failures:
        response_status = "tool_error"
        final_response = f"Не удалось получить подтвержденные данные: {failure_summary(failures[-1])}"
    elif observations:
        response_status, final_response = no_answer_template(observations)
    elif read_response_required(state):
        response_status = "no_result"
        final_response = "Не получил подтвержденных данных по этому запросу. Уточните контекст или повторите запрос."
    else:
        response_status = "answered" if fallback_text else "no_result"
        final_response = fallback_text or "Готово."

    return {
        "schema": GROUNDED_RESPONSE_SCHEMA,
        "response_status": response_status,
        "final_response": final_response,
        "evidence_refs": evidence_refs,
        "entity_refs": entity_refs,
        "grounding": {
            "source": "tool_observations" if observations else "local_state",
            "fallback_used": not observations and not read_response_required(state),
            "fallback_discarded": bool(observations or read_response_required(state)) and bool(fallback_text),
        },
    }


def current_turn_observations(state: dict[str, Any]) -> list[dict[str, Any]]:
    observations = [
        item
        for item in list(state.get("tool_observations") or [])
        if isinstance(item, dict) and item.get("schema") == OBSERVATION_SCHEMA
    ]
    if not observations:
        return []
    run_id = str(state.get("run_id") or "")
    turn_id = str(state.get("turn_id") or "")
    marked: list[dict[str, Any]] = []
    for item in observations:
        if (turn_id and item.get("turn_id") == turn_id) or (run_id and item.get("run_id") == run_id):
            marked.append(item)
    if marked:
        return marked
    if any(item.get("turn_id") or item.get("run_id") for item in observations):
        return []
    tool_result_count = len([item for item in list(state.get("tool_results") or []) if isinstance(item, dict)])
    if tool_result_count > 0:
        return observations[-tool_result_count:]
    if state.get("awaiting_tool_observation"):
        return []
    return observations


def read_response_required(state: dict[str, Any]) -> bool:
    plan = state.get("plan")
    intent = str(plan.get("intent") or "") if isinstance(plan, dict) else ""
    active_mode = str(state.get("active_mode") or "")
    current_step = str(state.get("current_step") or "")
    return active_mode in READ_MODES or intent in READ_INTENTS or current_step == "run_read_tools"


def evidence_refs_for(observations: list[dict[str, Any]]) -> list[dict[str, Any]]:
    refs: list[dict[str, Any]] = []
    for item in observations:
        output = item.get("normalized_output")
        answer = ""
        if isinstance(output, dict):
            answer = str(output.get("answer") or output.get("message") or "").strip()
        refs.append(
            {
                "tool_key": str(item.get("tool_key") or ""),
                "tool_call_id": str(item.get("tool_call_id") or ""),
                "step_index": int_value(item.get("step_index")),
                "status": str(item.get("status") or ""),
                "ok": bool(item.get("ok")),
                "audit_ref": str(item.get("audit_ref") or ""),
                "answer": answer[:500],
                "error_code": str(item.get("error_code") or ""),
                "entity_refs": normalize_entity_refs(item.get("entity_refs")),
            }
        )
    return refs


def entity_refs_for(observations: list[dict[str, Any]]) -> list[dict[str, Any]]:
    refs: list[dict[str, Any]] = []
    for item in observations:
        refs.extend(normalize_entity_refs(item.get("entity_refs")))
    return dedupe_entity_refs(refs)


def normalize_entity_refs(value: Any) -> list[dict[str, Any]]:
    if not isinstance(value, list):
        return []
    refs: list[dict[str, Any]] = []
    for item in value:
        if not isinstance(item, dict):
            continue
        entity_type = str(item.get("entity_type") or "entity")
        entity_id = str(item.get("entity_id") or "")
        label = str(item.get("label") or "")
        query = str(item.get("query") or "")
        aliases = unique_strings([label, query])
        refs.append(
            {
                "entity_type": entity_type,
                "entity_id": entity_id,
                "label": label,
                "aliases": aliases,
                "source": str(item.get("source") or ""),
            }
        )
    return refs


def dedupe_entity_refs(refs: list[dict[str, Any]]) -> list[dict[str, Any]]:
    deduped: list[dict[str, Any]] = []
    seen: set[str] = set()
    for ref in refs:
        key = "|".join(
            (
                str(ref.get("entity_type") or ""),
                str(ref.get("entity_id") or ""),
                str(ref.get("label") or "").lower(),
                str(ref.get("source") or ""),
            )
        )
        if key in seen:
            continue
        seen.add(key)
        deduped.append(ref)
    return deduped


def answers_from_observations(observations: list[dict[str, Any]]) -> list[str]:
    answers: list[str] = []
    seen: set[str] = set()
    for item in observations:
        if item.get("supporting_only") or not item.get("ok"):
            continue
        output = item.get("normalized_output")
        answer = ""
        if isinstance(output, dict):
            answer = str(output.get("answer") or output.get("message") or "").strip()
        if not answer:
            answer = str(item.get("user_message") or "").strip()
        if answer and answer not in seen:
            seen.add(answer)
            answers.append(answer)
    return answers


def no_answer_template(observations: list[dict[str, Any]]) -> tuple[str, str]:
    item_count = 0
    for item in observations:
        if not item.get("ok"):
            continue
        output = item.get("normalized_output")
        if not isinstance(output, dict):
            continue
        items = output.get("items")
        if isinstance(items, list):
            item_count += len(items)
    if item_count > 0:
        return "partial", f"Получил подтвержденные записи: {item_count}. Уточните, какие поля вывести."
    return "no_result", "По подтвержденным данным результата нет."


def failure_summary(observation: dict[str, Any]) -> str:
    message = str(
        observation.get("user_message")
        or observation.get("error_message")
        or observation.get("error_code")
        or ""
    ).strip()
    if message:
        return message
    tool_key = str(observation.get("tool_key") or "tool")
    return f"{tool_key} вернул ошибку"


def unique_strings(values: list[str]) -> list[str]:
    out: list[str] = []
    seen: set[str] = set()
    for value in values:
        current = value.strip()
        if not current or current in seen:
            continue
        seen.add(current)
        out.append(current)
    return out


def int_value(value: Any) -> int:
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    if isinstance(value, str) and value.isdigit():
        return int(value)
    return 0
