package office

import (
	"encoding/json"
	"riders.agency/pulse/api/internal/platform/errs"
)

// Demo visitors may configure or disable individual steps, never add executable tools.
// The apply handler independently requires the durable human decision.
func ValidateDefinition(raw []byte) error {
	var def struct {
		Nodes []struct {
			Key  string `json:"action_key"`
			Type string `json:"action_type_key"`
		} `json:"nodes"`
		Edges []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"edges"`
	}
	if json.Unmarshal(raw, &def) != nil || len(def.Nodes) != len(actions) || len(def.Edges) != len(actions)-1 {
		return errs.ErrForbidden
	}
	expected := map[string]string{}
	edges := map[string]bool{}
	for i, a := range actions {
		expected[a.Key] = a.Type
		if i > 0 {
			edges[actions[i-1].Key+":"+a.Key] = true
		}
	}
	for _, n := range def.Nodes {
		if expected[n.Key] != n.Type || n.Type == "" {
			return errs.ErrForbidden
		}
		delete(expected, n.Key)
	}
	for _, e := range def.Edges {
		key := e.From + ":" + e.To
		if !edges[key] {
			return errs.ErrForbidden
		}
		delete(edges, key)
	}
	return nil
}
