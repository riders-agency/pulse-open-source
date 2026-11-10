from __future__ import annotations

import os
import uuid
from typing import Any, TypedDict

import psycopg
import pytest
from psycopg import sql

from pulse_agent.checkpointing import CheckpointRuntime, checkpoint_conninfo
from pulse_agent.config import Settings


def test_checkpoint_scope_is_namespaced_by_project_and_session() -> None:
    runtime = CheckpointRuntime(Settings())

    first = runtime.scope("company-1", "project-1", "session-1", "thread-1")
    same = runtime.scope("company-1", "project-1", "session-1", "thread-1")
    other_project = runtime.scope("company-1", "project-2", "session-1", "thread-1")
    other_session = runtime.scope("company-1", "project-1", "session-2", "thread-1")

    assert first == same
    assert first.checkpoint_thread_id != other_project.checkpoint_thread_id
    assert first.checkpoint_namespace != other_project.checkpoint_namespace
    assert first.checkpoint_thread_id != other_session.checkpoint_thread_id
    assert first.public_thread_id == "thread-1"


def test_disabled_checkpoint_runtime_preserves_graph_compile_and_public_thread_id() -> None:
    runtime = CheckpointRuntime(Settings(checkpoint_database_url=""))
    builder = FakeGraphBuilder()

    graph = runtime.compile(builder)
    scope = runtime.scope("company-1", "project-1", "session-1", "thread-1")

    assert graph == "compiled"
    assert builder.compile_kwargs == {}
    assert runtime.graph_config(scope) == {"configurable": {"thread_id": "thread-1"}}


def test_checkpoint_conninfo_sets_restricted_schema_search_path() -> None:
    settings = Settings(
        checkpoint_database_url="postgres://pulse_ai@127.0.0.1:5432/pulse_ai?sslmode=disable",
        checkpoint_schema="ai_agent_checkpoint",
    )

    conninfo = checkpoint_conninfo(settings)

    assert "user=pulse_ai" in conninfo
    assert "options=-csearch_path=ai_agent_checkpoint,public" in conninfo


def test_checkpoint_runtime_persists_state_postgres() -> None:
    database_url = os.getenv("PULSE_TEST_CHECKPOINT_DATABASE_URL", "")
    if not database_url:
        pytest.skip("PULSE_TEST_CHECKPOINT_DATABASE_URL is not set")
    schema = f"ai_agent_checkpoint_{uuid.uuid4().hex[:8]}"
    with psycopg.connect(database_url, autocommit=True) as conn:
        conn.execute(sql.SQL("CREATE SCHEMA {}").format(sql.Identifier(schema)))
    settings = Settings(
        checkpoint_database_url=database_url,
        checkpoint_schema=schema,
        checkpoint_retention_days="1",
        run_migrations=True,
    )
    scope = CheckpointRuntime(settings).scope("company-1", "project-1", "session-1", "thread-1")
    try:
        first_runtime = CheckpointRuntime(settings)
        first_graph = first_runtime.compile(_tiny_graph_builder())
        first_graph.invoke({"body": "persist me"}, config=first_runtime.graph_config(scope))
        first_runtime.close()

        second_runtime = CheckpointRuntime(settings)
        second_graph = second_runtime.compile(_tiny_graph_builder())
        snapshot = second_runtime.get_state(second_graph, scope)
        second_runtime.close()

        assert snapshot is not None
        assert snapshot.values["current_step"] == "persisted"
        assert snapshot.values["body"] == "persist me"
    finally:
        with psycopg.connect(database_url, autocommit=True) as conn:
            conn.execute(sql.SQL("DROP SCHEMA IF EXISTS {} CASCADE").format(sql.Identifier(schema)))


class FakeGraphBuilder:
    def __init__(self) -> None:
        self.compile_kwargs: dict[str, Any] = {}

    def compile(self, **kwargs: Any) -> str:
        self.compile_kwargs = kwargs
        return "compiled"


class TinyState(TypedDict, total=False):
    body: str
    current_step: str


def _tiny_graph_builder() -> Any:
    from langgraph.graph import END, START, StateGraph

    builder = StateGraph(TinyState)
    builder.add_node("persist", _persist_node)
    builder.add_edge(START, "persist")
    builder.add_edge("persist", END)
    return builder


def _persist_node(state: TinyState) -> TinyState:
    return {"body": state.get("body", ""), "current_step": "persisted"}
