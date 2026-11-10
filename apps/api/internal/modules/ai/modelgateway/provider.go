package modelgateway

import (
	"context"
	"time"
)

type Message struct {
	Role     string
	Content  string
	Metadata map[string]any
}

type ChatCall struct {
	RequestID      string
	CompanyID      string
	ProjectID      string
	UserID         string
	SessionID      string
	ThreadID       string
	Purpose        string
	ModelID        string
	Messages       []Message
	ResponseFormat map[string]any
	ToolSchemas    map[string]any
	MaxTokens      int
	Temperature    float64
	Timeout        time.Duration
}

type ChatResult struct {
	Content      string
	ResponseJSON map[string]any
	InputTokens  int64
	OutputTokens int64
	Metadata     map[string]any
}

type EmbedCall struct {
	RequestID string
	CompanyID string
	ProjectID string
	UserID    string
	ModelID   string
	Input     []string
	Timeout   time.Duration
}

type EmbedResult struct {
	Embeddings  [][]float64
	InputTokens int64
	Metadata    map[string]any
}

type Provider interface {
	Chat(context.Context, ChatCall) (ChatResult, error)
	Embed(context.Context, EmbedCall) (EmbedResult, error)
}
