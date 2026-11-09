package providers

import (
	"encoding/json"
	"testing"
)

func TestCatalogListFieldsSerializeAsArrays(t *testing.T) {
	raw, err := json.Marshal(Catalog())
	if err != nil {
		t.Fatalf("marshal catalog: %v", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("unmarshal catalog: %v", err)
	}
	for _, row := range rows {
		key, _ := row["action_type_key"].(string)
		for _, field := range []string{"criteria", "failure_categories", "expose", "template_scopes"} {
			if _, ok := row[field].([]any); !ok {
				t.Fatalf("%s.%s serialized as %T, want array", key, field, row[field])
			}
		}
	}
}
