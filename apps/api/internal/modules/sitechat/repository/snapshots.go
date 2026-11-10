package repository

import (
	"strings"
	"time"
)

func normalizePresence(in PresenceInput) PresenceInput {
	in.PublicKey = strings.TrimSpace(in.PublicKey)
	in.URL = trimLimit(in.URL, 2048)
	in.Title = trimLimit(in.Title, 512)
	in.Referrer = trimLimit(in.Referrer, 2048)
	in.Browser = trimLimit(in.Browser, 80)
	in.BrowserVersion = trimLimit(in.BrowserVersion, 40)
	in.OS = trimLimit(in.OS, 80)
	in.DeviceType = trimLimit(in.DeviceType, 40)
	if in.DeviceType == "" {
		in.DeviceType = "unknown"
	}
	if in.DurationSeconds < 0 {
		in.DurationSeconds = 0
	}
	if in.DurationSeconds > 24*60*60 {
		in.DurationSeconds = 24 * 60 * 60
	}
	in.Language = trimLimit(in.Language, 40)
	in.Timezone = trimLimit(in.Timezone, 80)
	if in.UTM == nil {
		in.UTM = map[string]any{}
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	return in
}

func buildSnapshot(visitorID, sessionID, widgetID string, in PresenceInput, first, invited bool, chatOpen bool, monitoring string) map[string]any {
	snapshot := map[string]any{
		"visitor_id":       visitorID,
		"visit_session_id": sessionID,
		"chat_widget_id":   widgetID,
		"current_url":      in.URL,
		"current_title":    in.Title,
		"viewport": map[string]any{
			"width":  in.ViewportWidth,
			"height": in.ViewportHeight,
		},
		"scroll":    map[string]any{"x": in.ScrollX, "y": in.ScrollY},
		"analytics": analyticsPayload(in),
		"activity": map[string]any{
			"last_seen_at":       time.Now().UTC().Format(time.RFC3339),
			"dwell_seconds":      in.DurationSeconds,
			"clicks_last_minute": 0,
			"is_active":          true,
			"connection_status":  "online",
			"call_status":        "",
			"call_session_id":    "",
		},
		"signals": map[string]any{
			"is_first_visit":    first,
			"has_rage_click":    false,
			"has_dead_click":    false,
			"is_high_intent":    false,
			"chat_open":         chatOpen,
			"operator_invited":  invited,
			"monitoring_active": monitoring == "active",
		},
		"monitoring_status": monitoring,
		"suggested_actions": []string{"offer_help", "start_monitoring"},
	}
	if len(in.ViewportContext) > 0 {
		snapshot["viewport_context"] = in.ViewportContext
	}
	return snapshot
}

func analyticsPayload(in PresenceInput) map[string]any {
	return map[string]any{
		"referrer":    in.Referrer,
		"utm":         in.UTM,
		"browser":     in.Browser,
		"os":          in.OS,
		"device_type": in.DeviceType,
		"language":    in.Language,
		"timezone":    in.Timezone,
		"screen": map[string]any{
			"width":  in.ScreenWidth,
			"height": in.ScreenHeight,
		},
		"metrika": map[string]any{
			"client_id":  in.MetrikaClientID,
			"user_id":    in.MetrikaUserID,
			"session_id": in.MetrikaSessionID,
		},
		"metadata": presenceMetadata(in.Metadata),
	}
}

func presenceMetadata(in map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{"page_session_id", "widget_instance_id"} {
		if value := presenceMetadataValue(in, key); value != "" {
			out[key] = value
		}
	}
	return out
}

func presenceMetadataValue(in map[string]any, key string) string {
	if in == nil {
		return ""
	}
	value, ok := in[key].(string)
	if !ok {
		return ""
	}
	return trimLimit(value, 120)
}

func applyEvents(snapshot map[string]any, events []EventInput) {
	if snapshot == nil {
		return
	}
	activity := ensureMap(snapshot, "activity")
	signals := ensureMap(snapshot, "signals")
	lastSequence := int64(0)
	if seq, ok := explicitSequenceValue(activity["last_sequence"]); ok {
		lastSequence = seq
	}
	maxSequence := lastSequence
	clicks := 0
	visitorLeft := false
	for _, ev := range events {
		if seq, ok := explicitSequenceValue(ev.Payload["sequence"]); ok {
			if seq <= lastSequence && ev.EventType != "visitor.left" {
				continue
			}
			if seq > maxSequence {
				maxSequence = seq
			}
		}
		switch ev.EventType {
		case "page", "page.navigate", "viewport", "scroll", "monitoring.snapshot", "monitoring.incremental":
			if u, ok := ev.Payload["url"].(string); ok && u != "" {
				snapshot["current_url"] = u
			}
			if t, ok := ev.Payload["title"].(string); ok && t != "" {
				snapshot["current_title"] = t
			}
			if w, ok := numberValue(ev.Payload["viewport_width"]); ok {
				snapshot["viewport"] = map[string]any{"width": w, "height": numberOrZero(ev.Payload["viewport_height"])}
			}
			if w, ok := numberValue(ev.Payload["width"]); ok {
				snapshot["viewport"] = map[string]any{"width": w, "height": numberOrZero(ev.Payload["height"])}
			}
			documentWidth, hasDocumentWidth := numberValue(ev.Payload["document_width"])
			documentHeight := numberOrZero(ev.Payload["document_height"])
			if hasDocumentWidth {
				snapshot["document"] = map[string]any{"width": documentWidth, "height": documentHeight}
			}
			if y, ok := numberValue(ev.Payload["scroll_y"]); ok {
				scroll := map[string]any{"x": numberOrZero(ev.Payload["scroll_x"]), "y": y}
				if hasDocumentWidth {
					scroll["document_width"] = documentWidth
					scroll["document_height"] = documentHeight
				}
				snapshot["scroll"] = scroll
			}
			if y, ok := numberValue(ev.Payload["y"]); ok && ev.EventType == "scroll" {
				scroll := map[string]any{"x": numberOrZero(ev.Payload["x"]), "y": y}
				if hasDocumentWidth {
					scroll["document_width"] = documentWidth
					scroll["document_height"] = documentHeight
				}
				snapshot["scroll"] = scroll
			}
			if mirror, ok := ev.Payload["mirror"].(map[string]any); ok && len(mirror) > 0 {
				snapshot["mirror"] = mirror
			} else if containers, ok := ev.Payload["scroll_containers"].([]any); ok {
				mirror, _ := snapshot["mirror"].(map[string]any)
				if mirror == nil {
					mirror = map[string]any{}
				}
				mirror["scroll_containers"] = containers
				snapshot["mirror"] = mirror
			}
			if context, ok := ev.Payload["viewport_context"].(map[string]any); ok && len(context) > 0 {
				snapshot["viewport_context"] = context
			}
		case "click":
			clicks++
			if x, ok := numberValue(ev.Payload["x"]); ok {
				snapshot["last_click"] = map[string]any{"x": x, "y": numberOrZero(ev.Payload["y"])}
			}
		case "cursor", "cursor.move":
			if x, ok := numberValue(ev.Payload["x"]); ok {
				snapshot["cursor"] = map[string]any{"x": x, "y": numberOrZero(ev.Payload["y"])}
			}
		case "touch", "touch.move":
			if x, ok := numberValue(ev.Payload["x"]); ok {
				snapshot["touch"] = map[string]any{"x": x, "y": numberOrZero(ev.Payload["y"])}
			}
		case "rage_click":
			signals["has_rage_click"] = true
		case "dead_click":
			signals["has_dead_click"] = true
		case "chat_open":
			signals["chat_open"] = true
		case "chat.closed":
			signals["chat_open"] = false
		case "widget_state":
			snapshot["widget_state"] = ev.Payload
			if open, ok := ev.Payload["open"].(bool); ok {
				signals["chat_open"] = open
			}
		case "visitor.left":
			visitorLeft = true
		}
	}
	activity["last_seen_at"] = time.Now().UTC().Format(time.RFC3339)
	if visitorLeft {
		activity["is_active"] = false
		activity["connection_status"] = "offline"
	} else {
		activity["is_active"] = true
		activity["connection_status"] = "online"
	}
	if clicks > 0 {
		activity["clicks_last_minute"] = clicks
	}
	if maxSequence > 0 {
		activity["last_sequence"] = maxSequence
	}
}

func mergePresenceSnapshot(existing, fresh map[string]any) map[string]any {
	if len(existing) == 0 {
		return fresh
	}
	if fresh == nil {
		return existing
	}
	for _, key := range []string{"visitor_id", "visit_session_id", "chat_widget_id", "current_url", "current_title", "viewport", "scroll", "analytics", "monitoring_status"} {
		if value, ok := fresh[key]; ok {
			existing[key] = value
		}
	}
	if value, ok := fresh["viewport_context"]; ok && value != nil {
		existing["viewport_context"] = value
	}
	activity := ensureMap(existing, "activity")
	freshActivity, _ := fresh["activity"].(map[string]any)
	for _, key := range []string{"last_seen_at", "is_active", "connection_status", "dwell_seconds"} {
		if value, ok := freshActivity[key]; ok {
			activity[key] = value
		}
	}
	signals := ensureMap(existing, "signals")
	freshSignals, _ := fresh["signals"].(map[string]any)
	for _, key := range []string{"is_first_visit", "chat_open", "operator_invited", "monitoring_active"} {
		if value, ok := freshSignals[key]; ok {
			signals[key] = value
		}
	}
	if _, ok := existing["suggested_actions"]; !ok {
		existing["suggested_actions"] = fresh["suggested_actions"]
	}
	return existing
}

func ensureMap(parent map[string]any, key string) map[string]any {
	value, _ := parent[key].(map[string]any)
	if value == nil {
		value = map[string]any{}
		parent[key] = value
	}
	return value
}

func eventCounts(events []EventInput) map[string]any {
	out := map[string]any{}
	for _, ev := range events {
		k := strings.TrimSpace(ev.EventType)
		if k == "" {
			continue
		}
		if n, ok := out[k].(int); ok {
			out[k] = n + 1
		} else {
			out[k] = 1
		}
	}
	return out
}
