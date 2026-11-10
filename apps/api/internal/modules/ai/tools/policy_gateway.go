package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrToolExecutionNotImplemented = errors.New("tool execution is not implemented in policy gateway")

type PolicyGateway struct {
	catalog ToolCatalog
}

func NewPolicyGateway(capabilities []ModuleCapability) *PolicyGateway {
	tools := make([]ToolDescriptor, 0)
	intents := make([]IntentDescriptor, 0)
	seenTools := map[string]string{}
	seenIntents := map[string]string{}
	for _, capability := range capabilities {
		for _, descriptor := range capability.Tools {
			if previous := seenTools[descriptor.ToolKey]; previous != "" {
				panic(fmt.Sprintf("duplicate AI tool key %q in modules %s and %s", descriptor.ToolKey, previous, capability.ModuleKey))
			}
			seenTools[descriptor.ToolKey] = capability.ModuleKey
			tools = append(tools, descriptor)
		}
		for _, descriptor := range capability.Intents {
			if previous := seenIntents[descriptor.Intent]; previous != "" {
				panic(fmt.Sprintf("duplicate AI intent key %q in modules %s and %s", descriptor.Intent, previous, capability.ModuleKey))
			}
			seenIntents[descriptor.Intent] = capability.ModuleKey
			intents = append(intents, descriptor)
		}
	}
	return &PolicyGateway{
		catalog: ToolCatalog{
			Tools:   tools,
			Intents: intents,
			RuntimePolicy: RuntimePolicy{
				ReadToolsMayRunInParallel:   true,
				WriteToolsMustBeSerial:      true,
				RequireIdempotencyForWrites: true,
				MaxToolCallsPerTurn:         8,
				MaxWriteProposalsPerTurn:    2,
			},
		},
	}
}

func (g *PolicyGateway) ListTools(context.Context, ToolCatalogRequest) (ToolCatalog, error) {
	return g.catalog, nil
}

func (g *PolicyGateway) ValidateToolCall(_ context.Context, req ToolCallSpec) (ValidationResult, error) {
	descriptor, ok := g.tool(req.ToolKey)
	if !ok {
		return ValidationResult{OK: false, ErrorCode: "unknown_tool", UserMessage: "Tool is not registered."}, nil
	}
	// The caller may request stricter handling, but cannot downgrade the registered tool policy.
	requiresApproval := descriptor.RequiresApproval || RiskRequiresApproval(descriptor.RiskLevel) || RiskRequiresApproval(req.RiskLevel)
	if requiresApproval && !req.RequiresApproval {
		return ValidationResult{OK: false, ErrorCode: "approval_required", UserMessage: "Write tool requires approval."}, nil
	}
	if requiresApproval && strings.TrimSpace(req.IdempotencyKey) == "" {
		return ValidationResult{OK: false, ErrorCode: "idempotency_required", UserMessage: "Write tool requires idempotency key."}, nil
	}
	normalized := req.InputJSON
	if len(normalized) == 0 {
		normalized = json.RawMessage(`{}`)
	}
	return ValidationResult{OK: true, NormalizedInputJSON: normalized}, nil
}

func (g *PolicyGateway) ExecuteReadTool(context.Context, ToolCallSpec) (ToolObservation, error) {
	return ToolObservation{}, ErrToolExecutionNotImplemented
}

func (g *PolicyGateway) CreateProposal(_ context.Context, req ToolCallSpec) (ToolProposal, error) {
	validation, err := g.ValidateToolCall(context.Background(), req)
	if err != nil {
		return ToolProposal{}, err
	}
	if !validation.OK {
		return ToolProposal{}, errors.New(validation.ErrorCode)
	}
	return ToolProposal{
		ToolCallID: req.ToolCallID,
		ToolKey:    req.ToolKey,
		CardID:     firstNonEmpty(req.ExpectedCardType, req.ToolKey) + ":" + firstNonEmpty(req.ToolCallID, req.StepID),
		CardType:   req.ExpectedCardType,
		Revision:   1,
		Status:     "needs_approval",
		InputJSON:  validation.NormalizedInputJSON,
		CardJSON:   json.RawMessage(`{}`),
		CreatedAt:  time.Now().UTC(),
	}, nil
}

func (g *PolicyGateway) PatchProposal(context.Context, PatchProposalRequest) (ToolProposal, error) {
	return ToolProposal{}, ErrToolExecutionNotImplemented
}

func (g *PolicyGateway) ApproveProposal(context.Context, ApprovalRequest) (ToolObservation, error) {
	return ToolObservation{}, ErrToolExecutionNotImplemented
}

func (g *PolicyGateway) RejectProposal(context.Context, ApprovalRequest) (ToolObservation, error) {
	return ToolObservation{}, ErrToolExecutionNotImplemented
}

func (g *PolicyGateway) tool(key string) (ToolDescriptor, bool) {
	for _, descriptor := range g.catalog.Tools {
		if descriptor.ToolKey == key {
			return descriptor, true
		}
	}
	return ToolDescriptor{}, false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
