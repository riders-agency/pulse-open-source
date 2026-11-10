package paymentworkflow

import (
	"encoding/json"
	"strings"
)

const collectType = "clients.payment_cycle.collect"
const dueType = "clients.payment_due.trigger"
const manualType = "clients.manual_payment.trigger"
const manualKey = "manual_payment_trigger"
const manualLabel = "Счет оплачен"

// NormalizeDraft keeps payment_due workflows in the canonical shape:
// manual_payment_trigger -> payment_due -> notification actions.
func NormalizeDraft(defJSON, graphJSON []byte) (json.RawMessage, json.RawMessage, bool) {
	def, ok := parseDef(defJSON)
	if !ok || !def.hasPaymentDue() {
		return defJSON, graphJSON, false
	}
	changed := def.stripCollectAndRewire()
	graph, graphChanged := normalizeGraph(graphJSON, def)
	if graphChanged {
		changed = true
	}
	if !changed {
		return defJSON, graphJSON, false
	}
	outDef, _ := json.Marshal(def)
	outGraph, _ := json.Marshal(graph)
	return outDef, outGraph, true
}

type defDoc struct {
	Nodes []defNode `json:"nodes"`
	Edges []defEdge `json:"edges"`
}

type defNode struct {
	ActionKey     string         `json:"action_key"`
	Label         string         `json:"label"`
	ActionTypeKey string         `json:"action_type_key"`
	Category      string         `json:"category,omitempty"`
	Config        map[string]any `json:"config"`
}

type defEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

type graphDoc struct {
	Nodes []graphNode `json:"nodes"`
	Edges []graphEdge `json:"edges"`
}

type graphNode struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Position map[string]any `json:"position"`
	Data     graphNodeData  `json:"data"`
}

type graphNodeData struct {
	Label         string         `json:"label"`
	ActionTypeKey string         `json:"action_type_key"`
	Config        map[string]any `json:"config"`
}

type graphEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

func parseDef(raw []byte) (defDoc, bool) {
	var def defDoc
	if len(strings.TrimSpace(string(raw))) == 0 {
		return def, false
	}
	if err := json.Unmarshal(raw, &def); err != nil {
		return def, false
	}
	return def, true
}

func (d defDoc) hasPaymentDue() bool {
	for _, n := range d.Nodes {
		if n.ActionTypeKey == dueType {
			return true
		}
	}
	return false
}

func (d *defDoc) triggerKey() string {
	for _, n := range d.Nodes {
		if n.ActionTypeKey == dueType && strings.TrimSpace(n.ActionKey) != "" {
			return strings.TrimSpace(n.ActionKey)
		}
	}
	return "payment_due"
}

func (d *defDoc) stripCollectAndRewire() bool {
	trigger := d.triggerKey()
	collectKey := "payment_cycle_collect"
	changed := false
	nextNodes := make([]defNode, 0, len(d.Nodes))
	sendKeys := make([]string, 0, 4)
	hasManual := false
	oldManualKeys := map[string]struct{}{}
	for _, n := range d.Nodes {
		if n.ActionTypeKey == collectType {
			collectKey = firstNonEmpty(strings.TrimSpace(n.ActionKey), collectKey)
			changed = true
			continue
		}
		cfg := rewriteConfigTemplates(n.Config)
		if configTemplatesChanged(n.Config, cfg) {
			n.Config = cfg
			changed = true
		}
		switch n.ActionTypeKey {
		case manualType:
			oldKey := strings.TrimSpace(n.ActionKey)
			if oldKey != "" {
				oldManualKeys[oldKey] = struct{}{}
			}
			if n.ActionKey != manualKey || n.Label != manualLabel || n.Category != "trigger" {
				changed = true
			}
			n.ActionKey = manualKey
			n.Label = manualLabel
			n.Category = "trigger"
			if n.Config == nil {
				n.Config = map[string]any{}
			}
			hasManual = true
		case "telegram.send_message", "max.send_message", "email.send":
			sendKeys = append(sendKeys, strings.TrimSpace(n.ActionKey))
		}
		nextNodes = append(nextNodes, n)
	}
	oldManualKeys[manualKey] = struct{}{}
	if !hasManual {
		nextNodes = append([]defNode{{
			ActionKey:     manualKey,
			Label:         manualLabel,
			ActionTypeKey: manualType,
			Category:      "trigger",
			Config:        map[string]any{},
		}}, nextNodes...)
		changed = true
	}
	if !changed && len(d.Nodes) == len(nextNodes) {
		// still rewire edges if they point at collect
		for _, e := range d.Edges {
			if e.From == collectKey || e.To == collectKey || !hasDefEdge(d.Edges, manualKey, trigger) {
				changed = true
				break
			}
		}
	}
	beforeEdges, _ := json.Marshal(d.Edges)
	d.Nodes = nextNodes
	d.Edges = rebuildEdges(d.Edges, trigger, collectKey, sendKeys, oldManualKeys)
	afterEdges, _ := json.Marshal(d.Edges)
	if string(beforeEdges) != string(afterEdges) {
		changed = true
	}
	return changed
}

func rebuildEdges(existing []defEdge, trigger, collectKey string, sendKeys []string, oldManualKeys map[string]struct{}) []defEdge {
	out := make([]defEdge, 0, len(sendKeys)+len(existing))
	seen := map[string]struct{}{}
	add := func(from, to, label string) {
		if from == "" || to == "" {
			return
		}
		key := from + "->" + to + "|" + label
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, defEdge{From: from, To: to, Label: label})
	}
	add(manualKey, trigger, "")
	for _, e := range existing {
		if e.From == collectKey || e.To == collectKey {
			continue
		}
		if _, ok := oldManualKeys[e.From]; ok {
			continue
		}
		add(e.From, e.To, e.Label)
	}
	for _, send := range sendKeys {
		add(trigger, send, "")
	}
	return out
}

func hasDefEdge(edges []defEdge, from, to string) bool {
	for _, e := range edges {
		if e.From == from && e.To == to {
			return true
		}
	}
	return false
}

func normalizeGraph(graphJSON []byte, def defDoc) (graphDoc, bool) {
	var graph graphDoc
	if len(strings.TrimSpace(string(graphJSON))) > 0 {
		_ = json.Unmarshal(graphJSON, &graph)
	}
	trigger := def.triggerKey()
	collectKey := "payment_cycle_collect"
	changed := false
	nextNodes := make([]graphNode, 0, len(graph.Nodes))
	sendIDs := make([]string, 0, 4)
	hasManual := false
	manualPosition := map[string]any{"x": 80, "y": 160}
	for _, n := range graph.Nodes {
		atk := strings.TrimSpace(n.Data.ActionTypeKey)
		if atk == "" {
			atk = strings.TrimSpace(n.ID)
		}
		if atk == collectType || n.ID == collectKey {
			if n.ID != "" {
				collectKey = n.ID
			}
			changed = true
			continue
		}
		cfg := rewriteConfigTemplates(n.Data.Config)
		if configTemplatesChanged(n.Data.Config, cfg) {
			n.Data.Config = cfg
			changed = true
		}
		switch atk {
		case dueType:
			manualPosition = positionLeftOf(n.Position)
		case manualType:
			if n.ID != manualKey || n.Data.Label != manualLabel || n.Data.ActionTypeKey != manualType {
				changed = true
			}
			n.ID = manualKey
			n.Type = "action"
			n.Data.Label = manualLabel
			n.Data.ActionTypeKey = manualType
			if n.Data.Config == nil {
				n.Data.Config = map[string]any{}
			}
			hasManual = true
		case "telegram.send_message", "max.send_message", "email.send":
			sendIDs = append(sendIDs, n.ID)
		}
		nextNodes = append(nextNodes, n)
	}
	if !hasManual {
		nextNodes = append([]graphNode{{
			ID:       manualKey,
			Type:     "action",
			Position: manualPosition,
			Data: graphNodeData{
				Label:         manualLabel,
				ActionTypeKey: manualType,
				Config:        map[string]any{},
			},
		}}, nextNodes...)
		changed = true
	}
	graph.Nodes = nextNodes
	if len(graph.Edges) == 0 || edgesReference(graph.Edges, collectKey) {
		graph.Edges = graphEdgesFromDef(def, trigger, sendIDs)
		changed = true
	} else {
		filtered := make([]graphEdge, 0, len(graph.Edges))
		for _, e := range graph.Edges {
			if e.Source == collectKey || e.Target == collectKey {
				changed = true
				continue
			}
			if e.Source == manualKey {
				changed = true
				continue
			}
			filtered = append(filtered, e)
		}
		graph.Edges = filtered
		if !hasGraphEdge(graph.Edges, manualKey, trigger) {
			graph.Edges = append(graph.Edges, graphEdge{
				ID:     "e-" + manualKey + "-" + trigger,
				Source: manualKey,
				Target: trigger,
			})
			changed = true
		}
		for _, send := range sendIDs {
			if !hasGraphEdge(graph.Edges, trigger, send) {
				graph.Edges = append(graph.Edges, graphEdge{
					ID:     "e-" + trigger + "-" + send,
					Source: trigger,
					Target: send,
				})
				changed = true
			}
		}
	}
	if len(graph.Nodes) == 0 && len(def.Nodes) > 0 {
		graph.Nodes = graphNodesFromDef(def)
		graph.Edges = graphEdgesFromDef(def, trigger, sendIDs)
		changed = true
	}
	return graph, changed
}

func positionLeftOf(pos map[string]any) map[string]any {
	x := numberFromPosition(pos, "x", 380)
	y := numberFromPosition(pos, "y", 160)
	return map[string]any{"x": x - 300, "y": y}
}

func numberFromPosition(pos map[string]any, key string, fallback float64) float64 {
	if pos == nil {
		return fallback
	}
	switch v := pos[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case json.Number:
		n, err := v.Float64()
		if err == nil {
			return n
		}
	}
	return fallback
}

func edgesReference(edges []graphEdge, collectKey string) bool {
	for _, e := range edges {
		if e.Source == collectKey || e.Target == collectKey {
			return true
		}
	}
	return false
}

func hasGraphEdge(edges []graphEdge, from, to string) bool {
	for _, e := range edges {
		if e.Source == from && e.Target == to {
			return true
		}
	}
	return false
}

func graphEdgesFromDef(def defDoc, trigger string, sendIDs []string) []graphEdge {
	if len(def.Edges) > 0 {
		out := make([]graphEdge, 0, len(def.Edges))
		for i, e := range def.Edges {
			out = append(out, graphEdge{
				ID:     "e-def-" + itoa(i) + "-" + e.From + "-" + e.To,
				Source: e.From,
				Target: e.To,
			})
		}
		return out
	}
	out := make([]graphEdge, 0, len(sendIDs))
	for _, send := range sendIDs {
		out = append(out, graphEdge{ID: "e-" + trigger + "-" + send, Source: trigger, Target: send})
	}
	return out
}

func graphNodesFromDef(def defDoc) []graphNode {
	out := make([]graphNode, 0, len(def.Nodes))
	for i, n := range def.Nodes {
		key := strings.TrimSpace(n.ActionKey)
		if key == "" {
			continue
		}
		out = append(out, graphNode{
			ID:   key,
			Type: "action",
			Position: map[string]any{
				"x": 80 + (i%3)*300,
				"y": 160 + (i/3)*170,
			},
			Data: graphNodeData{
				Label:         n.Label,
				ActionTypeKey: n.ActionTypeKey,
				Config:        n.Config,
			},
		})
	}
	return out
}

func configTemplatesChanged(before, after map[string]any) bool {
	b, _ := json.Marshal(before)
	a, _ := json.Marshal(after)
	return string(b) != string(a)
}

func rewriteConfigTemplates(cfg map[string]any) map[string]any {
	if cfg == nil {
		return cfg
	}
	out := make(map[string]any, len(cfg))
	for k, v := range cfg {
		switch t := v.(type) {
		case string:
			out[k] = strings.ReplaceAll(t, "steps.payment_cycle_collect.", "steps.payment_due.")
		default:
			out[k] = v
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	n := i
	for n > 0 {
		pos--
		b[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(b[pos:])
}
