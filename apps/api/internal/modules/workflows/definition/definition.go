package definition

import (
	"encoding/json"
	"fmt"
)

const (
	FormatV1      = "riders-pulse-workflow-definition-v1"
	SchemaVersion = 1
)

type Definition struct {
	Format        string            `json:"format"`
	SchemaVersion int               `json:"schema_version"`
	Nodes         []json.RawMessage `json:"nodes"`
	Edges         []json.RawMessage `json:"edges"`
	Settings      Settings          `json:"settings"`
}

type Settings struct {
	Concurrency    string `json:"concurrency"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	OnError        string `json:"on_error"`
}

func Normalize(raw json.RawMessage) (Definition, error) {
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	var payload struct {
		Format        string            `json:"format"`
		SchemaVersion int               `json:"schema_version"`
		Nodes         []json.RawMessage `json:"nodes"`
		Edges         []json.RawMessage `json:"edges"`
		Settings      Settings          `json:"settings"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return Definition{}, fmt.Errorf("workflow definition json: %w", err)
	}
	def := Definition{
		Format:        payload.Format,
		SchemaVersion: payload.SchemaVersion,
		Nodes:         payload.Nodes,
		Edges:         payload.Edges,
		Settings:      payload.Settings,
	}
	if def.Format == "" {
		def.Format = FormatV1
	}
	if def.SchemaVersion == 0 {
		def.SchemaVersion = SchemaVersion
	}
	if def.Nodes == nil {
		def.Nodes = []json.RawMessage{}
	}
	if def.Edges == nil {
		def.Edges = []json.RawMessage{}
	}
	def.Settings = normalizeSettings(def.Settings)
	if err := Validate(def); err != nil {
		return Definition{}, err
	}
	return def, nil
}

func Validate(def Definition) error {
	if def.Format != FormatV1 {
		return fmt.Errorf("unsupported workflow definition format: %s", def.Format)
	}
	if def.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported workflow definition schema_version: %d", def.SchemaVersion)
	}
	switch def.Settings.Concurrency {
	case "single", "parallel":
	default:
		return fmt.Errorf("unsupported workflow concurrency: %s", def.Settings.Concurrency)
	}
	if def.Settings.TimeoutSeconds <= 0 {
		return fmt.Errorf("workflow timeout_seconds must be positive")
	}
	switch def.Settings.OnError {
	case "stop", "continue":
	default:
		return fmt.Errorf("unsupported workflow on_error: %s", def.Settings.OnError)
	}
	return nil
}

func Marshal(def Definition) (json.RawMessage, error) {
	if err := Validate(def); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(def)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func normalizeSettings(settings Settings) Settings {
	if settings.Concurrency == "" {
		settings.Concurrency = "single"
	}
	if settings.TimeoutSeconds <= 0 {
		settings.TimeoutSeconds = 300
	}
	if settings.OnError == "" {
		settings.OnError = "stop"
	}
	return settings
}
