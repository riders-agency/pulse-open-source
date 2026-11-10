package http

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
	airepo "riders.agency/pulse/api/internal/modules/ai/repository"
	aisvc "riders.agency/pulse/api/internal/modules/ai/service"
)

type langGraphTurnContext struct {
	CompanyID string
	ThreadID  string
	History   []*aiagentv1.ChatMessage
}

func loadLangGraphTurnContext(
	ctx context.Context,
	ai *aisvc.Commands,
	projectID uuid.UUID,
	sessionID uuid.UUID,
	userID uuid.UUID,
) (langGraphTurnContext, error) {
	out := langGraphTurnContext{ThreadID: "pulse:" + projectID.String() + ":" + sessionID.String()}
	if ai == nil {
		return out, nil
	}
	if _, err := ai.EmployeePolicy(ctx, projectID); err != nil {
		return langGraphTurnContext{}, err
	}
	session, err := ai.GetSession(ctx, projectID, sessionID, userID)
	if err != nil {
		return langGraphTurnContext{}, err
	}
	out.CompanyID = strings.TrimSpace(stringFromAny(session["company_id"]))
	out.ThreadID = firstNonEmpty(strings.TrimSpace(stringFromAny(session["langgraph_thread_id"])), out.ThreadID)
	historyMessages, err := ai.LangGraphTurnHistory(ctx, projectID, sessionID)
	if err != nil {
		return langGraphTurnContext{}, err
	}
	out.History = langGraphProtoHistory(historyMessages)
	return out, nil
}

func langGraphProtoHistory(messages []airepo.LangGraphHistoryMessage) []*aiagentv1.ChatMessage {
	out := make([]*aiagentv1.ChatMessage, 0, len(messages))
	for _, message := range messages {
		payload, err := structpb.NewStruct(message.Payload)
		if err != nil {
			payload = &structpb.Struct{Fields: map[string]*structpb.Value{}}
		}
		item := &aiagentv1.ChatMessage{
			MessageId:   message.MessageID,
			Role:        message.Role,
			Body:        message.Body,
			PayloadJson: payload,
		}
		if !message.CreatedAt.IsZero() {
			item.CreatedAt = timestamppb.New(message.CreatedAt)
		}
		out = append(out, item)
	}
	return out
}
