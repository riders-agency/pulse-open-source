package office

import (
	"encoding/json"
	"testing"
)

func TestDepartmentPermissions(t *testing.T) {
	if _, err := departmentRoute(map[string]any{"classify": map[string]any{"departments": []any{"admin"}}}); err == nil {
		t.Fatal("unknown department accepted")
	}
	def, _ := Definition()
	var doc struct{ Nodes []json.RawMessage }
	if json.Unmarshal(def, &doc) != nil || len(doc.Nodes) != 7 {
		t.Fatal("incomplete workflow")
	}
}
