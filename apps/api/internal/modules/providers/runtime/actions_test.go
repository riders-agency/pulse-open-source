package runtime

import (
	"context"
	neturlpkg "net/url"
	"slices"
	"strings"
	"testing"

	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/registry"
)

func TestRegisterHandlersUsesProviderAdapters(t *testing.T) {
	reg := registry.New()
	if err := RegisterHandlers(reg, &Runner{}); err != nil {
		t.Fatalf("register handlers: %v", err)
	}
	gotDescriptors := reg.ListActions()
	got := make([]string, 0, len(gotDescriptors))
	for _, desc := range gotDescriptors {
		got = append(got, desc.TypeKey)
	}
	want := []string{"ai.apinet.normalize_text", "email.send", "max.send_message", "telegram.answer_callback_query", "telegram.edit_message_text", "telegram.pin_message", "telegram.send_document", "telegram.send_message", "telegram.unpin_message", "uon.create_lead", "yclients.certificate.issue"}
	if !slices.Equal(got, want) {
		t.Fatalf("registered provider handlers = %#v, want %#v", got, want)
	}
	for _, typeKey := range want {
		handler, ok := reg.Action(typeKey)
		if !ok {
			t.Fatalf("handler %s missing", typeKey)
		}
		result := handler.Execute(context.Background(), contracts.ActionRequest{
			ActionTypeKey: typeKey,
			ConfigJSON:    []byte(`{}`),
			RunInputJSON:  []byte(`{}`),
			Steps:         map[string]any{},
		})
		if result.ErrorType != "runtime" || result.Err == nil {
			t.Fatalf("handler %s should fail through runtime availability guard, got error_type=%q err=%v", typeKey, result.ErrorType, result.Err)
		}
	}
}

func TestProviderRuntimeBlocksOutboundEffectsWhenDisabled(t *testing.T) {
	t.Setenv("PULSE_OUTBOUND_EFFECTS_MODE", "disabled")
	reg := registry.New()
	if err := RegisterHandlers(reg, &Runner{}); err != nil {
		t.Fatalf("register handlers: %v", err)
	}
	handler, ok := reg.Action("max.send_message")
	if !ok {
		t.Fatal("max handler missing")
	}
	result := handler.Execute(context.Background(), contracts.ActionRequest{
		ActionKey:     "notify",
		ActionTypeKey: "max.send_message",
		ConfigJSON:    []byte(`{"text":"hello"}`),
		RunInputJSON:  []byte(`{}`),
		Steps:         map[string]any{},
	})
	if result.ErrorType != "outbound_disabled" || result.Err == nil {
		t.Fatalf("expected outbound_disabled, got error_type=%q err=%v", result.ErrorType, result.Err)
	}
}

func TestBuildPlainEmailMessageEncodesUnicodeSubject(t *testing.T) {
	msg := string(buildPlainEmailMessage("sender@example.com", []string{"to@example.com"}, "Заказ 1001", "body"))
	if !strings.Contains(msg, "Subject: =?UTF-8?") {
		t.Fatalf("expected encoded UTF-8 subject, got %q", msg)
	}
}

func TestFormatEmailFromDisplayEncodesUnicodeName(t *testing.T) {
	got, err := formatEmailFromDisplay("Демо-магазин", "certificate@shop.example.test")
	if err != nil {
		t.Fatalf("format sender: %v", err)
	}
	if !strings.Contains(got, "=?utf-8?") || !strings.Contains(got, "<certificate@shop.example.test>") {
		t.Fatalf("expected RFC 2047 display name and sender address, got %q", got)
	}
}

func TestFormatEmailFromDisplayRejectsHeaderInjection(t *testing.T) {
	for _, tc := range []struct {
		name string
		addr string
	}{
		{name: "Демо-магазин\r\nBcc: attacker@example.test", addr: "certificate@shop.example.test"},
		{name: "Демо-магазин", addr: "certificate@shop.example.test\r\nBcc: attacker@example.test"},
	} {
		if _, err := formatEmailFromDisplay(tc.name, tc.addr); err == nil {
			t.Fatalf("expected line break to be rejected: %#v", tc)
		}
	}
}

func TestApplyDependencyExtendedFields(t *testing.T) {
	fields := map[string]any{
		"tour_name":          "Камчатка",
		"region":             "Дальний Восток",
		"tourist_count_text": "2 взр.",
	}
	applyDependencyExtendedFields(fields, map[string]any{
		"extended_field_ids": map[string]any{
			"tour_name":          "193320",
			"region":             "192647",
			"tourist_count_text": "193640",
		},
	})
	extended, ok := fields["extended_fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected extended_fields, got %#v", fields)
	}
	if extended["193320"] != "Камчатка" || extended["192647"] != "Дальний Восток" || extended["193640"] != "2 взр." {
		t.Fatalf("unexpected extended fields: %#v", extended)
	}
	if _, exists := fields["tour_name"]; exists {
		t.Fatalf("logical fields must not leak to U-ON post: %#v", fields)
	}
}

func TestMergeRenderedExtraFieldsPreservesMultilineValues(t *testing.T) {
	runInput := []byte(`{
		"_pulse": {"webhook_slug": "ag-tour-claim"},
		"webhook": {
			"requirements_note": "Строка 1\nСтрока 2 \"в кавычках\"",
			"utm_source": "google"
		}
	}`)
	got := mergeRenderedExtraFields(
		`{"requirements_note":"{{ ag-tour-claim.requirements_note }}","utm_source":"{{ ag-tour-claim.utm_source }}"}`,
		runInput,
		nil,
		templateSystem{},
	)
	if got["requirements_note"] != "Строка 1\nСтрока 2 \"в кавычках\"" || got["utm_source"] != "google" {
		t.Fatalf("unexpected rendered extra fields: %#v", got)
	}
}

func TestApplyDependencyLeadDefaults(t *testing.T) {
	fields := map[string]any{
		"source": "example.test / заявка на тур",
		"r_u_id": "42",
	}
	applyDependencyLeadDefaults(fields, map[string]any{
		"default_manager_id": "7",
		"default_office_id":  "1",
		"default_lead_fields": map[string]any{
			"status_id": "3",
		},
	})
	if fields["r_u_id"] != "42" {
		t.Fatalf("explicit r_u_id must not be overwritten: %#v", fields)
	}
	if fields["r_co_id"] != "1" || fields["status_id"] != "3" {
		t.Fatalf("expected lead defaults, got %#v", fields)
	}
}

func TestNormalizeUONLeadDateFields(t *testing.T) {
	fields := map[string]any{
		"date_from": "01.12.2026",
		"date_to":   "2026-12-10",
	}
	normalizeUONLeadDateFields(fields)
	if fields["date_from"] != "2026-12-01" {
		t.Fatalf("date_from = %q, want 2026-12-01", fields["date_from"])
	}
	if fields["date_to"] != "2026-12-10" {
		t.Fatalf("date_to = %q, want unchanged ISO date", fields["date_to"])
	}
}

func TestNormalizeUONLeadDateFieldsRussianTourRanges(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		from string
		to   string
	}{
		{"same month with price", "8 — 19 декабря 2026 г. — 11 950 $/чел.", "2026-12-08", "2026-12-19"},
		{"different months", "27 января — 6 февраля 2027 г. — 10 800 $/чел.", "2027-01-27", "2027-02-06"},
		{"different years", "28 декабря 2026 — 8 января 2027 г.", "2026-12-28", "2027-01-08"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := map[string]any{"date_from": tt.raw}
			normalizeUONLeadDateFields(fields)
			if fields["date_from"] != tt.from || fields["date_to"] != tt.to {
				t.Fatalf("dates = %q, %q; want %q, %q", fields["date_from"], fields["date_to"], tt.from, tt.to)
			}
		})
	}
}

func TestNormalizeUONLeadDateFieldsDropsUnknownDate(t *testing.T) {
	fields := map[string]any{"date_from": "Дата 1 — 11 950 $/чел.", "note": "Дата 1 — 11 950 $/чел."}
	normalizeUONLeadDateFields(fields)
	if _, ok := fields["date_from"]; ok {
		t.Fatalf("unrecognized date was sent to U-ON: %#v", fields)
	}
	if fields["note"] == "" {
		t.Fatal("original date must remain available in the note")
	}
}

func TestMaxAuthorizationHeaderUsesRawToken(t *testing.T) {
	if got := maxAuthorizationHeader("Bearer secret-token", map[string]any{"authorization_scheme": "bearer"}); got != "secret-token" {
		t.Fatalf("Authorization = %q, want raw token", got)
	}
	if got := maxAuthorizationHeader(" secret-token ", nil); got != "secret-token" {
		t.Fatalf("Authorization = %q, want trimmed raw token", got)
	}
}

func TestResolveMaxRecipientSupportsChatOrUserID(t *testing.T) {
	recipient, err := resolveMaxRecipient(actionConfig{ChatID: "https://web.max.ru/chat/-74739869656479?x=1"}, nil, nil, nil, templateSystem{})
	if err != nil {
		t.Fatalf("resolve chat_id: %v", err)
	}
	if recipient.Param != "chat_id" || recipient.Value != "-74739869656479" {
		t.Fatalf("unexpected chat recipient: %#v", recipient)
	}

	recipient, err = resolveMaxRecipient(actionConfig{UserID: "{{ user.id }}"}, nil, []byte(`{"user":{"id":12345}}`), nil, templateSystem{})
	if err != nil {
		t.Fatalf("resolve user_id: %v", err)
	}
	if recipient.Param != "user_id" || recipient.Value != "12345" {
		t.Fatalf("unexpected user recipient: %#v", recipient)
	}

	if _, err := resolveMaxRecipient(actionConfig{ChatID: "1", UserID: "2"}, nil, nil, nil, templateSystem{}); err == nil {
		t.Fatalf("expected both-recipient validation error")
	}
}

func TestApplyMaxBodyAndQueryOptions(t *testing.T) {
	notify := false
	disablePreview := false
	body := map[string]any{"text": "hello"}
	err := applyMaxBodyOptions(actionConfig{Notify: &notify, Format: "HTML"}, map[string]any{"default_notify": true}, body)
	if err != nil {
		t.Fatalf("apply body options: %v", err)
	}
	if body["notify"] != false || body["format"] != "html" {
		t.Fatalf("unexpected body options: %#v", body)
	}

	q := neturlpkg.Values{}
	applyMaxQueryOptions(actionConfig{DisableLinkPreview: &disablePreview}, map[string]any{"default_disable_link_preview": true}, q)
	if q.Get("disable_link_preview") != "false" {
		t.Fatalf("disable_link_preview = %q, want false", q.Get("disable_link_preview"))
	}

	if err := applyMaxBodyOptions(actionConfig{Format: "MarkdownV2"}, nil, map[string]any{}); err == nil {
		t.Fatalf("expected invalid format error")
	}
}
