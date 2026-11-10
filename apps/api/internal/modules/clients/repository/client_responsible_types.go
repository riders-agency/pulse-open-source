package repository

import "github.com/google/uuid"

type ClientResponsibleRow struct {
	ClientResponsibleID  string `json:"client_responsible_id"`
	CompanyID            string `json:"company_id"`
	ClientID             string `json:"client_id"`
	UserID               string `json:"user_id,omitempty"`
	CompanyLegalEntityID string `json:"company_legal_entity_id,omitempty"`
	LegalEntityName      string `json:"legal_entity_name,omitempty"`
	DisplayName          string `json:"display_name"`
	Role                 string `json:"role"`
	IsPrimary            bool   `json:"is_primary"`
	Status               string `json:"status"`
	Comment              string `json:"comment,omitempty"`
}

type ClientResponsibleInput struct {
	UserID               *uuid.UUID `json:"user_id"`
	CompanyLegalEntityID *uuid.UUID `json:"company_legal_entity_id"`
	DisplayName          string     `json:"display_name"`
	Role                 string     `json:"role"`
	IsPrimary            bool       `json:"is_primary"`
	Status               string     `json:"status"`
	Comment              string     `json:"comment"`
}

type ClientResponsiblePatch struct {
	UserID               *uuid.UUID `json:"user_id"`
	CompanyLegalEntityID *uuid.UUID `json:"company_legal_entity_id"`
	DisplayName          *string    `json:"display_name"`
	Role                 *string    `json:"role"`
	IsPrimary            *bool      `json:"is_primary"`
	Status               *string    `json:"status"`
	Comment              *string    `json:"comment"`
}
