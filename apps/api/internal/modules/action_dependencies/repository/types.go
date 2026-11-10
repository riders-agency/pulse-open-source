package repository

import (
	"encoding/json"
)

type DepCreate struct {
	Name              string          `json:"name"`
	DependencyTypeKey string          `json:"dependency_type_key"`
	PlaintextSecret   string          `json:"plaintext_secret"`
	MetadataJSON      json.RawMessage `json:"metadata_json"`
}

type DepPatch struct {
	Name         *string         `json:"name"`
	Status       *string         `json:"status"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
}

type TelegramRelayPollingPatch struct {
	Enabled            bool `json:"enabled"`
	TimeoutSeconds     int  `json:"timeout_seconds"`
	Limit              int  `json:"limit"`
	DeleteWebhook      bool `json:"delete_webhook"`
	DropPendingUpdates bool `json:"drop_pending_updates"`
}
