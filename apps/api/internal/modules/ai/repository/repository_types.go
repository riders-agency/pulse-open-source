package repository

import (
	"encoding/json"
	"time"
)

type EmployeePatch struct {
	DisplayName        *string        `json:"display_name"`
	Status             *string        `json:"status"`
	ActionDependencyID *string        `json:"action_dependency_id"`
	ModelID            *string        `json:"model_id"`
	Profile            map[string]any `json:"profile_json"`
	ToolPolicy         map[string]any `json:"tool_policy_json"`
	MemoryPolicy       map[string]any `json:"memory_policy_json"`
	UISettings         map[string]any `json:"ui_settings_json"`
}

type SessionInput struct {
	ScreenScope      string          `json:"screen_scope"`
	ScreenEntityType string          `json:"screen_entity_type"`
	ScreenEntityID   string          `json:"screen_entity_id"`
	WorkflowID       string          `json:"workflow_id"`
	WorkflowRunID    string          `json:"workflow_run_id"`
	TaskID           string          `json:"task_id"`
	ClientID         string          `json:"client_id"`
	ActionDependency string          `json:"action_dependency_id"`
	ModelID          string          `json:"model_id"`
	PageContextJSON  json.RawMessage `json:"page_context_json"`
}

type MessageInput struct {
	// ReceivedAt is assigned by the server before tools run, never from client JSON.
	ReceivedAt            time.Time       `json:"-"`
	Body                  string          `json:"body"`
	PageContextJSON       json.RawMessage `json:"page_context_json"`
	ClientContextJSON     json.RawMessage `json:"client_context_json"`
	UIStateJSON           json.RawMessage `json:"ui_state_json"`
	ActiveToolCallID      string          `json:"active_tool_call_id"`
	ActivePatchProposalID string          `json:"active_patch_proposal_id"`
	ReplyToMessageID      string          `json:"reply_to_message_id"`
	ClientCommand         string          `json:"client_command"`
	LocalDate             string          `json:"local_date"`
	Timezone              string          `json:"timezone"`
}
