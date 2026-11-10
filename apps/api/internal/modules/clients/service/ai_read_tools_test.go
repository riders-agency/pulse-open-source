package service

import (
	"strings"
	"testing"

	clientsrepo "riders.agency/pulse/api/internal/modules/clients/repository"
)

func TestClientQueryFromTextExtractsSemanticClientName(t *testing.T) {
	tests := map[string]string{
		"Найди клиента Орион":                         "Орион",
		"Когда оплата у Орион?":                       "Орион",
		"Создай задачу проверить оплату Орион завтра": "Орион",
		"Орион": "Орион",
	}
	for input, want := range tests {
		if got := clientQueryFromText(input); got != want {
			t.Fatalf("clientQueryFromText(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestClientSearchTokensDropCommandWords(t *testing.T) {
	got := clientSearchTokens("Создай задачу проверить оплату Орион завтра")
	if len(got) != 1 || got[0] != "орион" {
		t.Fatalf("clientSearchTokens returned %#v, want [орион]", got)
	}
}

func TestProjectPaymentActionsAnswerListsNearestActivePayments(t *testing.T) {
	answer, payload := projectPaymentActionsAnswer([]clientsrepo.PaymentActionRow{
		{
			PaymentActionID: "payment-1",
			ClientID:        "client-1",
			ClientName:      "Орион",
			ProjectID:       "project-1",
			DueDate:         "2026-06-30",
			TotalAmount:     1000,
			Currency:        "RUB",
			Status:          "pending",
		},
		{
			PaymentActionID: "payment-2",
			ClientID:        "client-2",
			ClientName:      "KSS",
			ProjectID:       "project-1",
			DueDate:         "2026-07-01",
			TotalAmount:     2000,
			Currency:        "RUB",
			Status:          "sent",
		},
	})

	if !strings.Contains(answer, "Ближайшие оплаты:\n1. Орион — 30.06.2026, 1 000 RUB") ||
		!strings.Contains(answer, "\n2. KSS — 01.07.2026, 2 000 RUB") ||
		strings.Contains(answer, "статус") {
		t.Fatalf("answer = %q", answer)
	}
	items, ok := payload["items"].([]map[string]any)
	if !ok || len(items) != 2 {
		t.Fatalf("items = %#v", payload["items"])
	}
	if items[0]["client_name"] != "Орион" || items[0]["due_date"] != "2026-06-30" {
		t.Fatalf("first item = %#v", items[0])
	}
}

func TestPaymentStatusActiveFilter(t *testing.T) {
	for _, status := range []string{"pending", "sent", "snoozed", "needs_decision"} {
		if !isActivePaymentStatus(status) {
			t.Fatalf("status %q should be active", status)
		}
	}
	for _, status := range []string{"paid", "cancelled", "canceled", "completed"} {
		if isActivePaymentStatus(status) {
			t.Fatalf("status %q should not be active", status)
		}
	}
}
