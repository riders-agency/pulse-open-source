package execution

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTraceUsesProviderCallsAndDropsUnrelatedData(t *testing.T) {
	trace := Parse(map[string]any{"schema": "pulse-execution-v1", "status": "waiting_approval", "input_tokens": 999,
		"prompt": "private context", "api_key": "private credential", "model_calls": []map[string]any{{"input_tokens": 11, "output_tokens": 7, "provider": "test", "model": "fixture"}}})
	if trace == nil || trace.InputTokens != 11 || trace.OutputTokens != 7 || trace.ModelCallCount != 1 {
		t.Fatalf("trace: %+v", trace)
	}
	raw, _ := json.Marshal(trace)
	if strings.Contains(string(raw), "private") {
		t.Fatal("unrelated sensitive data retained")
	}
	if Parse(map[string]any{"schema": "pulse-execution-v1", "status": "completed", "model_calls": []map[string]any{{"input_tokens": -1}}}) != nil {
		t.Fatal("invalid accounting accepted")
	}
}
