package contracts_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	aiagentv1 "riders.agency/pulse/api/internal/gen/ai_agent/v1"
)

type protoFixture struct {
	TurnRequest                 json.RawMessage   `json:"turnRequest"`
	ToolSpec                    json.RawMessage   `json:"toolSpec"`
	ModelChatRequest            json.RawMessage   `json:"modelChatRequest"`
	ExpectedDeterministicSHA256 map[string]string `json:"expectedDeterministicSha256"`
}

func TestAIProtoSerializationFixture(t *testing.T) {
	fixture := loadProtoFixture(t)

	cases := []struct {
		name string
		raw  json.RawMessage
		msg  proto.Message
	}{
		{name: "turnRequest", raw: fixture.TurnRequest, msg: &aiagentv1.TurnRequest{}},
		{name: "toolSpec", raw: fixture.ToolSpec, msg: &aiagentv1.ToolSpec{}},
		{name: "modelChatRequest", raw: fixture.ModelChatRequest, msg: &aiagentv1.ModelChatRequest{}},
	}

	unmarshal := protojson.UnmarshalOptions{DiscardUnknown: false}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := unmarshal.Unmarshal(tc.raw, tc.msg); err != nil {
				t.Fatalf("unmarshal fixture: %v", err)
			}
			deterministic, err := (proto.MarshalOptions{Deterministic: true}).Marshal(tc.msg)
			if err != nil {
				t.Fatalf("marshal deterministic protobuf: %v", err)
			}
			sum := sha256.Sum256(deterministic)
			got := hex.EncodeToString(sum[:])
			if got != fixture.ExpectedDeterministicSHA256[tc.name] {
				t.Fatalf("deterministic protobuf sha256 = %s, want %s", got, fixture.ExpectedDeterministicSHA256[tc.name])
			}
			if _, err := (protojson.MarshalOptions{UseProtoNames: false}).Marshal(tc.msg); err != nil {
				t.Fatalf("marshal canonical protojson: %v", err)
			}
		})
	}
}

func loadProtoFixture(t *testing.T) protoFixture {
	t.Helper()
	path := filepath.Join(repoRoot(t), "contracts", "proto", "fixtures", "ai_agent_v1_serialization.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture protoFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return fixture
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "contracts", "proto", "fixtures", "ai_agent_v1_serialization.json")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("repository root not found from %s", wd)
		}
	}
}
