package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
	uonprovider "riders.agency/pulse/api/internal/modules/providers/uon"
)

func (r *Runner) executeUonCreateLead(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	tpl := templateSystemForRun(runID)
	dep, err := r.loadDependency(ctx, node.Config.ActionDependencyID)
	if err != nil {
		return failedAction(input, "{}", map[string]any{"provider": "uon", "ok": false, "error": err.Error()}, false, "dependency", err)
	}
	if dep.TypeKey != "uon.account" {
		err := fmt.Errorf("uon.create_lead requires uon.account dependency")
		return failedAction(input, "{}", map[string]any{"provider": "uon", "ok": false, "error": err.Error()}, false, "dependency_type", err)
	}
	if dep.Status != "active" {
		err := fmt.Errorf("uon dependency is %s", dep.Status)
		return failedAction(input, "{}", map[string]any{"provider": "uon", "ok": false, "status": dep.Status}, false, "dependency_status", err)
	}
	apiKey := strings.TrimSpace(dep.Plaintext)
	if apiKey == "" {
		err := errors.New("u-on api_key is empty")
		return failedAction(input, "{}", map[string]any{"provider": "uon", "ok": false, "error": err.Error()}, false, "dependency", err)
	}

	channel := strings.TrimSpace(renderTemplate(firstNonEmpty(node.Config.Channel, stringMetadata(dep.Metadata, "default_channel")), runInput, steps, tpl))
	note := strings.TrimSpace(renderTemplate(node.Config.Note, runInput, steps, tpl))
	if note == "" {
		note = strings.TrimSpace(renderTemplate(node.Config.Text, runInput, steps, tpl))
	}
	fields := map[string]any{
		"source":    uonprovider.BuildSource(stringMetadata(dep.Metadata, "lead_source_prefix"), channel),
		"note":      note,
		"u_name":    strings.TrimSpace(renderTemplate(node.Config.UName, runInput, steps, tpl)),
		"u_surname": strings.TrimSpace(renderTemplate(node.Config.USurname, runInput, steps, tpl)),
		"u_sname":   strings.TrimSpace(renderTemplate(node.Config.USname, runInput, steps, tpl)),
		"u_phone":   strings.TrimSpace(renderTemplate(firstNonEmpty(node.Config.UPhone, node.Config.Phone), runInput, steps, tpl)),
		"u_email":   strings.TrimSpace(renderTemplate(node.Config.UEmail, runInput, steps, tpl)),
	}
	applyDependencyLeadDefaults(fields, dep.Metadata)
	if extra := mergeRenderedExtraFields(node.Config.ExtraFieldsJSON, runInput, steps, tpl); len(extra) > 0 {
		for key, value := range extra {
			fields[key] = value
		}
	}
	normalizeUONLeadDateFields(fields)
	applyDependencyExtendedFields(fields, dep.Metadata)
	if note == "" {
		err := errors.New("u-on note is required")
		return failedAction(input, "{}", map[string]any{"provider": "uon", "ok": false, "error": err.Error()}, false, "validation", err)
	}
	if !uonprovider.HasContact(fields) && !boolMetadata(dep.Metadata, "allow_note_only") {
		err := errors.New("u-on requires phone or valid email; enable allow_note_only in integration metadata for visitor-only leads")
		return failedAction(input, "{}", map[string]any{"provider": "uon", "ok": false, "error": err.Error()}, false, "validation", err)
	}

	requestInput, _ := json.Marshal(map[string]any{
		"source":           fields["source"],
		"note":             fields["note"],
		"u_name":           fields["u_name"],
		"u_surname":        fields["u_surname"],
		"u_sname":          fields["u_sname"],
		"u_phone":          fields["u_phone"],
		"u_email":          fields["u_email"],
		"success_criteria": node.Config.SuccessCriteria,
	})
	timeout := durationOrDefault(node.Config.RetryPolicy.Timeout, 15*time.Second)
	apiBase := stringMetadata(dep.Metadata, "api_base")
	result, err := uonprovider.LeadCreate(ctx, r.httpClient(), apiKey, apiBase, fields, timeout)
	provider := uonprovider.SanitizeResponse(result.HTTP, []byte(result.Raw))
	if err != nil {
		providerJSON, _ := json.Marshal(provider)
		retryable := result.HTTP >= 500 || result.HTTP == 0
		report := diagnosticsForFailure(node.ActionTypeKey, provider, retryable, "provider_http", err)
		output := providerdiagnostics.OutputWithReport("{}", report)
		return actionExecution{InputJSON: string(requestInput), OutputJSON: output, ProviderResponseJSON: string(providerJSON), Retryable: retryable, ErrorType: "provider_http", Err: err}
	}

	delivered := map[string]any{
		"delivered":   true,
		"provider":    "uon",
		"provider_ok": true,
		"result":      result.ResultCode,
		"source":      fields["source"],
	}
	if leadID := strings.TrimSpace(fmt.Sprint(provider["id"])); leadID != "" && leadID != "<nil>" {
		delivered["lead_id"] = leadID
	}
	if warnings, ok := provider["warnings"].([]string); ok && len(warnings) > 0 {
		delivered["warnings"] = warnings
	}
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(requestInput), provider, delivered, uonExpose(fields))
}

func mergeRenderedExtraFields(raw string, runInput []byte, steps map[string]any, tpl templateSystem) map[string]any {
	var templates map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &templates); err == nil {
		out := make(map[string]any, len(templates))
		for key, value := range templates {
			if template, ok := value.(string); ok {
				rendered := strings.TrimSpace(renderTemplate(template, runInput, steps, tpl))
				if rendered != "" {
					out[key] = rendered
				}
				continue
			}
			if value != nil {
				out[key] = value
			}
		}
		return out
	}

	rendered := strings.TrimSpace(renderTemplate(raw, runInput, steps, tpl))
	if rendered == "" || rendered == "{}" {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(rendered), &out); err != nil || len(out) == 0 {
		return nil
	}
	return out
}

func uonExpose(fields map[string]any) map[string]any {
	return map[string]any{
		"source":    fields["source"],
		"note":      fields["note"],
		"u_name":    fields["u_name"],
		"u_surname": fields["u_surname"],
		"u_sname":   fields["u_sname"],
		"u_phone":   fields["u_phone"],
		"u_email":   fields["u_email"],
		"r_u_id":    fields["r_u_id"],
		"r_ci_id":   fields["r_ci_id"],
		"r_co_id":   fields["r_co_id"],
	}
}

func applyDependencyLeadDefaults(fields map[string]any, metadata map[string]any) {
	if fields == nil || metadata == nil {
		return
	}
	if defaults := metadataAnyMap(metadata, "default_lead_fields"); len(defaults) > 0 {
		for key, value := range defaults {
			setUONDefaultField(fields, key, value)
		}
	}
	for metaKey, uonKey := range map[string]string{
		"default_manager_id": "r_u_id",
		"default_company_id": "r_ci_id",
		"default_office_id":  "r_co_id",
		"default_status_id":  "status_id",
	} {
		setUONDefaultField(fields, uonKey, metadata[metaKey])
	}
}

func setUONDefaultField(fields map[string]any, key string, value any) {
	key = strings.TrimSpace(key)
	if key == "" || value == nil {
		return
	}
	if existing, ok := fields[key]; ok && strings.TrimSpace(fmt.Sprint(existing)) != "" {
		return
	}
	if s := strings.TrimSpace(fmt.Sprint(value)); s != "" && s != "<nil>" {
		fields[key] = s
	}
}

func applyDependencyExtendedFields(fields map[string]any, metadata map[string]any) {
	ids := metadataStringMap(metadata, "extended_field_ids")
	if len(ids) == 0 {
		return
	}
	extended := map[string]any{}
	if existing, ok := fields["extended_fields"].(map[string]any); ok {
		for key, value := range existing {
			extended[key] = value
		}
	}
	for logical, fieldID := range ids {
		fieldID = strings.TrimSpace(fieldID)
		if fieldID == "" {
			continue
		}
		value, exists := fields[logical]
		if !exists || strings.TrimSpace(fmt.Sprint(value)) == "" {
			continue
		}
		extended[fieldID] = value
		delete(fields, logical)
	}
	if len(extended) > 0 {
		fields["extended_fields"] = extended
	}
}

func metadataStringMap(metadata map[string]any, key string) map[string]string {
	raw, ok := metadata[key]
	if !ok || raw == nil {
		return nil
	}
	out := map[string]string{}
	switch typed := raw.(type) {
	case map[string]any:
		for k, v := range typed {
			if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
				out[strings.TrimSpace(k)] = s
			}
		}
	case map[string]string:
		for k, v := range typed {
			if s := strings.TrimSpace(v); s != "" {
				out[strings.TrimSpace(k)] = s
			}
		}
	}
	return out
}

func metadataAnyMap(metadata map[string]any, key string) map[string]any {
	raw, ok := metadata[key]
	if !ok || raw == nil {
		return nil
	}
	switch typed := raw.(type) {
	case map[string]any:
		return typed
	case map[string]string:
		out := map[string]any{}
		for k, v := range typed {
			out[k] = v
		}
		return out
	default:
		return nil
	}
}

func boolMetadata(metadata map[string]any, key string) bool {
	if metadata == nil {
		return false
	}
	v, ok := metadata[key].(bool)
	return ok && v
}
