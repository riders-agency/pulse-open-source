from __future__ import annotations

import pytest

from pulse_agent.graph.intent_routes import INTENT_ROUTES, branch_for_intent, route_for_intent, validate_intent_routes
from pulse_agent.graph.intent_schema import IntentKey


def test_intent_route_registry_covers_every_intent_key() -> None:
    validate_intent_routes()

    assert {intent.value for intent in IntentKey} <= set(INTENT_ROUTES)
    assert branch_for_intent(IntentKey.WORKFLOW_PATCH.value) == "workflow_patch_workflow"
    assert route_for_intent(IntentKey.UNKNOWN.value).active_mode == "error_recovery"


def test_unregistered_runtime_intent_fails_closed() -> None:
    with pytest.raises(ValueError, match="Unhandled intent route"):
        route_for_intent("new.unregistered.intent")
