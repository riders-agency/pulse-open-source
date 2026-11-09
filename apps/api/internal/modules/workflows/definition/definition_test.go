package definition

import (
	"encoding/json"
	"testing"
)

func TestNormalizeLegacyDefinitionAddsContractDefaults(t *testing.T) {
	def, err := Normalize(json.RawMessage(`{"nodes":[],"edges":[]}`))
	if err != nil {
		t.Fatalf("normalize legacy: %v", err)
	}
	if def.Format != FormatV1 {
		t.Fatalf("format = %q", def.Format)
	}
	if def.SchemaVersion != SchemaVersion {
		t.Fatalf("schema_version = %d", def.SchemaVersion)
	}
	if len(def.Nodes) != 0 || len(def.Edges) != 0 {
		t.Fatalf("expected empty nodes/edges")
	}
	if def.Settings.Concurrency != "single" || def.Settings.TimeoutSeconds != 300 || def.Settings.OnError != "stop" {
		t.Fatalf("unexpected default settings: %#v", def.Settings)
	}
}

func TestNormalizeV1DefinitionPreservesSettings(t *testing.T) {
	def, err := Normalize(json.RawMessage(`{
	  "format":"riders-pulse-workflow-definition-v1",
	  "schema_version":1,
	  "nodes":[{"id":"a"}],
	  "edges":[{"from":"a","to":"b"}],
	  "settings":{"concurrency":"parallel","timeout_seconds":60,"on_error":"continue"}
	}`))
	if err != nil {
		t.Fatalf("normalize v1: %v", err)
	}
	if len(def.Nodes) != 1 || len(def.Edges) != 1 {
		t.Fatalf("expected nodes/edges to be preserved")
	}
	if def.Settings.Concurrency != "parallel" || def.Settings.TimeoutSeconds != 60 || def.Settings.OnError != "continue" {
		t.Fatalf("unexpected settings: %#v", def.Settings)
	}
}

func TestNormalizeRejectsUnsupportedFormat(t *testing.T) {
	_, err := Normalize(json.RawMessage(`{"format":"other","schema_version":1}`))
	if err == nil {
		t.Fatal("expected unsupported format error")
	}
}

func TestValidateRejectsInvalidSettings(t *testing.T) {
	err := Validate(Definition{
		Format:        FormatV1,
		SchemaVersion: SchemaVersion,
		Settings: Settings{
			Concurrency:    "single",
			TimeoutSeconds: -1,
			OnError:        "stop",
		},
	})
	if err == nil {
		t.Fatal("expected invalid timeout error")
	}
}
