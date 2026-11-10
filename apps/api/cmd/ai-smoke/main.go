package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/platform/httptransport"
)

type modelItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	provider := flag.String("provider", "apinet", "provider key")
	baseURL := flag.String("base-url", envDefault("APINET_BASE_URL", "https://apinet.cloud/v1"), "OpenAI-compatible base URL")
	apiKeyEnv := flag.String("api-key-env", "APINET_API_KEY", "environment variable containing API key")
	listModels := flag.Bool("list-models", false, "list models")
	probeIntents := flag.Bool("probe-intents", false, "run intent classification probe")
	maxOutputTokens := flag.Int("max-output-tokens", 500, "max chat completion tokens")
	flag.String("project-name", "", "project name for smoke prompt")
	flag.Bool("evaluate-models", false, "kept for compatibility; smoke selects the first ranked model")
	flag.Int("max-requests", 10, "maximum requests budget")
	flag.Parse()

	if *provider != "apinet" {
		fail("only apinet smoke is implemented")
	}
	apiKey := strings.TrimSpace(os.Getenv(*apiKeyEnv))
	if apiKey == "" {
		fail("API key env is empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client := httptransport.NewDefaultClient(0)
	models, err := fetchModels(ctx, client, *baseURL, apiKey)
	if err != nil {
		fail(err.Error())
	}
	if *listModels {
		fmt.Printf("provider=apinet models=%d first=%s\n", len(models), firstModelID(models))
	}
	if *probeIntents {
		modelID := firstModelID(models)
		if modelID == "" {
			fail("no model available")
		}
		content, err := chat(ctx, client, *baseURL, apiKey, modelID, *maxOutputTokens)
		if err != nil {
			fail(err.Error())
		}
		fmt.Printf("provider=apinet model=%s intent_probe=%s\n", modelID, compact(content))
	}
}

func fetchModels(ctx context.Context, client *http.Client, baseURL, apiKey string) ([]modelItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("models status %d", resp.StatusCode)
	}
	var root struct {
		Data []modelItem `json:"data"`
	}
	if err := json.Unmarshal(body, &root); err == nil && len(root.Data) > 0 {
		return root.Data, nil
	}
	var arr []modelItem
	if err := json.Unmarshal(body, &arr); err == nil && len(arr) > 0 {
		return arr, nil
	}
	return nil, fmt.Errorf("models response has no data")
}

func chat(ctx context.Context, client *http.Client, baseURL, apiKey, modelID string, maxTokens int) (string, error) {
	payload := map[string]any{
		"model": modelID,
		"messages": []map[string]string{
			{"role": "system", "content": "Return only compact JSON."},
			{"role": "user", "content": `Classify: "Создай задачу проверить оплату Орион завтра". Return {"key":"task.create","client_query":"Орион"}.`},
		},
		"temperature": 0.1,
		"max_tokens":  maxTokens,
	}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("chat status %d", resp.StatusCode)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("chat returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

func firstModelID(models []modelItem) string {
	for _, preferred := range []string{"gpt-5.5", "gpt-5", "gpt-4.1"} {
		for _, item := range models {
			if strings.Contains(strings.ToLower(item.ID), preferred) {
				return item.ID
			}
		}
	}
	if len(models) == 0 {
		return ""
	}
	return models[0].ID
}

func compact(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 300 {
		return value[:300] + "..."
	}
	return value
}

func envDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
