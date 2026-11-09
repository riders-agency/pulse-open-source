package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturlpkg "net/url"
	"strconv"
	"strings"
	"time"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func (r *Runner) executeMax(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	dep, err := r.loadDependency(ctx, node.Config.ActionDependencyID)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": err.Error()}, false, "dependency", err)
	}
	if dep.TypeKey != "max.bot" {
		err := fmt.Errorf("max.send_message requires max.bot dependency")
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": err.Error()}, false, "dependency_type", err)
	}
	if dep.Status != "active" {
		err := fmt.Errorf("max dependency is %s", dep.Status)
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "status": dep.Status}, false, "dependency_status", err)
	}
	token := strings.TrimSpace(dep.Plaintext)
	if token == "" {
		err := errors.New("MAX bot token is empty")
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": err.Error()}, false, "dependency", err)
	}
	recipient, err := resolveMaxRecipient(node.Config, dep.Metadata, runInput, steps, tpl)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": err.Error()}, false, "validation", err)
	}

	textTemplate := node.Config.Text
	if textTemplate == "" && node.Config.MessageTemplateID != "" {
		if body, terr := r.loadMessageTemplate(ctx, node.Config.MessageTemplateID); terr == nil {
			textTemplate = body
		}
	}
	text := strings.TrimSpace(renderTemplate(textTemplate, runInput, steps, tpl))
	if text == "" {
		perr := errors.New("max message text is required")
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": perr.Error()}, false, "validation", perr)
	}
	textParts := splitMaxMessageText(text)
	if len(textParts) > maxMessageTextPartCount {
		perr := fmt.Errorf("max message text requires more than %d parts", maxMessageTextPartCount)
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": perr.Error()}, false, "validation", perr)
	}

	apiBase := strings.TrimRight(strings.TrimSpace(firstNonEmpty(stringMetadata(dep.Metadata, "api_base"), "https://platform-api2.max.ru")), "/")
	bodyObj := map[string]any{"text": text}
	if err := applyMaxBodyOptions(node.Config, dep.Metadata, bodyObj); err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	documentURLs, documentURLsErr := parseMaxDocumentURLs(
		renderTemplate(node.Config.DocumentURL, runInput, steps, tpl),
		renderTemplate(node.Config.DocumentURLsJSON, runInput, steps, tpl),
	)
	if documentURLsErr != nil {
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": documentURLsErr.Error()}, false, "validation", documentURLsErr)
	}
	if documentURLsErr = requireDocumentAttachments("MAX", node.Config.DocumentRequired, len(documentURLs)); documentURLsErr != nil {
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": documentURLsErr.Error()}, false, "validation", documentURLsErr)
	}
	documentAttached := len(documentURLs) > 0
	attachments := make([]map[string]any, 0, len(documentURLs))
	if documentAttached {
		for _, documentURL := range documentURLs {
			document, uploadErr := r.prepareMaxFileAttachment(ctx, apiBase, token, documentURL, maxDocumentMetadata(dep.Metadata, node.Config.DocumentAllowedHosts))
			if uploadErr != nil {
				return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": uploadErr.Error()}, true, "document_upload", uploadErr)
			}
			attachments = append(attachments, document.Attachment)
		}
	}
	messageBodies := buildMaxMessageBodies(bodyObj, attachments)

	uu, perr := neturlpkg.Parse(apiBase + "/messages")
	if perr != nil {
		err := fmt.Errorf("max api_base URL invalid")
		return failedAction(input, "{}", map[string]any{"provider": "max", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	q := uu.Query()
	q.Set(recipient.Param, recipient.Value)
	applyMaxQueryOptions(node.Config, dep.Metadata, q)
	uu.RawQuery = q.Encode()
	reqURL := uu.String()

	requestInput, _ := json.Marshal(map[string]any{
		"recipient_type":    recipient.Param,
		"recipient_id":      recipient.Value,
		"text_chars":        len([]rune(text)),
		"text_bytes":        len([]byte(text)),
		"text_parts":        len(textParts),
		"format":            bodyObj["format"],
		"notify":            bodyObj["notify"],
		"transport":         "max",
		"document_attached": documentAttached,
		"attachments_count": len(documentURLs),
		"tls_trust":         "system+mincifry_ca",
		"success_criteria":  node.Config.SuccessCriteria,
	})
	timeout := durationOrDefault(node.Config.RetryPolicy.Timeout, 30*time.Second)
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	providers := make([]map[string]any, 0, len(messageBodies))
	messageIDs := make([]any, 0, len(messageBodies))
	for messageIndex, messageBody := range messageBodies {
		payload, _ := json.Marshal(messageBody)
		var provider map[string]any
		responseStatus := 0
		for sendAttempt := 0; ; sendAttempt++ {
			req, requestErr := http.NewRequestWithContext(reqCtx, http.MethodPost, reqURL, bytes.NewReader(payload))
			if requestErr != nil {
				err := fmt.Errorf("max request build: %w", requestErr)
				return failedAction(string(requestInput), "{}", map[string]any{"provider": "max", "ok": false}, false, "request", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", maxAuthorizationHeader(token, dep.Metadata))

			resp, requestErr := r.maxHTTPClient().Do(req)
			if requestErr != nil {
				perr := fmt.Errorf("max request failed: %w", requestErr)
				return failedAction(string(requestInput), "{}", map[string]any{"provider": "max", "ok": false, "transport": "max", "tls_trust": "system+mincifry_ca", "messages_sent": len(providers)}, true, "network", perr)
			}
			raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
			_ = resp.Body.Close()
			responseStatus = resp.StatusCode
			provider = sanitizeMaxHTTPResponse(responseStatus, raw)
			provider["transport"] = "max"
			provider["tls_trust"] = "system+mincifry_ca"
			if !maxAttachmentProcessingRetry(documentAttached, sendAttempt, provider) {
				break
			}
			if waitErr := waitMaxAttachmentProcessing(reqCtx, sendAttempt); waitErr != nil {
				return failedAction(string(requestInput), "{}", provider, true, "attachment_processing", waitErr)
			}
		}
		if responseStatus < 200 || responseStatus >= 300 {
			provider["message_index"] = messageIndex
			provider["messages_sent"] = len(providers)
			providerJSON, _ := json.Marshal(provider)
			attachmentProcessing := maxAttachmentNotReady(provider)
			retryable := responseStatus == http.StatusTooManyRequests || responseStatus >= 500 || attachmentProcessing
			errorType := "provider_http"
			if attachmentProcessing {
				errorType = "attachment_processing"
			}
			perr := fmt.Errorf("max returned HTTP %d for message %d of %d", responseStatus, messageIndex+1, len(messageBodies))
			report := diagnosticsForFailure(node.ActionTypeKey, provider, retryable, errorType, perr)
			output := providerdiagnostics.OutputWithReport("{}", report)
			return actionExecution{InputJSON: string(requestInput), OutputJSON: output, ProviderResponseJSON: string(providerJSON), Retryable: retryable, ErrorType: errorType, Err: perr}
		}
		providers = append(providers, provider)
		if messageID := provider["message_id"]; messageID != nil {
			messageIDs = append(messageIDs, messageID)
		}
	}
	provider := cloneMaxMessageBody(providers[0])
	provider["provider"] = "max"
	provider["ok"] = true
	provider["status_code"] = http.StatusOK
	provider["transport"] = "max"
	provider["tls_trust"] = "system+mincifry_ca"
	provider["provider_response_redacted"] = true
	provider["messages_sent"] = len(providers)
	provider["message_ids"] = messageIDs
	provider["responses"] = providers
	if len(messageIDs) > 0 {
		provider["message_id"] = messageIDs[0]
	}
	delivered := map[string]any{
		"delivered":         true,
		"provider":          "max",
		"message_id":        provider["message_id"],
		"recipient_type":    recipient.Param,
		"recipient_id":      recipient.Value,
		"provider_ok":       true,
		"document_attached": documentAttached,
		"attachments_count": len(documentURLs),
		"messages_sent":     len(providers),
		"message_ids":       messageIDs,
	}
	delivered[recipient.Param] = recipient.Value
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(requestInput), provider, delivered, maxExpose(recipient, provider))
}

type maxRecipient struct {
	Param string
	Value string
}

func maxExpose(recipient maxRecipient, provider map[string]any) map[string]any {
	out := map[string]any{"recipient_type": recipient.Param, "recipient_id": recipient.Value, recipient.Param: recipient.Value}
	if v := provider["message_id"]; v != nil {
		out["message_id"] = v
	}
	return out
}

func stripMaxRecipientID(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return s
	}
	if strings.LastIndexByte(s, '/') >= 0 {
		parts := strings.Split(s, "/")
		s = strings.TrimSpace(parts[len(parts)-1])
	}
	if idx := strings.IndexAny(s, "?#"); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}
	return s
}

func resolveMaxRecipient(config actionConfig, metadata map[string]any, runInput []byte, steps map[string]any, system templateSystem) (maxRecipient, error) {
	chatID := stripMaxRecipientID(strings.TrimSpace(renderTemplate(firstNonEmpty(config.ChatID, stringMetadata(metadata, "default_chat_id")), runInput, steps, system)))
	userID := stripMaxRecipientID(strings.TrimSpace(renderTemplate(firstNonEmpty(config.UserID, stringMetadata(metadata, "default_user_id")), runInput, steps, system)))
	if chatID != "" && userID != "" {
		return maxRecipient{}, errors.New("max recipient must be either chat_id or user_id, not both")
	}
	if chatID != "" {
		return parseMaxRecipient("chat_id", chatID)
	}
	if userID != "" {
		return parseMaxRecipient("user_id", userID)
	}
	return maxRecipient{}, errors.New("max recipient required: set chat_id or user_id")
}

func parseMaxRecipient(param, raw string) (maxRecipient, error) {
	value := stripMaxRecipientID(raw)
	if _, err := strconv.ParseInt(value, 10, 64); err != nil || value == "" {
		return maxRecipient{}, fmt.Errorf("max %s must be an integer id", param)
	}
	return maxRecipient{Param: param, Value: value}, nil
}

func applyMaxBodyOptions(config actionConfig, metadata map[string]any, body map[string]any) error {
	if config.Notify != nil {
		body["notify"] = *config.Notify
	} else if b, ok := maxBoolMetadata(metadata, "default_notify"); ok {
		body["notify"] = b
	}
	format := strings.ToLower(strings.TrimSpace(firstNonEmpty(config.Format, stringMetadata(metadata, "default_format"))))
	if format == "" {
		return nil
	}
	if format != "markdown" && format != "html" {
		return errors.New("max format must be markdown or html")
	}
	body["format"] = format
	return nil
}

func applyMaxQueryOptions(config actionConfig, metadata map[string]any, q neturlpkg.Values) {
	if config.DisableLinkPreview != nil {
		q.Set("disable_link_preview", strconv.FormatBool(*config.DisableLinkPreview))
		return
	}
	if b, ok := maxBoolMetadata(metadata, "default_disable_link_preview"); ok {
		q.Set("disable_link_preview", strconv.FormatBool(b))
	}
}
