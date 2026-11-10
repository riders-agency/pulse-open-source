package contracts

import "encoding/json"

type PaymentActionRow struct {
	PaymentActionID string               `json:"payment_action_id"`
	CompanyID       string               `json:"company_id"`
	ClientID        string               `json:"client_id"`
	ClientName      string               `json:"client_name"`
	ProjectID       string               `json:"project_id"`
	ProjectName     string               `json:"project_name"`
	WorkflowRunID   string               `json:"workflow_run_id,omitempty"`
	Status          string               `json:"status"`
	Source          string               `json:"source"`
	PaymentMethod   string               `json:"payment_method"`
	PeriodStart     string               `json:"period_start"`
	PeriodEnd       string               `json:"period_end"`
	DueDate         string               `json:"due_date"`
	TotalAmount     float64              `json:"total_amount"`
	Currency        string               `json:"currency"`
	LastRemindedAt  any                  `json:"last_reminded_at,omitempty"`
	SnoozedUntil    any                  `json:"snoozed_until,omitempty"`
	ConfirmedAt     any                  `json:"confirmed_at,omitempty"`
	Comment         string               `json:"comment,omitempty"`
	Items           []PaymentActionItem  `json:"items,omitempty"`
	Facts           []PaymentFactRow     `json:"facts,omitempty"`
	Documents       []PaymentDocumentRow `json:"documents,omitempty"`
	CreatedAt       any                  `json:"created_at,omitempty"`
}

type PaymentActionItem struct {
	PaymentActionItemID string  `json:"payment_action_item_id"`
	ClientServiceID     string  `json:"client_service_id"`
	ServiceTypeKey      string  `json:"service_type_key,omitempty"`
	Title               string  `json:"title"`
	DueDate             string  `json:"due_date"`
	Amount              float64 `json:"amount"`
	Currency            string  `json:"currency"`
}

type PaymentFactRow struct {
	PaymentFactID string  `json:"payment_fact_id"`
	Source        string  `json:"source"`
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaidAt        any     `json:"paid_at"`
	CreatedAt     any     `json:"created_at"`
}

type PaymentDocumentRow struct {
	PaymentActionDocumentID string `json:"payment_action_document_id"`
	PaymentActionID         string `json:"payment_action_id"`
	FileName                string `json:"file_name"`
	ContentType             string `json:"content_type"`
	SizeBytes               int64  `json:"size_bytes"`
	Kind                    string `json:"kind"`
	DownloadURL             string `json:"download_url,omitempty"`
	CreatedAt               any    `json:"created_at,omitempty"`
}

type PaymentDocumentInput struct {
	FileName    string
	ContentType string
	SizeBytes   int64
	Kind        string
	StoragePath string
}

type PaymentEventRow struct {
	EventID         string  `json:"event_id"`
	EventType       string  `json:"event_type"`
	EventSource     string  `json:"event_source"`
	EventSubject    string  `json:"event_subject"`
	OccurredAt      any     `json:"occurred_at,omitempty"`
	IdempotencyKey  string  `json:"idempotency_key"`
	ClientID        string  `json:"client_id"`
	ClientServiceID string  `json:"client_service_id,omitempty"`
	PaymentActionID string  `json:"payment_action_id,omitempty"`
	PaymentFactID   string  `json:"payment_fact_id,omitempty"`
	Amount          float64 `json:"amount,omitempty"`
	Currency        string  `json:"currency,omitempty"`
	PeriodStart     string  `json:"period_start,omitempty"`
	PeriodEnd       string  `json:"period_end,omitempty"`
	PaidAt          any     `json:"paid_at,omitempty"`
	Source          string  `json:"source"`
	PaymentMethod   string  `json:"payment_method"`
	TriggerType     string  `json:"trigger_type"`
	Status          string  `json:"status"`
	ProcessedAt     any     `json:"processed_at,omitempty"`
}

type ConfirmPaymentInput struct {
	Amount        *float64 `json:"amount"`
	Source        string   `json:"source"`
	PaymentMethod string   `json:"payment_method"`
	PaidAt        string   `json:"paid_at"`
	NextDueDate   string   `json:"next_due_date"`
	Comment       string   `json:"comment"`
	Items         []struct {
		PaymentActionItemID string   `json:"payment_action_item_id"`
		Amount              *float64 `json:"amount"`
	} `json:"items"`
}

type SnoozePaymentInput struct {
	SnoozedUntil string `json:"snoozed_until"`
	Comment      string `json:"comment"`
}

type ManualPaymentRecordInput struct {
	ProjectID       string          `json:"project_id"`
	ClientServiceID string          `json:"client_service_id"`
	Amount          *float64        `json:"amount"`
	Currency        string          `json:"currency"`
	PeriodStart     string          `json:"period_start"`
	PeriodEnd       string          `json:"period_end"`
	PaidAt          string          `json:"paid_at"`
	PaymentMethod   string          `json:"payment_method"`
	NextDueDate     string          `json:"next_due_date"`
	Comment         string          `json:"comment"`
	IdempotencyKey  string          `json:"idempotency_key"`
	MetadataJSON    json.RawMessage `json:"metadata"`
}

type ManualPaymentRecordResult struct {
	EventID         string `json:"event_id"`
	EventType       string `json:"event_type"`
	EventSource     string `json:"event_source"`
	EventSubject    string `json:"event_subject"`
	IdempotencyKey  string `json:"idempotency_key"`
	PaymentActionID string `json:"payment_action_id"`
	PaymentFactID   string `json:"payment_fact_id"`
	Status          string `json:"status"`
	Duplicate       bool   `json:"duplicate"`
}
