package retry

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

type Class string

const (
	ClassValidation     Class = "validation"
	ClassAuth           Class = "auth"
	ClassPermanent      Class = "permanent"
	ClassTransient      Class = "transient"
	ClassRateLimited    Class = "rate_limited"
	ClassTimeoutUnknown Class = "timeout_unknown"
	ClassCanceled       Class = "canceled"
	ClassUnknownOutcome Class = "unknown_outcome"
)

type Policy struct {
	Owner        string
	InitialDelay time.Duration
	MaxDelay     time.Duration
	MaxElapsed   time.Duration
	JitterRatio  float64
}

type Failure struct {
	ActionTypeKey        string
	AttemptNo            int
	MaxAttempts          int
	Retryable            bool
	ErrorType            string
	Err                  error
	ProviderResponseJSON string
	StartedAt            time.Time
	Now                  time.Time
	Seed                 string
	InitialDelay         time.Duration
	MaxDelay             time.Duration
	MaxElapsed           time.Duration
}

type Evaluation struct {
	Class          Class         `json:"class"`
	Retryable      bool          `json:"retryable"`
	ShouldRetry    bool          `json:"should_retry"`
	Exhausted      bool          `json:"exhausted"`
	AttemptNo      int           `json:"attempt_no"`
	MaxAttempts    int           `json:"max_attempts"`
	NextDelay      time.Duration `json:"-"`
	NextDelayMS    int64         `json:"next_delay_ms,omitempty"`
	MaxElapsedMS   int64         `json:"max_elapsed_ms"`
	Owner          string        `json:"owner"`
	RedactedReason string        `json:"redacted_reason"`
}

func Evaluate(f Failure) Evaluation {
	now := f.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	policy := PolicyForAction(f.ActionTypeKey)
	if f.InitialDelay > 0 {
		policy.InitialDelay = f.InitialDelay
	}
	if f.MaxDelay > 0 {
		policy.MaxDelay = f.MaxDelay
	}
	if f.MaxElapsed > 0 {
		policy.MaxElapsed = f.MaxElapsed
	}
	attemptNo := f.AttemptNo
	if attemptNo < 1 {
		attemptNo = 1
	}
	maxAttempts := f.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	class := Classify(f.ErrorType, f.Err, f.Retryable, f.ProviderResponseJSON)
	retryable := f.Retryable && classRetryable(class)
	delay := NextDelay(policy, attemptNo, f.Seed)
	exhausted := attemptNo >= maxAttempts
	if !f.StartedAt.IsZero() && policy.MaxElapsed > 0 && now.Sub(f.StartedAt)+delay > policy.MaxElapsed {
		exhausted = true
	}
	shouldRetry := retryable && !exhausted
	if !shouldRetry {
		delay = 0
	}
	return Evaluation{
		Class:          class,
		Retryable:      retryable,
		ShouldRetry:    shouldRetry,
		Exhausted:      exhausted,
		AttemptNo:      attemptNo,
		MaxAttempts:    maxAttempts,
		NextDelay:      delay,
		NextDelayMS:    delay.Milliseconds(),
		MaxElapsedMS:   policy.MaxElapsed.Milliseconds(),
		Owner:          policy.Owner,
		RedactedReason: Redact(errorMessage(f.Err)),
	}
}

func PolicyForAction(actionTypeKey string) Policy {
	policy := Policy{
		Owner:        "runtime.retry",
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		MaxElapsed:   15 * time.Minute,
		JitterRatio:  0.2,
	}
	actionTypeKey = strings.TrimSpace(actionTypeKey)
	switch {
	case strings.HasPrefix(actionTypeKey, "telegram."):
		policy.Owner = "provider.telegram"
		policy.InitialDelay = 3 * time.Second
	case strings.HasPrefix(actionTypeKey, "ai."):
		policy.Owner = "provider.ai"
		policy.InitialDelay = 2 * time.Second
	case strings.HasPrefix(actionTypeKey, "max."):
		policy.Owner = "provider.max"
		policy.InitialDelay = 3 * time.Second
	case actionTypeKey == "email.send":
		policy.Owner = "provider.smtp"
		policy.InitialDelay = 5 * time.Second
	case actionTypeKey == "http.request":
		policy.Owner = "provider.http"
		policy.InitialDelay = 2 * time.Second
	case actionTypeKey == "uon.create_lead":
		policy.Owner = "provider.uon"
		policy.InitialDelay = 3 * time.Second
	case strings.HasPrefix(actionTypeKey, "clients."):
		policy.Owner = "module.clients"
		policy.InitialDelay = 500 * time.Millisecond
		policy.MaxDelay = 5 * time.Second
		policy.MaxElapsed = 2 * time.Minute
	}
	return policy
}

func Classify(errorType string, err error, retryable bool, providerResponseJSON string) Class {
	typ := strings.ToLower(strings.TrimSpace(errorType))
	msg := strings.ToLower(errorMessage(err))
	if errors.Is(err, context.Canceled) || strings.Contains(msg, "context canceled") || typ == "canceled" {
		return ClassCanceled
	}
	if typ == "unknown_outcome" {
		return ClassUnknownOutcome
	}
	if typ == "auth" || strings.Contains(msg, "unauthorized") || strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "invalid token") || strings.Contains(msg, "auth failed") {
		return ClassAuth
	}
	if typ == "validation" || typ == "config" || typ == "request" || typ == "success_criteria" {
		return ClassValidation
	}
	if typ == "missing_handler" || typ == "dependency_type" || typ == "permanent" {
		return ClassPermanent
	}
	if typ == "provider_http" {
		return classifyHTTP(providerResponseJSON, retryable)
	}
	if typ == "rate_limited" || strings.Contains(msg, "rate limit") || strings.Contains(msg, "too many requests") {
		return ClassRateLimited
	}
	if typ == "network" || typ == "timeout" || errors.Is(err, context.DeadlineExceeded) || strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") {
			return ClassTimeoutUnknown
		}
		return ClassTransient
	}
	if typ == "dependency" || typ == "runtime" || typ == "smtp" || typ == "download" {
		if retryable {
			return ClassTransient
		}
		return ClassPermanent
	}
	if retryable {
		return ClassTransient
	}
	return ClassPermanent
}

func NextDelay(policy Policy, attemptNo int, seed string) time.Duration {
	if attemptNo < 1 {
		attemptNo = 1
	}
	if policy.InitialDelay <= 0 {
		policy.InitialDelay = time.Second
	}
	if policy.MaxDelay <= 0 {
		policy.MaxDelay = policy.InitialDelay
	}
	exp := math.Pow(2, float64(attemptNo-1))
	delay := time.Duration(float64(policy.InitialDelay) * exp)
	if delay > policy.MaxDelay {
		delay = policy.MaxDelay
	}
	if policy.JitterRatio > 0 && seed != "" {
		jitter := deterministicUnit(seed)
		factor := 1 + ((jitter*2 - 1) * policy.JitterRatio)
		delay = time.Duration(float64(delay) * factor)
	}
	if delay < time.Millisecond {
		return time.Millisecond
	}
	return delay.Truncate(time.Millisecond)
}

func classRetryable(class Class) bool {
	switch class {
	case ClassTransient, ClassRateLimited:
		return true
	default:
		return false
	}
}

func classifyHTTP(providerResponseJSON string, retryable bool) Class {
	status := httpStatus(providerResponseJSON)
	switch {
	case status == 401 || status == 403:
		return ClassAuth
	case status == 408 || status == 504:
		return ClassTimeoutUnknown
	case status == 429:
		return ClassRateLimited
	case status >= 500:
		return ClassTransient
	case status >= 400:
		return ClassPermanent
	default:
		if retryable {
			return ClassTransient
		}
		return ClassPermanent
	}
}

func httpStatus(raw string) int {
	var root any
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &root); err != nil {
		return 0
	}
	return findStatus(root)
}

func findStatus(value any) int {
	switch v := value.(type) {
	case map[string]any:
		for _, key := range []string{"status_code", "http_status", "status", "code", "http"} {
			if n := number(v[key]); n > 0 {
				return n
			}
		}
		for _, nested := range v {
			if n := findStatus(nested); n > 0 {
				return n
			}
		}
	case []any:
		for _, item := range v {
			if n := findStatus(item); n > 0 {
				return n
			}
		}
	}
	return 0
}

func number(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	case string:
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		return n
	default:
		return 0
	}
}

var (
	emailPattern  = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	bearerPattern = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/\-=]+`)
	secretPattern = regexp.MustCompile(`(?i)\b(token|secret|password|authorization|api[_-]?key)\s*[:=]\s*[^\s,;]+`)
)

func Redact(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return "action failed"
	}
	message = emailPattern.ReplaceAllString(message, "[redacted-email]")
	message = bearerPattern.ReplaceAllString(message, "Bearer [redacted]")
	message = secretPattern.ReplaceAllString(message, "$1=[redacted]")
	if len(message) > 240 {
		message = strings.TrimSpace(message[:240]) + "..."
	}
	return message
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func deterministicUnit(seed string) float64 {
	sum := sha256.Sum256([]byte(seed))
	n := binary.BigEndian.Uint64(sum[:8])
	return float64(n) / float64(math.MaxUint64)
}
