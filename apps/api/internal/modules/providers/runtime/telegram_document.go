package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	neturlpkg "net/url"
	"strings"
	"time"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
	"riders.agency/pulse/api/internal/platform/httptransport"
)

func (r *Runner) executeTelegramDocument(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	dep, err := r.loadDependency(ctx, node.Config.ActionDependencyID)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "dependency", err)
	}
	if dep.TypeKey != "telegram.bot" && dep.TypeKey != "telegram.relay" {
		err := fmt.Errorf("telegram.send_document requires telegram.bot or telegram.relay dependency")
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
	docURL := strings.TrimSpace(renderTemplate(firstNonEmpty(node.Config.DocumentURL, node.Config.Text), runInput, steps, tpl))
	if docURL == "" {
		err := errors.New("telegram document_url is required")
		return failedAction(input, "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	caption := strings.TrimSpace(renderTemplate(node.Config.Caption, runInput, steps, tpl))

	requestInput, _ := json.Marshal(map[string]any{
		"chat_id":          chatID,
		"document_url":     docURL,
		"caption":          caption,
		"success_criteria": node.Config.SuccessCriteria,
	})
	timeout := durationOrDefault(node.Config.RetryPolicy.Timeout, 60*time.Second)

	downloadCtx, cancelDownload := context.WithTimeout(ctx, timeout)
	fileBytes, fileName, err := downloadDocument(downloadCtx, r.httpClient(), docURL)
	cancelDownload()
	if err != nil {
		return failedAction(string(requestInput), "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, true, "download", err)
	}

	tgURL := "https://api.telegram.org/bot" + dep.Plaintext + "/sendDocument"
	body, contentType, err := buildTelegramDocumentMultipart(chatID, caption, fileName, fileBytes, node.Config.DisableNotification)
	if err != nil {
		return failedAction(string(requestInput), "{}", map[string]any{"provider": "telegram", "ok": false, "error": err.Error()}, false, "request", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, tgURL, body)
	if err != nil {
		cancel()
		return failedAction(string(requestInput), "{}", map[string]any{"provider": "telegram", "ok": false, "error": "request build failed"}, false, "request", errors.New("telegram request build failed"))
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := r.httpClient().Do(req)
	cancel()
	if err != nil {
		return failedAction(string(requestInput), "{}", map[string]any{"provider": "telegram", "ok": false, "error": "telegram request failed"}, true, "network", errors.New("telegram request failed"))
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
	provider := sanitizeTelegramResponse(resp.StatusCode, raw)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		providerJSON, _ := json.Marshal(provider)
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		err := fmt.Errorf("telegram returned HTTP %d", resp.StatusCode)
		report := diagnosticsForFailure(node.ActionTypeKey, provider, retryable, "provider_http", err)
		output := providerdiagnostics.OutputWithReport("{}", report)
		return actionExecution{InputJSON: string(requestInput), OutputJSON: output, ProviderResponseJSON: string(providerJSON), Retryable: retryable, ErrorType: "provider_http", Err: err}
	}

	delivered := map[string]any{
		"delivered":   true,
		"provider":    "telegram",
		"message_id":  provider["message_id"],
		"chat_id":     chatID,
		"provider_ok": true,
		"document":    fileName,
	}
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(requestInput), provider, delivered, telegramExpose(chatID, provider))
}

func downloadDocument(ctx context.Context, client *http.Client, docURL string) ([]byte, string, error) {
	if client == nil {
		client = httptransport.NewDefaultClient(30 * time.Second)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, docURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("document download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("document download HTTP %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return nil, "", err
	}
	fileName := "document.pdf"
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if fn := params["filename"]; fn != "" {
				fileName = fn
			}
		}
	}
	if fileName == "document.pdf" {
		if u, err := neturlpkg.Parse(docURL); err == nil {
			if base := strings.TrimSpace(u.Path); base != "" {
				if leaf := strings.TrimPrefix(base, "/"); leaf != "" {
					if idx := strings.LastIndex(leaf, "/"); idx >= 0 {
						leaf = leaf[idx+1:]
					}
					if leaf != "" {
						fileName = leaf
					}
				}
			}
		}
	}
	return raw, fileName, nil
}

func buildTelegramDocumentMultipart(chatID, caption, fileName string, fileBytes []byte, disableNotification bool) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return nil, "", err
	}
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return nil, "", err
		}
	}
	if disableNotification {
		if err := writer.WriteField("disable_notification", "true"); err != nil {
			return nil, "", err
		}
	}
	part, err := writer.CreateFormFile("document", fileName)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(fileBytes); err != nil {
		return nil, "", err
	}
	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return &buf, contentType, nil
}
