package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, value any) {
	payload, err := json.Marshal(value)
	if err != nil {
		payload = []byte(`{"type":"ERROR","message":"stream serialization failed"}`)
	}
	_, _ = fmt.Fprintf(w, "event: %s\n", event)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
	if flusher != nil {
		flusher.Flush()
	}
}

func streamAssistantText(w http.ResponseWriter, flusher http.Flusher, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	runes := []rune(text)
	chunkSize := 8
	if utf8.RuneCountInString(text) < 48 {
		chunkSize = 4
	}
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		writeSSE(w, flusher, "delta", map[string]any{
			"type": "MESSAGE_DELTA",
			"text": string(runes[start:end]),
		})
		time.Sleep(18 * time.Millisecond)
	}
}

func lastToolBody(session map[string]any) string {
	messages, ok := session["messages"].([]map[string]any)
	if ok {
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i]["role"] == "tool" || messages[i]["role"] == "approval" {
				return stringFromAny(messages[i]["body"])
			}
		}
	}
	raw, err := json.Marshal(session["messages"])
	if err == nil {
		var normalized []map[string]any
		if json.Unmarshal(raw, &normalized) == nil {
			for i := len(normalized) - 1; i >= 0; i-- {
				if stringFromAny(normalized[i]["role"]) == "tool" || stringFromAny(normalized[i]["role"]) == "approval" {
					return stringFromAny(normalized[i]["body"])
				}
			}
		}
	}
	return ""
}

func aiStreamErrorMessage(err error) string {
	text := strings.TrimSpace(errorString(err))
	if text == "" {
		return "Не удалось выполнить ход AI-сотрудника."
	}
	switch {
	case strings.Contains(text, "langgraph_checkpoint_error") || strings.Contains(text, "langgraph checkpoint"):
		return "AI runtime не смог сохранить состояние диалога: checkpoint недоступен."
	case strings.Contains(text, "ai_v3_proposal_gate_not_ready"):
		return "AI-сотрудник сейчас не принимает write-команды: proposal gate выключен."
	case strings.Contains(text, "ai_v3_write_gate_not_ready"):
		return "AI-сотрудник сейчас не применяет изменения: write gate выключен."
	case strings.Contains(text, "python LangGraph unavailable"):
		return "Python LangGraph runtime недоступен."
	default:
		return "Не удалось выполнить ход AI-сотрудника: " + text
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func stringFromAny(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
