package repository

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ServiceTypeRow struct {
	ClientServiceTypeID string `json:"client_service_type_id"`
	CompanyID           string `json:"company_id,omitempty"`
	ProjectID           string `json:"project_id,omitempty"`
	ScopeKind           string `json:"scope_kind,omitempty"`
	ServiceTypeKey      string `json:"service_type_key"`
	DisplayName         string `json:"display_name"`
	Description         string `json:"description,omitempty"`
	Status              string `json:"status,omitempty"`
	SortOrder           int    `json:"sort_order"`
}

type ServiceTypeInput struct {
	ServiceTypeKey string `json:"service_type_key"`
	DisplayName    string `json:"display_name"`
	Description    string `json:"description"`
	SortOrder      int    `json:"sort_order"`
	Status         string `json:"status"`
}

type ServiceTypePatch struct {
	ServiceTypeKey *string `json:"service_type_key"`
	DisplayName    *string `json:"display_name"`
	Description    *string `json:"description"`
	SortOrder      *int    `json:"sort_order"`
	Status         *string `json:"status"`
}

type ServiceRow struct {
	ClientServiceID     string          `json:"client_service_id"`
	CompanyID           string          `json:"company_id"`
	ClientID            string          `json:"client_id"`
	ProjectID           string          `json:"project_id"`
	ClientServiceTypeID string          `json:"client_service_type_id"`
	ServiceTypeKey      string          `json:"service_type_key"`
	ServiceTypeName     string          `json:"service_type_name"`
	Name                string          `json:"name"`
	Status              string          `json:"status"`
	Amount              float64         `json:"amount"`
	RawPriceText        string          `json:"raw_price_text,omitempty"`
	Currency            string          `json:"currency"`
	BillingPeriod       string          `json:"billing_period"`
	NextDueDate         string          `json:"next_due_date,omitempty"`
	RemindDaysBefore    int             `json:"remind_days_before"`
	RemindEveryDays     int             `json:"remind_every_days"`
	MetadataJSON        json.RawMessage `json:"metadata_json,omitempty"`
}

type ServiceInput struct {
	ProjectID           uuid.UUID       `json:"project_id"`
	ClientServiceTypeID uuid.UUID       `json:"client_service_type_id"`
	ServiceTypeKey      string          `json:"service_type_key"`
	Name                string          `json:"name"`
	Status              string          `json:"status"`
	Amount              *float64        `json:"amount"`
	RawPriceText        string          `json:"raw_price_text"`
	Currency            string          `json:"currency"`
	BillingPeriod       string          `json:"billing_period"`
	NextDueDate         string          `json:"next_due_date"`
	RemindDaysBefore    int             `json:"remind_days_before"`
	RemindEveryDays     int             `json:"remind_every_days"`
	MetadataJSON        json.RawMessage `json:"metadata_json"`
}

type ServicePatch struct {
	ProjectID           *uuid.UUID      `json:"project_id"`
	ClientServiceTypeID *uuid.UUID      `json:"client_service_type_id"`
	ServiceTypeKey      *string         `json:"service_type_key"`
	Name                *string         `json:"name"`
	Status              *string         `json:"status"`
	Amount              *float64        `json:"amount"`
	RawPriceText        *string         `json:"raw_price_text"`
	Currency            *string         `json:"currency"`
	BillingPeriod       *string         `json:"billing_period"`
	NextDueDate         *string         `json:"next_due_date"`
	RemindDaysBefore    *int            `json:"remind_days_before"`
	RemindEveryDays     *int            `json:"remind_every_days"`
	MetadataJSON        json.RawMessage `json:"metadata_json"`
}
