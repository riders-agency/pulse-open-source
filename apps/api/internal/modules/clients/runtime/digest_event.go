package runtime

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"
)

func paymentEventSubject(event paymentDigestEvent) string {
	client := strings.TrimSpace(event.ClientName)
	if client == "" {
		client = "Клиент"
	}
	service := firstNonEmpty(strings.TrimSpace(event.ServiceName), strings.TrimSpace(event.ServiceSummary), "услуги")
	return fmt.Sprintf("Счет оплачен: %s · %s · %s", client, formatMoneyRUB(event.Amount), service)
}

func paymentEventTelegram(event paymentDigestEvent) string {
	client := firstNonEmpty(strings.TrimSpace(event.ClientName), "Клиент")
	service := firstNonEmpty(strings.TrimSpace(event.ServiceName), strings.TrimSpace(event.ServiceSummary), "услуги")
	period := paymentPeriodDisplay(event)
	next := firstNonEmpty(formatDueDateRU(firstNonEmpty(event.NextPaymentDueDate, event.NextPeriodStart)), "—")
	source := firstNonEmpty(strings.TrimSpace(event.SourceLabel), paymentSourceLabel(event.Source))
	method := firstNonEmpty(strings.TrimSpace(event.PaymentMethodLabel), paymentMethodLabel(event.PaymentMethod))
	var b strings.Builder
	fmt.Fprintf(&b, "<b>💰 Счет оплачен</b>\n\n")
	fmt.Fprintf(&b, "<b>%s</b> · <b>%s</b>\n", html.EscapeString(client), html.EscapeString(formatMoneyRUB(event.Amount)))
	fmt.Fprintf(&b, "Услуга: %s\n", html.EscapeString(service))
	fmt.Fprintf(&b, "Период: %s\n", html.EscapeString(firstNonEmpty(period, "—")))
	fmt.Fprintf(&b, "Следующий платеж: %s\n", html.EscapeString(next))
	fmt.Fprintf(&b, "Способ: %s\n", html.EscapeString(method))
	fmt.Fprintf(&b, "Источник: %s\n", html.EscapeString(source))
	if line := telegramLinkLine("Клиент", event.Links.ClientLink, event.ClientURL); line != "" {
		b.WriteString("\n" + line)
	}
	if line := telegramLinkLine("Услуга", event.Links.ClientServiceLink, event.ServiceURL); line != "" {
		b.WriteString("\n" + line)
	}
	if line := telegramLinkLine("Оплата", event.Links.PaymentLink, event.PaymentURL); line != "" {
		b.WriteString("\n" + line)
	}
	return strings.TrimSpace(b.String())
}

func paymentEventPlain(event paymentDigestEvent) string {
	client := firstNonEmpty(strings.TrimSpace(event.ClientName), "Клиент")
	source := firstNonEmpty(strings.TrimSpace(event.SourceLabel), paymentSourceLabel(event.Source))
	service := firstNonEmpty(strings.TrimSpace(event.ServiceName), strings.TrimSpace(event.ServiceSummary), "услуги")
	period := firstNonEmpty(paymentPeriodDisplay(event), "—")
	next := firstNonEmpty(formatDueDateRU(firstNonEmpty(event.NextPaymentDueDate, event.NextPeriodStart)), "—")
	method := firstNonEmpty(strings.TrimSpace(event.PaymentMethodLabel), paymentMethodLabel(event.PaymentMethod))
	lines := []string{
		"Счет оплачен",
		"",
		"Клиент: " + client,
		"Услуга: " + service,
		"Сумма: " + formatMoneyRUB(event.Amount),
		"Период: " + period,
		"Следующий платеж: " + next,
		"Способ оплаты: " + method,
		"Источник события: " + source,
	}
	if line := plainLinkLine("Клиент", event.Links.ClientLink, event.ClientURL); line != "" {
		lines = append(lines, "", line)
	}
	if line := plainLinkLine("Услуга", event.Links.ClientServiceLink, event.ServiceURL); line != "" {
		lines = append(lines, line)
	}
	if line := plainLinkLine("Оплата", event.Links.PaymentLink, event.PaymentURL); line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func paymentEventEmailHTML(event paymentDigestEvent) string {
	client := firstNonEmpty(strings.TrimSpace(event.ClientName), "Клиент")
	service := firstNonEmpty(strings.TrimSpace(event.ServiceName), strings.TrimSpace(event.ServiceSummary), "услуги")
	period := firstNonEmpty(paymentPeriodDisplay(event), "—")
	next := firstNonEmpty(formatDueDateRU(firstNonEmpty(event.NextPaymentDueDate, event.NextPeriodStart)), "—")
	source := firstNonEmpty(strings.TrimSpace(event.SourceLabel), paymentSourceLabel(event.Source))
	method := firstNonEmpty(strings.TrimSpace(event.PaymentMethodLabel), paymentMethodLabel(event.PaymentMethod))
	linkHTML := strings.Join([]string{
		emailLinkHTML("Открыть клиента", event.Links.ClientLink, event.ClientURL),
		emailLinkHTML("Открыть услугу", event.Links.ClientServiceLink, event.ServiceURL),
		emailLinkHTML("Открыть оплату", event.Links.PaymentLink, event.PaymentURL),
	}, "")
	return fmt.Sprintf(`<tr><td style="padding:18px 24px;border-bottom:1px solid #252a36;background:#172033;">
      <div style="font-size:13px;color:#7dd3fc;text-transform:uppercase;letter-spacing:.06em;margin-bottom:8px;">Счет оплачен</div>
      <div style="color:#e8ecf4;font-size:15px;line-height:1.55;">
        <div><strong>Клиент:</strong> %s</div>
        <div><strong>Услуга:</strong> %s</div>
        <div><strong>Сумма:</strong> %s</div>
        <div><strong>Период:</strong> %s</div>
        <div><strong>Следующий платеж:</strong> %s</div>
        <div><strong>Способ оплаты:</strong> %s</div>
        <div><strong>Источник события:</strong> %s</div>
      </div>
      <div style="margin-top:12px;">%s</div>
    </td></tr>`,
		html.EscapeString(client),
		html.EscapeString(service),
		html.EscapeString(formatMoneyRUB(event.Amount)),
		html.EscapeString(period),
		html.EscapeString(next),
		html.EscapeString(method),
		html.EscapeString(source),
		linkHTML,
	)
}

func paymentDigestEventFromRunInput(runInput []byte) paymentDigestEvent {
	if len(bytesTrim(runInput)) == 0 {
		return paymentDigestEvent{}
	}
	var root any
	if json.Unmarshal(runInput, &root) != nil {
		return paymentDigestEvent{}
	}
	source := firstNonEmpty(
		stringAtRunPath(root, "current_payment.source"),
		stringAtRunPath(root, "payment_event.source"),
		stringAtRunPath(root, "payment_fact.source"),
		stringAtRunPath(root, "payment_action.source"),
	)
	actionID := firstNonEmpty(
		stringAtRunPath(root, "current_payment.payment_action_id"),
		stringAtRunPath(root, "payment_event.payment_action_id"),
		stringAtRunPath(root, "payment_action.payment_action_id"),
		stringAtRunPath(root, "payment_action_id"),
		stringAtRunPath(root, "_pulse.payment_action_id"),
	)
	links := paymentDigestLinksFromRunInput(root)
	event := paymentDigestEvent{
		PreRendered:     strings.EqualFold(stringAtRunPath(root, "_pulse.notification_current_payment_pre_rendered"), "true"),
		PaymentActionID: actionID,
		PaymentID: firstNonEmpty(
			stringAtRunPath(root, "current_payment.payment_id"),
			stringAtRunPath(root, "payment_event.payment_id"),
			stringAtRunPath(root, "payment_fact.payment_fact_id"),
		),
		PaymentFactID: firstNonEmpty(
			stringAtRunPath(root, "current_payment.payment_fact_id"),
			stringAtRunPath(root, "payment_event.payment_fact_id"),
			stringAtRunPath(root, "payment_fact.payment_fact_id"),
		),
		ClientID: firstNonEmpty(
			stringAtRunPath(root, "current_payment.client_id"),
			stringAtRunPath(root, "payment_event.client_id"),
			stringAtRunPath(root, "client.client_id"),
			stringAtRunPath(root, "_pulse.client_id"),
		),
		ClientServiceID: firstNonEmpty(
			stringAtRunPath(root, "current_payment.client_service_id"),
			stringAtRunPath(root, "payment_event.client_service_id"),
			stringAtRunPath(root, "payment_event.client_service.client_service_id"),
			stringAtRunPath(root, "client_service.client_service_id"),
		),
		ClientName: firstNonEmpty(
			stringAtRunPath(root, "current_payment.client_name"),
			stringAtRunPath(root, "payment_event.client.name"),
			stringAtRunPath(root, "client.name"),
		),
		ServiceName: firstNonEmpty(
			stringAtRunPath(root, "current_payment.service_name"),
			stringAtRunPath(root, "payment_event.service_name"),
			stringAtRunPath(root, "payment_event.client_service.name"),
			stringAtRunPath(root, "client_service.name"),
		),
		ServiceSummary: firstNonEmpty(
			stringAtRunPath(root, "current_payment.service_summary"),
			stringAtRunPath(root, "payment_event.service_summary"),
			stringAtRunPath(root, "payment_action.service_summary"),
		),
		Source: source,
		SourceLabel: firstNonEmpty(
			stringAtRunPath(root, "current_payment.source_label"),
			stringAtRunPath(root, "payment_event.source_label"),
			paymentSourceLabel(source),
		),
		PaymentMethod: firstNonEmpty(
			stringAtRunPath(root, "current_payment.payment_method"),
			stringAtRunPath(root, "payment_event.payment_method"),
			stringAtRunPath(root, "payment_fact.payment_method"),
			stringAtRunPath(root, "payment_action.payment_method"),
		),
		PaymentMethodLabel: firstNonEmpty(
			stringAtRunPath(root, "current_payment.payment_method_label"),
			stringAtRunPath(root, "payment_event.payment_method_label"),
		),
		Amount: firstFloatAtRunPath(root,
			"current_payment.amount",
			"payment_event.amount",
			"payment_fact.amount",
			"payment_action.amount",
			"totals.amount",
		),
		Currency: firstNonEmpty(
			stringAtRunPath(root, "current_payment.currency"),
			stringAtRunPath(root, "payment_event.currency"),
			stringAtRunPath(root, "payment_fact.currency"),
			stringAtRunPath(root, "totals.currency"),
			"RUB",
		),
		PaidAt: firstNonEmpty(
			stringAtRunPath(root, "current_payment.paid_at"),
			stringAtRunPath(root, "payment_event.paid_at"),
			stringAtRunPath(root, "payment_fact.paid_at"),
			stringAtRunPath(root, "totals.paid_at"),
		),
		PeriodStart: firstNonEmpty(
			stringAtRunPath(root, "current_payment.period_start"),
			stringAtRunPath(root, "payment_event.period_start"),
			stringAtRunPath(root, "payment_action.period_start"),
			stringAtRunPath(root, "totals.period_start"),
		),
		PeriodEnd: firstNonEmpty(
			stringAtRunPath(root, "current_payment.period_end"),
			stringAtRunPath(root, "payment_event.period_end"),
			stringAtRunPath(root, "payment_action.period_end"),
			stringAtRunPath(root, "totals.period_end"),
		),
		PeriodLabel: firstNonEmpty(
			stringAtRunPath(root, "current_payment.period_label"),
			stringAtRunPath(root, "payment_event.period_label"),
			stringAtRunPath(root, "payment_action.period_label"),
		),
		NextPeriodStart: firstNonEmpty(
			stringAtRunPath(root, "current_payment.next_period_start"),
			stringAtRunPath(root, "payment_event.next_period_start"),
			stringAtRunPath(root, "payment_action.next_period_start"),
			stringAtRunPath(root, "totals.next_period_start"),
		),
		NextPaymentDueDate: firstNonEmpty(
			stringAtRunPath(root, "current_payment.next_payment_due_date"),
			stringAtRunPath(root, "payment_event.next_payment_due_date"),
			stringAtRunPath(root, "payment_action.next_payment_due_date"),
		),
		PaymentURL: firstNonEmpty(
			absoluteURL(links.PaymentLink),
			stringAtRunPath(root, "current_payment.links.payment_link.url"),
			stringAtRunPath(root, "payment_event.links.payment_link.url"),
			stringAtRunPath(root, "payment_event.url"),
			stringAtRunPath(root, "payment_action.url"),
		),
		ClientURL: firstNonEmpty(
			absoluteURL(links.ClientLink),
			stringAtRunPath(root, "current_payment.links.client_link.url"),
			stringAtRunPath(root, "payment_event.links.client_link.url"),
			stringAtRunPath(root, "payment_event.client.url"),
			stringAtRunPath(root, "client.url"),
		),
		ServiceURL: firstNonEmpty(
			absoluteURL(links.ClientServiceLink),
			stringAtRunPath(root, "current_payment.links.client_service_link.url"),
			stringAtRunPath(root, "payment_event.links.client_service_link.url"),
			stringAtRunPath(root, "payment_event.client_service.url"),
			stringAtRunPath(root, "client_service.url"),
		),
		ProjectURL: firstNonEmpty(
			stringAtRunPath(root, "payment_event.project.url"),
			stringAtRunPath(root, "project.url"),
		),
		Links: links,
	}
	event.Present = actionID != "" || event.Amount > 0 || source != ""
	return event
}

func currentPaymentContextFromEvent(event paymentDigestEvent) map[string]any {
	if !event.Present {
		return map[string]any{}
	}
	links := paymentDigestLinksMap(event)
	return map[string]any{
		"payment_id":            firstNonEmpty(event.PaymentID, event.PaymentFactID, event.PaymentActionID),
		"payment_action_id":     event.PaymentActionID,
		"payment_fact_id":       event.PaymentFactID,
		"client_id":             event.ClientID,
		"client_service_id":     event.ClientServiceID,
		"client_name":           firstNonEmpty(event.ClientName, "Клиент"),
		"service_name":          firstNonEmpty(event.ServiceName, event.ServiceSummary),
		"service_summary":       event.ServiceSummary,
		"amount":                event.Amount,
		"currency":              firstNonEmpty(event.Currency, "RUB"),
		"period_label":          paymentPeriodDisplay(event),
		"period_start":          event.PeriodStart,
		"period_end":            event.PeriodEnd,
		"next_period_start":     event.NextPeriodStart,
		"next_payment_due_date": firstNonEmpty(event.NextPaymentDueDate, event.NextPeriodStart),
		"paid_at":               event.PaidAt,
		"source":                event.Source,
		"source_label":          firstNonEmpty(event.SourceLabel, paymentSourceLabel(event.Source)),
		"payment_method":        event.PaymentMethod,
		"payment_method_label":  firstNonEmpty(event.PaymentMethodLabel, paymentMethodLabel(event.PaymentMethod)),
		"links":                 links,
	}
}

func paymentEventContextFromEvent(event paymentDigestEvent) map[string]any {
	out := currentPaymentContextFromEvent(event)
	out["present"] = event.Present
	out["event_type"] = "payment.recorded"
	out["trigger_type"] = "manual_payment_trigger"
	return out
}

func paymentDigestLinksFromRunInput(root any) paymentDigestLinks {
	return paymentDigestLinks{
		ClientLink:        paymentDigestLinkFromRunInput(root, "current_payment.links.client_link", "payment_event.links.client_link", "links.client_link"),
		ClientServiceLink: paymentDigestLinkFromRunInput(root, "current_payment.links.client_service_link", "payment_event.links.client_service_link", "links.client_service_link"),
		PaymentLink:       paymentDigestLinkFromRunInput(root, "current_payment.links.payment_link", "payment_event.links.payment_link", "links.payment_link"),
	}
}

func paymentDigestLinkFromRunInput(root any, paths ...string) paymentDigestLink {
	for _, path := range paths {
		link := paymentDigestLink{
			Label:           stringAtRunPath(root, path+".label"),
			Href:            stringAtRunPath(root, path+".href"),
			URL:             stringAtRunPath(root, path+".url"),
			ClientID:        stringAtRunPath(root, path+".client_id"),
			ClientServiceID: stringAtRunPath(root, path+".client_service_id"),
			PaymentID:       stringAtRunPath(root, path+".payment_id"),
			PaymentActionID: stringAtRunPath(root, path+".payment_action_id"),
		}
		if link.Label != "" || link.Href != "" || link.URL != "" {
			return link
		}
	}
	return paymentDigestLink{}
}

func paymentDigestLinksMap(event paymentDigestEvent) map[string]any {
	return map[string]any{
		"client_link":         paymentDigestLinkMap(event.Links.ClientLink, event.ClientURL),
		"client_service_link": paymentDigestLinkMap(event.Links.ClientServiceLink, event.ServiceURL),
		"payment_link":        paymentDigestLinkMap(event.Links.PaymentLink, event.PaymentURL),
	}
}

func paymentDigestLinkMap(link paymentDigestLink, fallbackURL string) map[string]any {
	url := firstNonEmpty(absoluteURL(link), fallbackURL)
	return map[string]any{
		"label":             link.Label,
		"client_id":         link.ClientID,
		"client_service_id": link.ClientServiceID,
		"payment_id":        link.PaymentID,
		"payment_action_id": link.PaymentActionID,
		"href":              link.Href,
		"url":               url,
	}
}

func paymentPeriodDisplay(event paymentDigestEvent) string {
	if strings.TrimSpace(event.PeriodLabel) != "" {
		return strings.TrimSpace(event.PeriodLabel)
	}
	start := strings.TrimSpace(event.PeriodStart)
	end := strings.TrimSpace(event.PeriodEnd)
	if start == "" && end == "" {
		return ""
	}
	startTime, startErr := time.Parse("2006-01-02", start)
	if startErr == nil && end != "" {
		endTime, endErr := time.Parse("2006-01-02", end)
		if endErr == nil {
			monthEnd := time.Date(startTime.Year(), startTime.Month()+1, 0, 0, 0, 0, 0, time.UTC)
			if startTime.Day() == 1 && endTime.Format("2006-01-02") == monthEnd.Format("2006-01-02") {
				return russianMonthLabel(startTime.Month()) + " " + fmt.Sprint(startTime.Year())
			}
		}
	}
	return strings.TrimPrefix(periodLabelRU(start, end), " за период ")
}

func russianMonthLabel(month time.Month) string {
	switch month {
	case time.January:
		return "январь"
	case time.February:
		return "февраль"
	case time.March:
		return "март"
	case time.April:
		return "апрель"
	case time.May:
		return "май"
	case time.June:
		return "июнь"
	case time.July:
		return "июль"
	case time.August:
		return "август"
	case time.September:
		return "сентябрь"
	case time.October:
		return "октябрь"
	case time.November:
		return "ноябрь"
	case time.December:
		return "декабрь"
	default:
		return ""
	}
}

func paymentMethodLabel(method string) string {
	switch strings.TrimSpace(method) {
	case "cash":
		return "наличные"
	case "legal_invoice", "":
		return "по счету с юрлица"
	case "bank_transfer":
		return "банковский перевод"
	case "card":
		return "карта"
	case "other":
		return "другое"
	default:
		return strings.TrimSpace(method)
	}
}

func telegramLinkLine(label string, link paymentDigestLink, fallbackURL string) string {
	url := firstNonEmpty(absoluteURL(link), strings.TrimSpace(fallbackURL))
	if url == "" {
		return ""
	}
	text := firstNonEmpty(strings.TrimSpace(link.Label), label)
	return fmt.Sprintf(`%s: <a href="%s">%s</a>`, html.EscapeString(label), html.EscapeString(url), html.EscapeString(text))
}

func plainLinkLine(label string, link paymentDigestLink, fallbackURL string) string {
	url := firstNonEmpty(absoluteURL(link), strings.TrimSpace(fallbackURL))
	if url == "" {
		return ""
	}
	return label + ": " + url
}

func emailLinkHTML(label string, link paymentDigestLink, fallbackURL string) string {
	url := firstNonEmpty(absoluteURL(link), strings.TrimSpace(fallbackURL))
	if url == "" {
		return ""
	}
	return fmt.Sprintf(`<div><a href="%s" style="color:#6ea8fe;text-decoration:none;font-weight:600;">%s</a></div>`,
		html.EscapeString(url),
		html.EscapeString(label),
	)
}

func absoluteURL(link paymentDigestLink) string {
	if strings.TrimSpace(link.URL) != "" {
		return strings.TrimSpace(link.URL)
	}
	href := strings.TrimSpace(link.Href)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if !strings.HasPrefix(href, "/") {
		href = "/" + href
	}
	return publicAppBaseURL() + href
}

func ternaryString(cond bool, yes, no string) string {
	if cond {
		return yes
	}
	return no
}

func firstFloatAtRunPath(root any, paths ...string) float64 {
	for _, path := range paths {
		value, ok := valueAtPathRaw(root, path)
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return typed
		case int:
			return float64(typed)
		case string:
			var parsed float64
			if _, err := fmt.Sscanf(strings.ReplaceAll(typed, ",", "."), "%f", &parsed); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func paymentSourceLabel(source string) string {
	switch strings.TrimSpace(source) {
	case "manual":
		return "ручное внесение"
	case "bank":
		return "банк"
	case "email", "incoming_email":
		return "входящее письмо"
	case "kontur_elba":
		return "Контур Эльба"
	case "other_integration", "api_key", "integration":
		return "интеграция"
	case "workflow":
		return "сценарий"
	default:
		if strings.TrimSpace(source) == "" {
			return "не указан"
		}
		return strings.TrimSpace(source)
	}
}

func periodLabelRU(start, end string) string {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if start == "" && end == "" {
		return ""
	}
	if start == end || end == "" {
		return " за период " + formatDueDateRU(start)
	}
	return " за период " + formatDueDateRU(start) + " — " + formatDueDateRU(end)
}

func clientCountLabel(n int) string {
	n = n % 100
	if n >= 11 && n <= 14 {
		return "клиентов"
	}
	switch n % 10 {
	case 1:
		return "клиент"
	case 2, 3, 4:
		return "клиента"
	default:
		return "клиентов"
	}
}

func formatMoneyRUB(amount float64) string {
	s := fmt.Sprintf("%.0f", amount)
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append([]string{s}, parts...)
	}
	return strings.Join(parts, " ") + " ₽"
}

func formatDueDateRU(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "—"
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z07:00", "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.Format("02.01.2006")
		}
	}
	if len(raw) >= 10 {
		return raw[:10]
	}
	return raw
}
