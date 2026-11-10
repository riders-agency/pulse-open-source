package execution

import "encoding/json"

type Node struct {
	Name      string `json:"node"`
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
}

type ModelCall struct {
	Node         string `json:"node"`
	Purpose      string `json:"purpose"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	LatencyMS    int64  `json:"latency_ms"`
}

type Tool struct {
	Name   string `json:"tool"`
	Status string `json:"status"`
	Event  string `json:"event"`
}

// Trace contains operational evidence only; prompts, credentials and tool arguments are excluded.
type Trace struct {
	Schema         string      `json:"schema"`
	Status         string      `json:"status"`
	Nodes          []Node      `json:"nodes"`
	ModelCalls     []ModelCall `json:"model_calls"`
	Tools          []Tool      `json:"tools"`
	InputTokens    int64       `json:"input_tokens"`
	OutputTokens   int64       `json:"output_tokens"`
	ModelCallCount int         `json:"model_call_count"`
	UsageSource    string      `json:"usage_source"`
}

func Parse(value any) *Trace {
	raw, err := json.Marshal(value)
	if err != nil || len(raw) > 64*1024 {
		return nil
	}
	var out Trace
	if json.Unmarshal(raw, &out) != nil || out.Schema != "pulse-execution-v1" {
		return nil
	}
	if len(out.Nodes) > 64 || len(out.ModelCalls) > 32 || len(out.Tools) > 64 {
		return nil
	}
	if out.Status != "completed" && out.Status != "waiting_approval" {
		return nil
	}
	out.InputTokens, out.OutputTokens = 0, 0
	for _, call := range out.ModelCalls {
		if call.InputTokens < 0 || call.OutputTokens < 0 || call.InputTokens > 100_000_000 || call.OutputTokens > 100_000_000 {
			return nil
		}
		out.InputTokens += call.InputTokens
		out.OutputTokens += call.OutputTokens
	}
	out.ModelCallCount = len(out.ModelCalls)
	out.UsageSource = "provider_response"
	return &out
}

func (t *Trace) Usage() map[string]any {
	return map[string]any{"input_tokens": t.InputTokens, "output_tokens": t.OutputTokens, "model_call_count": t.ModelCallCount, "source": t.UsageSource}
}
