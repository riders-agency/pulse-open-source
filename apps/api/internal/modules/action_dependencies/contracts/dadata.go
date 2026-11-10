package contracts

import "errors"

var ErrDaDataNotConfigured = errors.New("dadata.party integration is not configured")

type DaDataIntegrationRow struct {
	Configured         bool   `json:"configured"`
	ActionDependencyID string `json:"action_dependency_id,omitempty"`
	Name               string `json:"name,omitempty"`
	BaseURL            string `json:"base_url,omitempty"`
	UpdatedAt          any    `json:"updated_at,omitempty"`
	Source             string `json:"source,omitempty"`
}

type DaDataIntegrationInput struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

type DaDataLookupConfig struct {
	APIKey  string
	BaseURL string
}
