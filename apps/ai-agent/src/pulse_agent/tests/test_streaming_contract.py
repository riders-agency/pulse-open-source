from __future__ import annotations

from pulse_agent.schemas.events import AgentEvent, AgentEventType


def test_streaming_event_contract_orders_seq() -> None:
    events = [
        AgentEvent(seq=1, run_id="r", thread_id="th", session_id="s", event_type=AgentEventType.RUN_STARTED),
        AgentEvent(seq=2, run_id="r", thread_id="th", session_id="s", event_type=AgentEventType.CARD_CREATED),
        AgentEvent(seq=3, run_id="r", thread_id="th", session_id="s", event_type=AgentEventType.DONE),
    ]
    assert [event.seq for event in events] == sorted(event.seq for event in events)
    assert events[-1].event_type == AgentEventType.DONE
