package tools

import (
	"context"
	"encoding/json"
	"time"
)

type ToolGateway interface {
	ListTools(ctx context.Context, req ToolCatalogRequest) (ToolCatalog, error)
	ValidateToolCall(ctx context.Context, req ToolCallSpec) (ValidationResult, error)
	ExecuteReadTool(ctx context.Context, req ToolCallSpec) (ToolObservation, error)
	CreateProposal(ctx context.Context, req ToolCallSpec) (ToolProposal, error)
	PatchProposal(ctx context.Context, req PatchProposalRequest) (ToolProposal, error)
	ApproveProposal(ctx context.Context, req ApprovalRequest) (ToolObservation, error)
	RejectProposal(ctx context.Context, req ApprovalRequest) (ToolObservation, error)
}

type ToolCatalogRequest struct {
	RequestID       string          `json:"request_id"`
	CompanyID       string          `json:"company_id"`
	ProjectID       string          `json:"project_id"`
	UserID          string          `json:"user_id"`
	SessionID       string          `json:"session_id"`
	PageContextJSON json.RawMessage `json:"page_context_json"`
}

type ToolCatalog struct {
	Tools         []ToolDescriptor   `json:"tools"`
	Intents       []IntentDescriptor `json:"intents"`
	RuntimePolicy RuntimePolicy      `json:"runtime_policy"`
}

type RuntimePolicy struct {
	ReadToolsMayRunInParallel   bool `json:"read_tools_may_run_in_parallel"`
	WriteToolsMustBeSerial      bool `json:"write_tools_must_be_serial"`
	RequireIdempotencyForWrites bool `json:"require_idempotency_for_writes"`
	MaxToolCallsPerTurn         int  `json:"max_tool_calls_per_turn"`
	MaxWriteProposalsPerTurn    int  `json:"max_write_proposals_per_turn"`
}

type ToolCallSpec struct {
	ToolCallID       string          `json:"tool_call_id"`
	ProjectID        string          `json:"project_id"`
	CompanyID        string          `json:"company_id"`
	UserID           string          `json:"user_id"`
	SessionID        string          `json:"session_id"`
	ToolKey          string          `json:"tool_key"`
	IntentKey        string          `json:"intent_key"`
	RiskLevel        RiskLevel       `json:"risk_level"`
	RequiresApproval bool            `json:"requires_approval"`
	InputJSON        json.RawMessage `json:"input_json"`
	PageContextJSON  json.RawMessage `json:"page_context_json"`
	IdempotencyKey   string          `json:"idempotency_key"`
	ExpectedCardType string          `json:"expected_card_type"`
	ParentPlanID     string          `json:"parent_plan_id"`
	StepID           string          `json:"step_id"`
}

type ValidationResult struct {
	OK                  bool            `json:"ok"`
	NormalizedInputJSON json.RawMessage `json:"normalized_input_json"`
	MissingFields       []MissingField  `json:"missing_fields"`
	ErrorCode           string          `json:"error_code"`
	UserMessage         string          `json:"user_message"`
}

type MissingField struct {
	Field string `json:"field"`
	Label string `json:"label"`
	Query string `json:"query,omitempty"`
}

type ToolProposal struct {
	ToolCallID string          `json:"tool_call_id"`
	ToolKey    string          `json:"tool_key"`
	CardID     string          `json:"card_id"`
	CardType   string          `json:"card_type"`
	Revision   int             `json:"revision"`
	Status     string          `json:"status"`
	InputJSON  json.RawMessage `json:"input_json"`
	CardJSON   json.RawMessage `json:"card_json"`
	CreatedAt  time.Time       `json:"created_at"`
}

type PatchProposalRequest struct {
	ProjectID       string          `json:"project_id"`
	CompanyID       string          `json:"company_id"`
	UserID          string          `json:"user_id"`
	SessionID       string          `json:"session_id"`
	ToolCallID      string          `json:"tool_call_id"`
	CardID          string          `json:"card_id"`
	BaseRevision    int             `json:"base_revision"`
	PatchJSON       json.RawMessage `json:"patch_json"`
	PageContextJSON json.RawMessage `json:"page_context_json"`
}

type ApprovalRequest struct {
	ProjectID      string `json:"project_id"`
	CompanyID      string `json:"company_id"`
	UserID         string `json:"user_id"`
	SessionID      string `json:"session_id"`
	ToolCallID     string `json:"tool_call_id"`
	CardID         string `json:"card_id"`
	CardRevision   int    `json:"card_revision"`
	Decision       string `json:"decision"`
	IdempotencyKey string `json:"idempotency_key"`
}

type ToolObservation struct {
	ToolCallID   string          `json:"tool_call_id"`
	ToolKey      string          `json:"tool_key"`
	Status       string          `json:"status"`
	OutputJSON   json.RawMessage `json:"output_json"`
	ErrorCode    string          `json:"error_code"`
	UserMessage  string          `json:"user_message"`
	EntityLinks  []EntityLink    `json:"entity_links"`
	AuditEventID string          `json:"audit_event_id"`
}

type EntityLink struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Label      string `json:"label"`
	URL        string `json:"url"`
}

func RiskRequiresApproval(risk RiskLevel) bool {
	switch risk {
	case RiskRead:
		return false
	default:
		return true
	}
}
