package contracts

import (
	"encoding/json"
)

type DraftPatch struct {
	Definition json.RawMessage `json:"draft_definition_json,omitempty"`
	Graph      json.RawMessage `json:"draft_graph_json,omitempty"`
	DSLYAML    *string         `json:"draft_dsl_yaml,omitempty"`
	TestInput  json.RawMessage `json:"test_input_json,omitempty"`
}
