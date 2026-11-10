package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
)

func (r *Runner) executeTelegram(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	dep, err := r.loadDependency(ctx, node.Config.ActionDependencyID)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "dependency", err)
	}
	if dep.TypeKey != "telegram.bot" && dep.TypeKey != "telegram.relay" {
		err := fmt.Errorf("telegram.send_message requires telegram.bot or telegram.relay dependency")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "dependency_type", err)
	}
	if dep.Status != "active" {
		err := fmt.Errorf("telegram dependency is %s", dep.Status)
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "status": dep.Status}, false, "dependency_status", err)
	}

	chatID := strings.TrimSpace(renderTemplate(firstNonEmpty(node.Config.ChatID, stringMetadata(dep.Metadata, "default_chat_id")), runInput, steps, tpl))
	if chatID == "" {
		err := errors.New("telegram chat_id is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	textTemplate := node.Config.Text
	if textTemplate == "" && node.Config.MessageTemplateID != "" {
		if body, err := r.loadMessageTemplate(ctx, node.Config.MessageTemplateID); err == nil {
			textTemplate = body
		}
	}
	text := strings.TrimSpace(renderTemplate(textTemplate, runInput, steps, tpl))
	if text == "" {
		err := errors.New("telegram text is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	if shouldSkipTelegramBotReplyToChannel(runInput) {
		outputJSON := skippedOutputWithReason(node, "bot_reply_not_sent_to_channel")
		providerJSON, _ := json.Marshal(map[string]any{
			"provider": "telegram",
			"ok":       true,
			"skipped":  true,
			"reason":   "bot_reply_not_sent_to_channel",
		})
		skipInput, _ := json.Marshal(map[string]any{"chat_id": chatID, "text": text, "skipped": true})
		return actionExecution{InputJSON: string(skipInput), OutputJSON: outputJSON, ProviderResponseJSON: string(providerJSON)}
	}

	parseMode := firstNonEmpty(node.Config.ParseMode, stringMetadata(dep.Metadata, "parse_mode"))
	if strings.EqualFold(parseMode, "plain") {
		parseMode = ""
	}
	replyMarkup, err := renderTelegramReplyMarkup(node.Config.ReplyMarkupJSON, runInput, steps, tpl)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	replyToMessageIDRaw := strings.TrimSpace(renderTemplate(node.Config.ReplyToMessageID, runInput, steps, tpl))
	var replyToMessageID int64
	if replyToMessageIDRaw != "" {
		replyToMessageID, err = strconv.ParseInt(replyToMessageIDRaw, 10, 64)
		if err != nil || replyToMessageID <= 0 {
			err := errors.New("telegram reply_to_message_id must be a positive integer")
			return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
		}
	}
	parts := splitTelegramHTML(text, 3900)
	if len(parts) == 0 {
		err := errors.New("telegram text is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	requestInput, _ := json.Marshal(map[string]any{
		"chat_id":              chatID,
		"parts":                len(parts),
		"parse_mode":           parseMode,
		"disable_notification": node.Config.DisableNotification,
		"has_reply_markup":     replyMarkup != nil,
		"reply_to_message_id":  replyToMessageID,
		"success_criteria":     node.Config.SuccessCriteria,
	})
	timeout := durationOrDefault(node.Config.RetryPolicy.Timeout, 30*time.Second)
	tgURL := "https://api.telegram.org/bot" + dep.Plaintext + "/sendMessage"
	var lastProvider map[string]any
	for i, part := range parts {
		sendText := part
		if len(parts) > 1 {
			sendText = fmt.Sprintf("<i>Часть %d/%d</i>\n\n%s", i+1, len(parts), part)
		}
		body := map[string]any{
			"chat_id":              chatID,
			"text":                 sendText,
			"disable_notification": node.Config.DisableNotification,
		}
		if parseMode != "" {
			body["parse_mode"] = parseMode
		}
		if replyMarkup != nil && i == len(parts)-1 {
			body["reply_markup"] = replyMarkup
		}
		if replyToMessageID > 0 {
			body["reply_parameters"] = map[string]any{
				"message_id": replyToMessageID,
			}
		}
		payload, _ := json.Marshal(body)
		reqCtx, cancel := context.WithTimeout(ctx, timeout)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, tgURL, bytes.NewReader(payload))
		if err != nil {
			cancel()
			return failedAction(string(requestInput), "{}", map[string]any{"provider": "telegram", "ok": false, "error": "request build failed"}, false, "request", errors.New("telegram request build failed"))
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := r.httpClient().Do(req)
		if err != nil {
			cancel()
			return failedAction(string(requestInput), "{}", map[string]any{"provider": "telegram", "ok": false, "error": "telegram request failed"}, true, "network", errors.New("telegram request failed"))
		}
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		cancel()
		lastProvider = sanitizeTelegramResponse(resp.StatusCode, raw)
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			providerJSON, _ := json.Marshal(lastProvider)
			retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
			err := fmt.Errorf("telegram returned HTTP %d", resp.StatusCode)
			report := diagnosticsForFailure(node.ActionTypeKey, lastProvider, retryable, "provider_http", err)
			output := providerdiagnostics.OutputWithReport("{}", report)
			return actionExecution{InputJSON: string(requestInput), OutputJSON: output, ProviderResponseJSON: string(providerJSON), Retryable: retryable, ErrorType: "provider_http", Err: err}
		}
	}
	delivered := map[string]any{
		"delivered":   true,
		"provider":    "telegram",
		"message_id":  lastProvider["message_id"],
		"chat_id":     chatID,
		"provider_ok": true,
		"parts_sent":  len(parts),
	}
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(requestInput), lastProvider, delivered, telegramExpose(chatID, lastProvider))
}

func (r *Runner) executeTelegramEditMessageText(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	dep, err := r.loadTelegramDependency(ctx, node.Config.ActionDependencyID, "telegram.edit_message_text")
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, telegramDependencyErrorType(err), err)
	}
	chatID := strings.TrimSpace(renderTemplate(firstNonEmpty(node.Config.ChatID, stringMetadata(dep.Metadata, "default_chat_id")), runInput, steps, tpl))
	if chatID == "" {
		err := errors.New("telegram chat_id is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	messageIDRaw := strings.TrimSpace(renderTemplate(node.Config.MessageID, runInput, steps, tpl))
	messageID, err := strconv.ParseInt(messageIDRaw, 10, 64)
	if err != nil || messageID <= 0 {
		err := errors.New("telegram message_id is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	text := strings.TrimSpace(renderTemplate(node.Config.Text, runInput, steps, tpl))
	if text == "" {
		err := errors.New("telegram text is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	parseMode := firstNonEmpty(node.Config.ParseMode, stringMetadata(dep.Metadata, "parse_mode"))
	if strings.EqualFold(parseMode, "plain") {
		parseMode = ""
	}
	replyMarkup, err := renderTelegramReplyMarkup(node.Config.ReplyMarkupJSON, runInput, steps, tpl)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	body := map[string]any{"chat_id": chatID, "message_id": messageID, "text": text}
	if parseMode != "" {
		body["parse_mode"] = parseMode
	}
	if replyMarkup != nil {
		body["reply_markup"] = replyMarkup
	}
	requestInput, _ := json.Marshal(map[string]any{
		"chat_id":          chatID,
		"message_id":       messageID,
		"parse_mode":       parseMode,
		"has_reply_markup": replyMarkup != nil,
		"success_criteria": node.Config.SuccessCriteria,
	})
	provider, ex := r.callTelegramJSON(ctx, dep, "telegram.edit_message_text", "editMessageText", body, durationOrDefault(node.Config.RetryPolicy.Timeout, 30*time.Second))
	if ex.Err != nil {
		ex.InputJSON = string(requestInput)
		return ex
	}
	delivered := map[string]any{
		"edited":       true,
		"provider":     "telegram",
		"message_id":   messageID,
		"chat_id":      chatID,
		"provider_ok":  true,
		"reply_markup": replyMarkup != nil,
	}
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(requestInput), provider, delivered, map[string]any{"chat_id": chatID, "message_id": messageID})
}

func (r *Runner) executeTelegramAnswerCallbackQuery(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	dep, err := r.loadTelegramDependency(ctx, node.Config.ActionDependencyID, "telegram.answer_callback_query")
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, telegramDependencyErrorType(err), err)
	}
	callbackID := strings.TrimSpace(renderTemplate(node.Config.CallbackQueryID, runInput, steps, tpl))
	if callbackID == "" {
		err := errors.New("telegram callback_query_id is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	text := strings.TrimSpace(renderTemplate(node.Config.Text, runInput, steps, tpl))
	body := map[string]any{"callback_query_id": callbackID, "show_alert": node.Config.ShowAlert}
	if text != "" {
		body["text"] = text
	}
	requestInput, _ := json.Marshal(map[string]any{
		"callback_query_id": "redacted",
		"has_text":          text != "",
		"show_alert":        node.Config.ShowAlert,
		"success_criteria":  node.Config.SuccessCriteria,
	})
	provider, ex := r.callTelegramJSON(ctx, dep, "telegram.answer_callback_query", "answerCallbackQuery", body, durationOrDefault(node.Config.RetryPolicy.Timeout, 15*time.Second))
	if ex.Err != nil {
		ex.InputJSON = string(requestInput)
		return ex
	}
	delivered := map[string]any{"answered": true, "provider": "telegram", "provider_ok": true}
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(requestInput), provider, delivered, nil)
}

func (r *Runner) executeTelegramPinMessage(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	dep, err := r.loadTelegramDependency(ctx, node.Config.ActionDependencyID, "telegram.pin_message")
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, telegramDependencyErrorType(err), err)
	}
	chatID := strings.TrimSpace(renderTemplate(firstNonEmpty(node.Config.ChatID, stringMetadata(dep.Metadata, "default_chat_id")), runInput, steps, tpl))
	if chatID == "" {
		err := errors.New("telegram chat_id is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	messageIDRaw := strings.TrimSpace(renderTemplate(node.Config.MessageID, runInput, steps, tpl))
	messageID, err := strconv.ParseInt(messageIDRaw, 10, 64)
	if err != nil || messageID <= 0 {
		err := errors.New("telegram message_id is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	body := map[string]any{
		"chat_id":              chatID,
		"message_id":           messageID,
		"disable_notification": node.Config.DisableNotification,
	}
	requestInput, _ := json.Marshal(map[string]any{
		"chat_id":              chatID,
		"message_id":           messageID,
		"disable_notification": node.Config.DisableNotification,
		"success_criteria":     node.Config.SuccessCriteria,
	})
	provider, ex := r.callTelegramJSON(ctx, dep, "telegram.pin_message", "pinChatMessage", body, durationOrDefault(node.Config.RetryPolicy.Timeout, 15*time.Second))
	if ex.Err != nil {
		ex.InputJSON = string(requestInput)
		return ex
	}
	delivered := map[string]any{"pinned": true, "provider": "telegram", "chat_id": chatID, "message_id": messageID, "provider_ok": true}
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(requestInput), provider, delivered, map[string]any{"chat_id": chatID, "message_id": messageID})
}

func (r *Runner) executeTelegramUnpinMessage(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	dep, err := r.loadTelegramDependency(ctx, node.Config.ActionDependencyID, "telegram.unpin_message")
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, telegramDependencyErrorType(err), err)
	}
	chatID := strings.TrimSpace(renderTemplate(firstNonEmpty(node.Config.ChatID, stringMetadata(dep.Metadata, "default_chat_id")), runInput, steps, tpl))
	if chatID == "" {
		err := errors.New("telegram chat_id is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	messageIDRaw := strings.TrimSpace(renderTemplate(node.Config.MessageID, runInput, steps, tpl))
	messageID, err := strconv.ParseInt(messageIDRaw, 10, 64)
	if err != nil || messageID <= 0 {
		err := errors.New("telegram message_id is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	body := map[string]any{"chat_id": chatID, "message_id": messageID}
	requestInput, _ := json.Marshal(map[string]any{
		"chat_id":          chatID,
		"message_id":       messageID,
		"success_criteria": node.Config.SuccessCriteria,
	})
	provider, ex := r.callTelegramJSON(ctx, dep, "telegram.unpin_message", "unpinChatMessage", body, durationOrDefault(node.Config.RetryPolicy.Timeout, 15*time.Second))
	if ex.Err != nil {
		ex.InputJSON = string(requestInput)
		return ex
	}
	delivered := map[string]any{"unpinned": true, "provider": "telegram", "chat_id": chatID, "message_id": messageID, "provider_ok": true}
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(requestInput), provider, delivered, map[string]any{"chat_id": chatID, "message_id": messageID})
}

func telegramExpose(chatID string, provider map[string]any) map[string]any {
	out := map[string]any{"chat_id": chatID}
	if v := provider["message_id"]; v != nil {
		out["message_id"] = v
	}
	return out
}

func renderTelegramReplyMarkup(template string, runInput []byte, steps map[string]any, system templateSystem) (map[string]any, error) {
	template = strings.TrimSpace(template)
	if template == "" {
		return nil, nil
	}
	rendered := strings.TrimSpace(renderTemplate(template, runInput, steps, system))
	if rendered == "" {
		return nil, nil
	}
	var replyMarkup map[string]any
	if err := json.Unmarshal([]byte(rendered), &replyMarkup); err != nil {
		return nil, fmt.Errorf("telegram reply_markup_json must render to a JSON object: %w", err)
	}
	return replyMarkup, nil
}
