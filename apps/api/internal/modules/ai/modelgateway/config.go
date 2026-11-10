package modelgateway

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ProviderKey      string
	BaseURL          string
	APIKey           string
	DefaultModelID   string
	EmbeddingModelID string
	AllowedModels    map[string]struct{}
	DefaultMaxTokens int
	HardMaxTokens    int
	Timeout          time.Duration
}

func LoadConfigFromEnv(environment string) Config {
	provider := firstEnv("PULSE_MODEL_GATEWAY_PROVIDER", "AI_MODEL_PROVIDER")
	if provider == "" && !strings.EqualFold(environment, "production") {
		provider = "fake"
	}
	model := firstEnv("PULSE_MODEL_GATEWAY_MODEL", "APINET_MODEL", "OPENAI_MODEL")
	return Config{
		ProviderKey:      strings.ToLower(strings.TrimSpace(provider)),
		BaseURL:          strings.TrimRight(firstEnv("PULSE_MODEL_GATEWAY_BASE_URL", "APINET_BASE_URL", "OPENAI_BASE_URL"), "/"),
		APIKey:           firstEnv("PULSE_MODEL_GATEWAY_API_KEY", "APINET_API_KEY", "OPENAI_API_KEY"),
		DefaultModelID:   model,
		EmbeddingModelID: firstNonEmpty(firstEnv("PULSE_MODEL_GATEWAY_EMBED_MODEL", "OPENAI_EMBED_MODEL"), model),
		AllowedModels:    parseAllowedModels(firstEnv("PULSE_MODEL_GATEWAY_ALLOWED_MODELS")),
		DefaultMaxTokens: intEnv("PULSE_MODEL_GATEWAY_DEFAULT_MAX_TOKENS", 700),
		HardMaxTokens:    intEnv("PULSE_MODEL_GATEWAY_HARD_MAX_TOKENS", 1800),
		Timeout:          time.Duration(intEnv("PULSE_MODEL_GATEWAY_TIMEOUT_SECONDS", 20)) * time.Second,
	}
}

func (c Config) normalize() Config {
	c.ProviderKey = strings.ToLower(strings.TrimSpace(c.ProviderKey))
	c.BaseURL = strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	c.APIKey = strings.TrimSpace(c.APIKey)
	c.DefaultModelID = strings.TrimSpace(c.DefaultModelID)
	c.EmbeddingModelID = firstNonEmpty(strings.TrimSpace(c.EmbeddingModelID), c.DefaultModelID)
	if c.DefaultMaxTokens <= 0 {
		c.DefaultMaxTokens = 700
	}
	if c.HardMaxTokens <= 0 {
		c.HardMaxTokens = 1800
	}
	if c.Timeout <= 0 {
		c.Timeout = 20 * time.Second
	}
	return c
}

func parseAllowedModels(raw string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, part := range strings.Split(raw, ",") {
		value := strings.TrimSpace(part)
		if value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func intEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// ServerProviderConfigured distinguishes a shared server provider from project credentials.
func ServerProviderConfigured() bool {
	cfg := LoadConfigFromEnv(os.Getenv("RP_ENV"))
	switch cfg.ProviderKey {
	case "apinet", "openai", "openai_compatible":
		return cfg.APIKey != "" && cfg.BaseURL != "" && cfg.DefaultModelID != ""
	default:
		return false
	}
}
