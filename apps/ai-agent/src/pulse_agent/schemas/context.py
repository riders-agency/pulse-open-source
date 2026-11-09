from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field


class PageContext(BaseModel):
    page_kind: str = Field(default="", alias="pageKind")
    route: str = ""
    pathname: str = ""
    company_id: str = Field(default="", alias="companyId")
    project_id: str = Field(default="", alias="projectId")
    workflow_id: str = Field(default="", alias="workflowId")
    workflow_run_id: str = Field(default="", alias="workflowRunId")
    task_id: str = Field(default="", alias="taskId")
    client_id: str = Field(default="", alias="clientId")
    selected_node_id: str = Field(default="", alias="selectedNodeId")
    selected_action_key: str = Field(default="", alias="selectedActionKey")
    visible_panel: str = Field(default="", alias="visiblePanel")
    selection_json: dict[str, Any] = Field(default_factory=dict)
    ui_state_json: dict[str, Any] = Field(default_factory=dict)

    model_config = {"populate_by_name": True}
