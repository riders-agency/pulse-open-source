package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	maxMessageDocumentCount = 8
	maxMessageTextBytes     = 4000
	maxMessageTextPartCount = 16
)

func buildMaxMessageBodies(base map[string]any, attachments []map[string]any) []map[string]any {
	text := strings.TrimSpace(fmt.Sprint(base["text"]))
	textParts := splitMaxMessageText(text)
	messageCount := max(len(textParts), len(attachments))
	if messageCount == 0 {
		messageCount = 1
	}
	messages := make([]map[string]any, 0, messageCount)
	for index := 0; index < messageCount; index++ {
		message := cloneMaxMessageBody(base)
		switch {
		case index < len(textParts):
			message["text"] = textParts[index]
		case index < len(attachments):
			message["text"] = fmt.Sprintf("📎 Вложение %d из %d", index+1, len(attachments))
		default:
			message["text"] = "📎 Продолжение сообщения"
		}
		if index < len(attachments) {
			message["attachments"] = []map[string]any{attachments[index]}
		} else {
			delete(message, "attachments")
		}
		messages = append(messages, message)
	}
	return messages
}

func splitMaxMessageText(text string) []string {
	remaining := []rune(strings.TrimSpace(text))
	if len(remaining) == 0 {
		return nil
	}
	parts := make([]string, 0, (len([]byte(string(remaining)))+maxMessageTextBytes-1)/maxMessageTextBytes)
	for len([]byte(string(remaining))) > maxMessageTextBytes {
		cut := maxRunePrefixForBytes(remaining, maxMessageTextBytes)
		minimumNaturalCut := cut * 3 / 4
		for index := cut; index >= minimumNaturalCut; index-- {
			if remaining[index-1] == '\n' {
				cut = index
				break
			}
		}
		if cut == maxRunePrefixForBytes(remaining, maxMessageTextBytes) {
			for index := cut; index >= minimumNaturalCut; index-- {
				if remaining[index-1] == ' ' || remaining[index-1] == '\t' {
					cut = index
					break
				}
			}
		}
		part := strings.TrimSpace(string(remaining[:cut]))
		if part != "" {
			parts = append(parts, part)
		}
		remaining = []rune(strings.TrimSpace(string(remaining[cut:])))
	}
	if tail := strings.TrimSpace(string(remaining)); tail != "" {
		parts = append(parts, tail)
	}
	return parts
}

func maxRunePrefixForBytes(value []rune, byteLimit int) int {
	used := 0
	for index, current := range value {
		width := len(string(current))
		if used+width > byteLimit {
			return index
		}
		used += width
	}
	return len(value)
}

func cloneMaxMessageBody(source map[string]any) map[string]any {
	result := make(map[string]any, len(source)+1)
	for key, value := range source {
		result[key] = value
	}
	return result
}

func parseMaxDocumentURLs(singleURL, urlsJSON string) ([]string, error) {
	singleURL = strings.TrimSpace(singleURL)
	urlsJSON = strings.TrimSpace(urlsJSON)
	urls := []string{}
	if urlsJSON != "" {
		if err := json.Unmarshal([]byte(urlsJSON), &urls); err != nil {
			return nil, errors.New("MAX document_urls_json must be a JSON array")
		}
	} else if singleURL != "" {
		urls = append(urls, singleURL)
	}
	if len(urls) > maxMessageDocumentCount {
		return nil, fmt.Errorf("MAX contains more than %d document attachments", maxMessageDocumentCount)
	}
	seen := make(map[string]struct{}, len(urls))
	for index := range urls {
		urls[index] = strings.TrimSpace(urls[index])
		if urls[index] == "" {
			return nil, errors.New("MAX document attachment URL is empty")
		}
		if _, exists := seen[urls[index]]; exists {
			return nil, errors.New("MAX document attachment URLs must be unique")
		}
		seen[urls[index]] = struct{}{}
	}
	return urls, nil
}

const maxAttachmentProcessingPolls = 5

func maxAttachmentProcessingRetry(documentAttached bool, sendAttempt int, provider map[string]any) bool {
	return documentAttached && sendAttempt < maxAttachmentProcessingPolls-1 && maxAttachmentNotReady(provider)
}

func waitMaxAttachmentProcessing(ctx context.Context, sendAttempt int) error {
	delay := time.Duration(sendAttempt+1) * 750 * time.Millisecond
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func sanitizeMaxHTTPResponse(statusCode int, raw []byte) map[string]any {
	out := map[string]any{"provider": "max", "status_code": statusCode, "provider_response_redacted": true}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		out["ok"] = statusCode >= 200 && statusCode < 300
		out["description"] = strings.TrimSpace(string(raw))
		return out
	}
	out["raw_response_json"] = root
	if okV, exists := root["ok"]; exists {
		out["ok"] = okV
	} else {
		out["ok"] = true
	}
	if msg, exists := root["message"]; exists {
		if mm, okm := msg.(map[string]any); okm {
			if mm["id"] != nil {
				out["message_id"] = mm["id"]
			} else if body, okb := mm["body"].(map[string]any); okb && body["mid"] != nil {
				out["message_id"] = body["mid"]
			}
		}
	}
	for _, key := range []string{"description", "error", "error_description", "code", "reason"} {
		if v, exists := root[key]; exists {
			out[key] = v
		}
	}
	return out
}

func maxAuthorizationHeader(secret string, _ map[string]any) string {
	s := strings.TrimSpace(secret)
	if s == "" {
		return ""
	}
	if ls := strings.ToLower(s); strings.HasPrefix(ls, "bearer ") {
		s = strings.TrimSpace(s[7:])
	}
	return s
}
