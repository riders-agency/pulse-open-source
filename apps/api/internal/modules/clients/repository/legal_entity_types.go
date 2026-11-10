package repository

import "encoding/json"

type CompanyLegalEntityRow struct {
	CompanyLegalEntityID string          `json:"company_legal_entity_id"`
	CompanyID            string          `json:"company_id"`
	EntityType           string          `json:"entity_type"`
	Status               string          `json:"status"`
	Name                 string          `json:"name"`
	ShortName            string          `json:"short_name,omitempty"`
	FullName             string          `json:"full_name,omitempty"`
	INN                  string          `json:"inn,omitempty"`
	KPP                  string          `json:"kpp,omitempty"`
	OGRN                 string          `json:"ogrn,omitempty"`
	OGRNIP               string          `json:"ogrnip,omitempty"`
	LegalAddress         string          `json:"legal_address,omitempty"`
	ActualAddress        string          `json:"actual_address,omitempty"`
	OKVED                string          `json:"okved,omitempty"`
	Comment              string          `json:"comment,omitempty"`
	DaDataRawJSON        json.RawMessage `json:"dadata_raw_json,omitempty"`
	CreatedAt            any             `json:"created_at,omitempty"`
	UpdatedAt            any             `json:"updated_at,omitempty"`
}

type CompanyLegalEntityInput struct {
	EntityType    string          `json:"entity_type"`
	Status        string          `json:"status"`
	Name          string          `json:"name"`
	ShortName     string          `json:"short_name"`
	FullName      string          `json:"full_name"`
	INN           string          `json:"inn"`
	KPP           string          `json:"kpp"`
	OGRN          string          `json:"ogrn"`
	OGRNIP        string          `json:"ogrnip"`
	LegalAddress  string          `json:"legal_address"`
	ActualAddress string          `json:"actual_address"`
	OKVED         string          `json:"okved"`
	Comment       string          `json:"comment"`
	DaDataRawJSON json.RawMessage `json:"dadata_raw_json"`
}

type CompanyLegalEntityPatch struct {
	EntityType    *string         `json:"entity_type"`
	Status        *string         `json:"status"`
	Name          *string         `json:"name"`
	ShortName     *string         `json:"short_name"`
	FullName      *string         `json:"full_name"`
	INN           *string         `json:"inn"`
	KPP           *string         `json:"kpp"`
	OGRN          *string         `json:"ogrn"`
	OGRNIP        *string         `json:"ogrnip"`
	LegalAddress  *string         `json:"legal_address"`
	ActualAddress *string         `json:"actual_address"`
	OKVED         *string         `json:"okved"`
	Comment       *string         `json:"comment"`
	DaDataRawJSON json.RawMessage `json:"dadata_raw_json"`
}
