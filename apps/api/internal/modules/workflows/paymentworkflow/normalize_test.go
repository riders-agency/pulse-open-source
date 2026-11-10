package paymentworkflow

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeDraftRemovesCollect(t *testing.T) {
	def := []byte(`{
	  "nodes": [
	    {"action_key":"payment_due","action_type_key":"clients.payment_due.trigger","label":"Срок оплаты","config":{}},
	    {"action_key":"payment_cycle_collect","action_type_key":"clients.payment_cycle.collect","label":"Собрать","config":{}},
	    {"action_key":"telegram_payment_due","action_type_key":"telegram.send_message","label":"TG","config":{"text":"{{ steps.payment_cycle_collect.digest_telegram }}"}}
	  ],
	  "edges": [
	    {"from":"payment_due","to":"payment_cycle_collect"},
	    {"from":"payment_cycle_collect","to":"telegram_payment_due"}
	  ]
	}`)
	graph := []byte(`{"nodes":[],"edges":[]}`)
	outDef, outGraph, changed := NormalizeDraft(def, graph)
	if !changed {
		t.Fatal("expected change")
	}
	var parsed defDoc
	if err := json.Unmarshal(outDef, &parsed); err != nil {
		t.Fatal(err)
	}
	for _, n := range parsed.Nodes {
		if n.ActionTypeKey == collectType {
			t.Fatal("collect node should be removed")
		}
	}
	if !strings.Contains(string(outDef), "steps.payment_due.digest_telegram") {
		t.Fatalf("expected payment_due template, got %s", outDef)
	}
	found := false
	foundManual := false
	for _, e := range parsed.Edges {
		if e.From == "payment_due" && e.To == "telegram_payment_due" {
			found = true
		}
		if e.From == manualKey && e.To == "payment_due" {
			foundManual = true
		}
	}
	if !found {
		t.Fatalf("expected direct edge, edges=%v", parsed.Edges)
	}
	if !foundManual {
		t.Fatalf("expected manual payment edge, edges=%v", parsed.Edges)
	}
	var g graphDoc
	_ = json.Unmarshal(outGraph, &g)
	if len(g.Edges) == 0 {
		t.Fatal("expected graph edges rebuilt")
	}
}

func TestNormalizeDraftAddsManualPaymentTrigger(t *testing.T) {
	def := []byte(`{
	  "nodes": [
	    {"action_key":"payment_due","action_type_key":"clients.payment_due.trigger","label":"Срок оплаты","config":{}},
	    {"action_key":"email_payment_due","action_type_key":"email.send","label":"Email","config":{"body":"{{ steps.payment_due.digest_html }}"}}
	  ],
	  "edges": [
	    {"from":"payment_due","to":"email_payment_due"}
	  ]
	}`)
	graph := []byte(`{
	  "nodes": [
	    {"id":"payment_due","type":"action","position":{"x":360,"y":120},"data":{"label":"Срок оплаты","action_type_key":"clients.payment_due.trigger","config":{}}},
	    {"id":"email_payment_due","type":"action","position":{"x":680,"y":120},"data":{"label":"Email","action_type_key":"email.send","config":{}}}
	  ],
	  "edges": [
	    {"id":"e-payment_due-email_payment_due","source":"payment_due","target":"email_payment_due"}
	  ]
	}`)
	outDef, outGraph, changed := NormalizeDraft(def, graph)
	if !changed {
		t.Fatal("expected change")
	}
	var parsed defDoc
	if err := json.Unmarshal(outDef, &parsed); err != nil {
		t.Fatal(err)
	}
	hasManualNode := false
	hasManualEdge := false
	for _, n := range parsed.Nodes {
		if n.ActionKey == manualKey && n.ActionTypeKey == manualType {
			hasManualNode = true
		}
	}
	for _, e := range parsed.Edges {
		if e.From == manualKey && e.To == "payment_due" {
			hasManualEdge = true
		}
	}
	if !hasManualNode || !hasManualEdge {
		t.Fatalf("expected manual trigger node and edge, def=%s", string(outDef))
	}
	var g graphDoc
	if err := json.Unmarshal(outGraph, &g); err != nil {
		t.Fatal(err)
	}
	if !hasGraphEdge(g.Edges, manualKey, "payment_due") {
		t.Fatalf("expected graph manual edge, graph=%s", string(outGraph))
	}
}
