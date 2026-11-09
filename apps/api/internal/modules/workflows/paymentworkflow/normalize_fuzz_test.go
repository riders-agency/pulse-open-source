package paymentworkflow

import (
	"encoding/json"
	"testing"
)

func FuzzNormalizeDraft(f *testing.F) {
	for _, seed := range []struct {
		def   string
		graph string
	}{
		{
			def:   `{"nodes":[{"action_key":"payment_due","action_type_key":"clients.payment_due.trigger","config":{}},{"action_key":"payment_cycle_collect","action_type_key":"clients.payment_cycle.collect","config":{}}],"edges":[{"from":"payment_due","to":"payment_cycle_collect"}]}`,
			graph: `{"nodes":[],"edges":[]}`,
		},
		{
			def:   `{"nodes":[{"action_key":"payment_due","action_type_key":"clients.payment_due.trigger","config":{}},{"action_key":"email_payment_due","action_type_key":"email.send","config":{"body":"{{ steps.payment_cycle_collect.digest_html }}"}}],"edges":[]}`,
			graph: `{"nodes":[{"id":"payment_due","type":"action","position":{"x":360,"y":120},"data":{"action_type_key":"clients.payment_due.trigger","config":{}}}],"edges":[]}`,
		},
		{def: `{"nodes":[],"edges":[]}`, graph: `{"nodes":[],"edges":[]}`},
		{def: `not-json`, graph: `not-json`},
	} {
		f.Add(seed.def, seed.graph)
	}

	f.Fuzz(func(t *testing.T, defJSON, graphJSON string) {
		if len(defJSON) > 1<<15 || len(graphJSON) > 1<<15 {
			t.Skip("payment workflow fuzz input is intentionally capped")
		}
		outDef, outGraph, changed := NormalizeDraft([]byte(defJSON), []byte(graphJSON))
		if !changed {
			return
		}
		if !json.Valid(outDef) {
			t.Fatalf("NormalizeDraft changed definition but returned invalid JSON: %q", string(outDef))
		}
		if !json.Valid(outGraph) {
			t.Fatalf("NormalizeDraft changed graph but returned invalid JSON: %q", string(outGraph))
		}
		var parsed defDoc
		if err := json.Unmarshal(outDef, &parsed); err != nil {
			t.Fatalf("unmarshal changed definition: %v", err)
		}
		if parsed.hasPaymentDue() && !hasDefEdge(parsed.Edges, manualKey, parsed.triggerKey()) {
			t.Fatalf("payment_due workflow must have manual trigger edge: %s", string(outDef))
		}
	})
}
