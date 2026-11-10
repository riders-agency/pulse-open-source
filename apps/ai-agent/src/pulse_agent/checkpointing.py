from __future__ import annotations

import hashlib
import logging
from contextlib import ExitStack
from dataclasses import dataclass
from typing import Any

from psycopg.conninfo import make_conninfo

from pulse_agent.config import Settings

log = logging.getLogger(__name__)

STATE_SCHEMA_VERSION = 1
THREAD_SCOPE_STATUS_ALIASES = {
    "interrupted": "waiting_approval",
    "resuming": "running",
}
THREAD_SCOPE_ALLOWED_STATUSES = {"running", "waiting_approval", "succeeded", "failed", "canceled"}


class CheckpointRuntimeError(RuntimeError):
    pass


@dataclass(frozen=True)
class CheckpointScope:
    company_id: str
    project_id: str
    session_id: str
    public_thread_id: str
    checkpoint_thread_id: str
    checkpoint_namespace: str
    state_schema_version: int = STATE_SCHEMA_VERSION


class CheckpointRuntime:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._stack = ExitStack()
        self._checkpointer: Any | None = None
        self._store: Any | None = None

    @property
    def enabled(self) -> bool:
        return bool(self._settings.checkpoint_database_url)

    def compile(self, graph_builder: Any) -> Any:
        if not self.enabled:
            return graph_builder.compile()
        try:
            from langgraph.checkpoint.postgres import PostgresSaver
            from langgraph.store.postgres import PostgresStore

            conninfo = checkpoint_conninfo(self._settings)
            self._checkpointer = self._stack.enter_context(PostgresSaver.from_conn_string(conninfo))
            self._store = self._stack.enter_context(PostgresStore.from_conn_string(conninfo))
            if self._settings.run_migrations:
                self._checkpointer.setup()
                self._store.setup()
                self._ensure_metadata_tables()
            return graph_builder.compile(checkpointer=self._checkpointer, store=self._store)
        except Exception as exc:  # pragma: no cover - exact driver errors vary by environment.
            self.close()
            raise CheckpointRuntimeError(f"langgraph checkpoint setup failed: {type(exc).__name__}") from exc

    def close(self) -> None:
        self._stack.close()
        self._checkpointer = None
        self._store = None

    def scope(self, company_id: str, project_id: str, session_id: str, thread_id: str) -> CheckpointScope:
        public_thread_id = thread_id or f"pulse:{project_id}:{session_id}"
        namespace = checkpoint_namespace(company_id, project_id, session_id)
        scoped = checkpoint_thread_id(company_id, project_id, session_id, public_thread_id)
        return CheckpointScope(
            company_id=company_id,
            project_id=project_id,
            session_id=session_id,
            public_thread_id=public_thread_id,
            checkpoint_thread_id=scoped,
            checkpoint_namespace=namespace,
        )

    def graph_config(self, scope: CheckpointScope) -> dict[str, dict[str, str]]:
        if not self.enabled:
            return {"configurable": {"thread_id": scope.public_thread_id}}
        return {
            "configurable": {
                "thread_id": scope.checkpoint_thread_id,
            }
        }

    def record_thread_scope(self, scope: CheckpointScope, *, status: str) -> None:
        if not self.enabled or self._checkpointer is None:
            return
        status = normalize_thread_scope_status(status)
        try:
            with self._checkpointer.conn.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO ai_thread_scopes (
                      checkpoint_namespace,
                      checkpoint_thread_id,
                      public_thread_id,
                      company_id,
                      project_id,
                      session_id,
                      state_schema_version,
                      status,
                      first_seen_at,
                      last_seen_at,
                      expires_at
                    )
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, now(), now(), now() + (%s || ' days')::interval)
                    ON CONFLICT (checkpoint_namespace, checkpoint_thread_id) DO UPDATE
                       SET public_thread_id = EXCLUDED.public_thread_id,
                           company_id = EXCLUDED.company_id,
                           project_id = EXCLUDED.project_id,
                           session_id = EXCLUDED.session_id,
                           state_schema_version = EXCLUDED.state_schema_version,
                           status = EXCLUDED.status,
                           last_seen_at = now(),
                           expires_at = EXCLUDED.expires_at
                    """,
                    (
                        scope.checkpoint_namespace,
                        scope.checkpoint_thread_id,
                        scope.public_thread_id,
                        scope.company_id,
                        scope.project_id,
                        scope.session_id,
                        scope.state_schema_version,
                        status,
                        self._settings.checkpoint_retention_days,
                    ),
                )
        except Exception as exc:
            log.exception(
                "failed to record langgraph checkpoint scope",
                extra={"checkpoint_namespace": scope.checkpoint_namespace},
            )
            raise CheckpointRuntimeError(f"langgraph checkpoint metadata failed: {type(exc).__name__}") from exc

    def get_state(self, graph: Any, scope: CheckpointScope) -> Any | None:
        if not self.enabled or graph is None or not hasattr(graph, "get_state"):
            return None
        try:
            return graph.get_state(config=self.graph_config(scope))
        except Exception as exc:
            raise CheckpointRuntimeError(f"langgraph checkpoint state load failed: {type(exc).__name__}") from exc

    def update_state(self, graph: Any, scope: CheckpointScope, values: dict[str, Any]) -> bool:
        if not self.enabled or graph is None or not hasattr(graph, "update_state"):
            return False
        try:
            graph.update_state(config=self.graph_config(scope), values=values)
            return True
        except Exception as exc:
            raise CheckpointRuntimeError(f"langgraph checkpoint state update failed: {type(exc).__name__}") from exc

    def _ensure_metadata_tables(self) -> None:
        if self._checkpointer is None:
            return
        with self._checkpointer.conn.cursor() as cur:
            cur.execute(
                """
                CREATE TABLE IF NOT EXISTS ai_thread_scopes (
                  checkpoint_namespace text NOT NULL,
                  checkpoint_thread_id text NOT NULL,
                  public_thread_id text NOT NULL,
                  company_id text NOT NULL,
                  project_id text NOT NULL,
                  session_id text NOT NULL,
                  state_schema_version integer NOT NULL,
                  status text NOT NULL,
                  first_seen_at timestamptz NOT NULL DEFAULT now(),
                  last_seen_at timestamptz NOT NULL DEFAULT now(),
                  expires_at timestamptz NOT NULL,
                  PRIMARY KEY (checkpoint_namespace, checkpoint_thread_id)
                )
                """
            )
            cur.execute(
                """
                CREATE INDEX IF NOT EXISTS idx_ai_thread_scopes_cleanup
                  ON ai_thread_scopes (expires_at, status)
                """
            )


def checkpoint_conninfo(settings: Settings) -> str:
    return make_conninfo(
        settings.checkpoint_database_url,
        options=f"-csearch_path={settings.checkpoint_schema},public",
    )


def checkpoint_namespace(company_id: str, project_id: str, session_id: str) -> str:
    return "lg:v1:" + ":".join((_short(company_id), _short(project_id), _short(session_id)))


def checkpoint_thread_id(company_id: str, project_id: str, session_id: str, thread_id: str) -> str:
    digest = hashlib.sha256(f"{company_id}\0{project_id}\0{session_id}\0{thread_id}".encode()).hexdigest()[:32]
    return f"lg:v1:{_short(company_id)}:{_short(project_id)}:{_short(session_id)}:{digest}"


def normalize_thread_scope_status(status: str) -> str:
    normalized = (status or "").strip().lower()
    normalized = THREAD_SCOPE_STATUS_ALIASES.get(normalized, normalized)
    if normalized in THREAD_SCOPE_ALLOWED_STATUSES:
        return normalized
    return "failed"


def _short(value: str) -> str:
    cleaned = "".join(ch if ch.isalnum() else "_" for ch in value.strip().lower())
    cleaned = cleaned.strip("_") or "missing"
    if len(cleaned) <= 36:
        return cleaned
    return f"{cleaned[:16]}_{hashlib.sha256(cleaned.encode()).hexdigest()[:12]}"
