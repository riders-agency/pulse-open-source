from __future__ import annotations

from typing import Any, cast

from pulse_agent.graph.events import add_event
from pulse_agent.graph.intent_routes import active_mode_for_intent
from pulse_agent.graph.observations import (
    MAX_OBSERVATIONS_PER_TURN,
    append_tool_observation,
    next_tool_input,
    normalize_tool_observation,
    record_tool_result_event,
    replan_trace_entry,
    summarize_observations,
)
from pulse_agent.graph.tool_gateway import run_read_tool_if_enabled


def read_context(state: dict[str, Any]) -> dict[str, Any]:
    body = str(state.get("body") or state.get("current_body") or "")
    plan = state.get("plan") or {}
    intent = str(plan.get("intent") or "question.answer")
    tool_keys = plan.get("required_tools")
    read_tools = [str(tool) for tool in tool_keys] if isinstance(tool_keys, list) and tool_keys else ["context.read"]
    events = list(state.get("events") or [])
    preview = plan.get("tool_input_preview")
    tool_input = dict(preview) if isinstance(preview, dict) else {}
    query = tool_input.get("query") if "query" in tool_input else body
    tool_input.update(
        {
            "query": str(query or ""),
            "page_context": state.get("page_context") or {},
            "client_context": state.get("client_context") or {},
            "intent": intent,
        }
    )
    add_event(events, state, "STATUS", "Шаг: plan_tools")
    tool_results: list[dict[str, Any]] = []
    turn_observations: list[dict[str, Any]] = []
    tool_observations = [item for item in list(state.get("tool_observations") or []) if isinstance(item, dict)]
    replan_trace = [item for item in list(state.get("replan_trace") or []) if isinstance(item, dict)]
    for tool_key in read_tools[:MAX_OBSERVATIONS_PER_TURN]:
        add_event(
            events,
            state,
            "TOOL_PROPOSED",
            f"Запланировать read tool: {tool_key}",
            tool_call={"tool_key": tool_key, "risk_level": "read", "status": "proposed", "input_json": tool_input},
        )
    for index, tool_key in enumerate(read_tools[:MAX_OBSERVATIONS_PER_TURN]):
        step_input = next_tool_input(tool_input, tool_key, tool_observations)
        add_event(
            events,
            state,
            "TOOL_STARTED",
            "Читаю контекст через Go Tool Gateway",
            tool_call={"tool_key": tool_key, "risk_level": "read", "status": "running", "input_json": step_input},
        )
        observation = run_read_tool_if_enabled(state, tool_key, step_input)
        if observation is not None:
            tool_results.append(observation)
            normalized = normalize_tool_observation(
                tool_key,
                observation,
                step_input,
                step_index=len(turn_observations) + 1,
                run_id=str(state.get("run_id") or ""),
                turn_id=str(state.get("turn_id") or ""),
            )
            next_key = read_tools[index + 1] if index + 1 < len(read_tools) else ""
            if next_key and tool_key == "client.search":
                normalized["supporting_only"] = True
            turn_observations = append_tool_observation(turn_observations, normalized)
            tool_observations = append_tool_observation(tool_observations, normalized)
            replan_trace.append(replan_trace_entry(tool_key, turn_observations, next_key))
            record_tool_result_event(events, state, normalized, "Read tool выполнен")
    return {
        "events": events,
        "active_mode": active_mode_for_intent(intent),
        "current_step": "run_read_tools",
        "awaiting_tool_observation": len(tool_results) == 0,
        "tool_results": tool_results,
        "tool_observations": tool_observations,
        "replan_trace": replan_trace,
        "final_response": summarize_observations(turn_observations, str(plan.get("user_visible_summary") or "")),
    }


def workflow_question(state: dict[str, Any]) -> dict[str, Any]:
    body = str(state.get("body") or state.get("current_body") or "")
    events = list(state.get("events") or [])
    raw_page_context = state.get("page_context")
    page_context = cast(dict[str, Any], raw_page_context) if isinstance(raw_page_context, dict) else {}
    tool_input = {
        "query": body,
        "workflow_id": page_context.get("workflowId") or page_context.get("workflow_id") or "",
        "page_context": page_context,
        "intent": "workflow.answer_question",
    }
    add_event(events, state, "STATUS", "Шаг: run_read_tools")
    add_event(
        events,
        state,
        "TOOL_STARTED",
        "Читаю workflow draft через Go Tool Gateway",
        tool_call={
            "tool_key": "workflow.read_draft",
            "risk_level": "read",
            "status": "running",
            "input_json": tool_input,
        },
    )
    observation = run_read_tool_if_enabled(state, "workflow.read_draft", tool_input)
    tool_results: list[dict[str, Any]] = []
    turn_observations: list[dict[str, Any]] = []
    tool_observations = [item for item in list(state.get("tool_observations") or []) if isinstance(item, dict)]
    replan_trace = [item for item in list(state.get("replan_trace") or []) if isinstance(item, dict)]
    if observation is not None:
        tool_results.append(observation)
        normalized = normalize_tool_observation(
            "workflow.read_draft",
            observation,
            tool_input,
            step_index=1,
            run_id=str(state.get("run_id") or ""),
            turn_id=str(state.get("turn_id") or ""),
        )
        turn_observations = append_tool_observation(turn_observations, normalized)
        tool_observations = append_tool_observation(tool_observations, normalized)
        replan_trace.append(replan_trace_entry("workflow.read_draft", turn_observations))
        record_tool_result_event(events, state, normalized, "Workflow draft прочитан")
    return {
        "events": events,
        "active_mode": "project_read",
        "current_step": "run_read_tools",
        "awaiting_tool_observation": len(tool_results) == 0,
        "tool_results": tool_results,
        "tool_observations": tool_observations,
        "replan_trace": replan_trace,
        "final_response": summarize_observations(turn_observations, ""),
    }
