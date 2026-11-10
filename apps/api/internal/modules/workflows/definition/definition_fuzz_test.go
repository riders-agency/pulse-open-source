package definition

import (
	"encoding/json"
	"testing"
)

func FuzzNormalizeRoundTrip(f *testing.F) {
	for _, seed := range []string{
		`{}`,
		`{"nodes":[],"edges":[]}`,
		`{"format":"riders-pulse-workflow-definition-v1","schema_version":1,"nodes":[{"id":"start"}],"edges":[],"settings":{"concurrency":"single","timeout_seconds":30,"on_error":"stop"}}`,
		`{"format":"other","schema_version":1}`,
		`{"settings":{"timeout_seconds":-1}}`,
		`not-json`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 1<<15 {
			t.Skip("definition fuzz input is intentionally capped")
		}
		def, err := Normalize(json.RawMessage(raw))
		if err != nil {
			return
		}
		if def.Format != FormatV1 {
			t.Fatalf("normalized format = %q", def.Format)
		}
		if def.SchemaVersion != SchemaVersion {
			t.Fatalf("normalized schema_version = %d", def.SchemaVersion)
		}
		if def.Nodes == nil || def.Edges == nil {
			t.Fatalf("normalized nodes/edges must be non-nil: %#v", def)
		}
		if def.Settings.Concurrency == "" || def.Settings.TimeoutSeconds <= 0 || def.Settings.OnError == "" {
			t.Fatalf("normalized settings must be complete: %#v", def.Settings)
		}

		marshaled, err := Marshal(def)
		if err != nil {
			t.Fatalf("marshal normalized definition: %v", err)
		}
		if !json.Valid(marshaled) {
			t.Fatalf("marshal returned invalid JSON: %q", string(marshaled))
		}
		roundTripped, err := Normalize(marshaled)
		if err != nil {
			t.Fatalf("normalize marshaled definition: %v", err)
		}
		if roundTripped.Settings != def.Settings || len(roundTripped.Nodes) != len(def.Nodes) || len(roundTripped.Edges) != len(def.Edges) {
			t.Fatalf("round trip changed contract fields: before=%#v after=%#v", def, roundTripped)
		}
	})
}
