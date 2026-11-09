from pulse_agent.execution_trace import execution_summary, record_model_call, traced_node
from pulse_agent.export_graph import graph_contract


def test_trace_contains_actual_nodes_and_usage_without_context_leak():
    def model_step(state):
        record_model_call("intent_router", {"provider_key": "fixture", "model_id": "fixture-model",
                                           "input_tokens": 13, "output_tokens": 4, "api_key": "private"})
        return {"intent": {"intent_key": "task.create"}}

    result = traced_node("classify_intent", model_step)({"password": "private"})
    summary = execution_summary({**result, "__interrupt__": [object()]})
    assert summary["input_tokens"] == 13
    assert summary["output_tokens"] == 4
    assert summary["nodes"][0]["node"] == "classify_intent"
    assert summary["nodes"][-1]["status"] == "waiting"
    assert "private" not in str(summary)
    empty = traced_node("read_context", lambda _: {})({})
    assert execution_summary(empty)["model_call_count"] == 0


def test_topology_is_exported_from_the_real_graph():
    contract = graph_contract()
    assert "await_approval" in contract["nodes"]
    assert {"source": "task_create_workflow", "target": "await_approval", "conditional": False} in contract["edges"]
    assert not any(
        edge["source"] == "task_create_workflow" and edge["target"] == "approval_decision" for edge in contract["edges"]
    )


def test_execution_tools_are_scoped_to_current_turn() -> None:
    trace = execution_summary({
        "events": [
            {"tool_call": {"tool_key": "old.tool"}},
            {"tool_call": {"tool_key": "task.search", "status": "succeeded"}},
        ],
        "execution_event_start": 1,
    })
    assert [tool["tool"] for tool in trace["tools"]] == ["task.search"]
