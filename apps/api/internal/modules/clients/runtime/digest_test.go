package runtime

import (
	"strings"
	"testing"
	"time"
)

func TestBuildPaymentDigest(t *testing.T) {
	rows := []paymentCycleRow{
		{
			CycleState:      "scheduled",
			StateLabel:      "Запланирован",
			PaymentActionID: "11111111-1111-4111-8111-111111111101",
			ClientName:      "ООО Альфа",
			TotalAmount:     50000,
			DueDate:         "2026-06-10",
			ServiceCount:    2,
		},
		{
			CycleState:      "sent",
			StateLabel:      "Отправлен",
			PaymentActionID: "11111111-1111-4111-8111-111111111102",
			ClientName:      "Бета",
			TotalAmount:     12000,
			DueDate:         "2026-06-08",
			ServiceCount:    1,
		},
	}
	d := buildPaymentDigest(rows, "KSS", "11111111-1111-4111-8111-111111111100", paymentDigestEvent{})
	if len(splitTelegramHTML(d.TelegramHTML, 3900)) < 1 {
		t.Fatal("expected telegram digest chunks")
	}
	if d.Subject == "" || !strings.Contains(d.TelegramHTML, "ООО Альфа") {
		t.Fatalf("telegram digest missing client: %q", d.TelegramHTML)
	}
	if !strings.Contains(d.EmailHTML, "50 000") {
		t.Fatalf("email digest missing amount: %q", d.EmailHTML)
	}
	if !strings.Contains(d.MaxText, "2 клиента") {
		t.Fatalf("max digest missing summary: %q", d.MaxText)
	}
}

func TestBuildPaymentDigestKeepsProjectPictureWithPaymentEvent(t *testing.T) {
	oldNow := paymentDigestNow
	paymentDigestNow = func() time.Time { return time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC) }
	defer func() { paymentDigestNow = oldNow }()

	rows := []paymentCycleRow{
		{
			CycleState:      "scheduled",
			StateLabel:      "Запланирован",
			PaymentActionID: "11111111-1111-4111-8111-111111111101",
			ClientName:      "Орион",
			TotalAmount:     1000000,
			DueDate:         "2026-06-01",
			PeriodStart:     "2026-05-01",
			PeriodEnd:       "2026-05-31",
			ServiceCount:    1,
			ServiceSummary:  "техподдержка",
		},
		{
			CycleState:      "sent",
			StateLabel:      "Отправлен",
			PaymentActionID: "11111111-1111-4111-8111-111111111102",
			ClientName:      "Другой клиент",
			TotalAmount:     70000,
			DueDate:         "2026-06-05",
			ServiceCount:    1,
			ServiceSummary:  "хостинг",
		},
	}
	d := buildPaymentDigest(rows, "Riders", "11111111-1111-4111-8111-111111111100", paymentDigestEvent{
		Present:         true,
		ClientName:      "Орион",
		PaymentActionID: "11111111-1111-4111-8111-111111111101",
		ServiceSummary:  "техподдержка",
		Source:          "manual",
		PaymentMethod:   "legal_invoice",
		Amount:          1000000,
		PeriodStart:     "2026-05-01",
		PeriodEnd:       "2026-05-31",
		NextPeriodStart: "2026-06-01",
	})
	for _, want := range []string{"Счет оплачен", "Источник события: ручное внесение", "Способ оплаты: по счету с юрлица", "Другой клиент"} {
		if !strings.Contains(d.PlainText, want) && !strings.Contains(d.EmailHTML, want) {
			t.Fatalf("digest missing %q\nplain=%s\nhtml=%s", want, d.PlainText, d.EmailHTML)
		}
	}
	if strings.Index(d.PlainText, "Счет оплачен") > strings.Index(d.PlainText, "Общая картина по оплатам") {
		t.Fatalf("current payment must be before project snapshot:\n%s", d.PlainText)
	}
	if !strings.Contains(d.Subject, "Счет оплачен: Орион") {
		t.Fatalf("subject should include payment event: %q", d.Subject)
	}
}

func TestBuildPaymentDigestGroupsAndCoalescesRows(t *testing.T) {
	oldNow := paymentDigestNow
	paymentDigestNow = func() time.Time { return time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC) }
	defer func() { paymentDigestNow = oldNow }()

	rows := []paymentCycleRow{
		{
			CycleState:     "scheduled",
			StateLabel:     "Запланирован",
			ClientID:       "client-odintsovo",
			ClientName:     "Одинцово",
			TotalAmount:    10000,
			DueDate:        "2026-06-01",
			ServiceCount:   1,
			ServiceSummary: "поддержка",
		},
		{
			CycleState:     "scheduled",
			StateLabel:     "Запланирован",
			ClientID:       "client-odintsovo",
			ClientName:     "Одинцово",
			TotalAmount:    5000,
			DueDate:        "2026-06-02",
			ServiceCount:   1,
			ServiceSummary: "хостинг",
		},
		{
			CycleState:   "sent",
			StateLabel:   "Отправлен",
			ClientID:     "client-beta",
			ClientName:   "Бета",
			TotalAmount:  7000,
			DueDate:      "2026-06-10",
			ServiceCount: 1,
		},
	}
	d := buildPaymentDigest(rows, "Rides Agency", "11111111-1111-4111-8111-111111111100", paymentDigestEvent{})
	for _, want := range []string{"Ждут счета", "Просрочено", "Одинцово — 15 000 ₽", "поддержка, хостинг", "Ждем денег"} {
		if !strings.Contains(d.PlainText, want) {
			t.Fatalf("plain digest missing %q:\n%s", want, d.PlainText)
		}
	}
	if strings.Count(d.PlainText, "Одинцово") != 1 {
		t.Fatalf("expected client rows to be coalesced:\n%s", d.PlainText)
	}
	if !strings.Contains(d.TelegramHTML, "<b>🧾 Ждут счета") || !strings.Contains(d.TelegramHTML, "<b>💰 Ждем денег") {
		t.Fatalf("telegram digest should include grouped heading:\n%s", d.TelegramHTML)
	}
}

func TestBuildPaymentDigestSkipsCurrentBlockWhenPreRendered(t *testing.T) {
	d := buildPaymentDigest(nil, "Riders", "11111111-1111-4111-8111-111111111100", paymentDigestEvent{
		Present:     true,
		PreRendered: true,
		ClientName:  "Орион",
		Amount:      1000000,
		Source:      "manual",
	})
	if strings.Contains(d.TelegramHTML, "Счет оплачен") {
		t.Fatalf("digest must not duplicate pre-rendered payment block:\n%s", d.TelegramHTML)
	}
	if !strings.Contains(d.TelegramHTML, "Общая картина по оплатам") {
		t.Fatalf("digest should still render project snapshot:\n%s", d.TelegramHTML)
	}
}

func TestPaymentDigestEventFromRunInputSources(t *testing.T) {
	event := paymentDigestEventFromRunInput([]byte(`{
		"_pulse": {"notification_current_payment_pre_rendered": true},
		"current_payment": {
			"payment_action_id": "pa-1",
			"source": "email",
			"payment_method": "cash",
			"amount": 1200,
			"period_start": "2026-05-01",
			"period_end": "2026-05-31",
			"next_period_start": "2026-06-01",
			"client_name": "Орион",
			"links": {
				"client_link": {"label": "Орион", "href": "/clients/client-1"},
				"payment_link": {"label": "Оплата", "href": "/clients/client-1?tab=payments&payment_id=pf-1"}
			}
		}
	}`))
	if !event.Present {
		t.Fatal("expected event")
	}
	if !event.PreRendered {
		t.Fatal("expected pre-rendered notification flag")
	}
	if event.SourceLabel != "входящее письмо" {
		t.Fatalf("unexpected source label: %q", event.SourceLabel)
	}
	if event.PaymentMethodLabel != "" {
		t.Fatalf("run input has no explicit method label, got %q", event.PaymentMethodLabel)
	}
	if paymentMethodLabel(event.PaymentMethod) != "наличные" {
		t.Fatalf("unexpected payment method label: %q", paymentMethodLabel(event.PaymentMethod))
	}
	if event.NextPeriodStart != "2026-06-01" {
		t.Fatalf("unexpected next period: %q", event.NextPeriodStart)
	}
	if !strings.Contains(event.ClientURL, "/clients/client-1") {
		t.Fatalf("expected client link to become absolute url: %q", event.ClientURL)
	}
}
