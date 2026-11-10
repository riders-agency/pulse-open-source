package modelgateway

import (
	"context"
	"encoding/json"
	"hash/fnv"
)

type FakeProvider struct{}

func (FakeProvider) Chat(ctx context.Context, call ChatCall) (ChatResult, error) {
	if err := ctx.Err(); err != nil {
		return ChatResult{}, err
	}
	payload := map[string]any{
		"ok":       true,
		"provider": "fake",
		"purpose":  call.Purpose,
		"echo":     lastUserContent(call.Messages),
	}
	raw, _ := json.Marshal(payload)
	return ChatResult{
		Content:      string(raw),
		ResponseJSON: payload,
		InputTokens:  estimateTokens(call.Messages),
		OutputTokens: int64(len(raw) / 4),
		Metadata:     map[string]any{"provider": "fake", "model": call.ModelID},
	}, nil
}

func (FakeProvider) Embed(ctx context.Context, call EmbedCall) (EmbedResult, error) {
	if err := ctx.Err(); err != nil {
		return EmbedResult{}, err
	}
	embeddings := make([][]float64, 0, len(call.Input))
	for _, item := range call.Input {
		embeddings = append(embeddings, fakeEmbedding(item))
	}
	return EmbedResult{
		Embeddings:  embeddings,
		InputTokens: int64(len(call.Input)),
		Metadata:    map[string]any{"provider": "fake", "model": call.ModelID},
	}, nil
}

func lastUserContent(messages []Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	if len(messages) == 0 {
		return ""
	}
	return messages[len(messages)-1].Content
}

func estimateTokens(messages []Message) int64 {
	var total int64
	for _, message := range messages {
		total += int64(len(message.Content)/4 + 1)
	}
	return total
}

func fakeEmbedding(value string) []float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(value))
	sum := h.Sum32()
	return []float64{
		float64(len(value)),
		float64(sum%1000) / 1000,
		float64((sum/1000)%1000) / 1000,
	}
}
