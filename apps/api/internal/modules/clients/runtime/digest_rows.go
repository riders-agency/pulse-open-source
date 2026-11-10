package runtime

import (
	"fmt"
	"html"
	"strings"
	"time"
)

type paymentDigestGroup struct {
	Key   string
	Icon  string
	Title string
	Rows  []paymentCycleRow
	Total float64
}

func coalescePaymentRows(rows []paymentCycleRow, now time.Time) []paymentCycleRow {
	out := make([]paymentCycleRow, 0, len(rows))
	index := make(map[string]int, len(rows))
	for _, row := range rows {
		key := paymentDigestGroupKey(row, now) + ":" + paymentDigestClientKey(row)
		if key == ":" {
			out = append(out, row)
			continue
		}
		if pos, ok := index[key]; ok {
			out[pos] = mergePaymentCycleRows(out[pos], row)
			continue
		}
		index[key] = len(out)
		out = append(out, row)
	}
	return out
}

func paymentDigestGroups(rows []paymentCycleRow, now time.Time) []paymentDigestGroup {
	order := []paymentDigestGroup{
		{Key: "sent", Icon: "💰", Title: "Ждем денег"},
		{Key: "scheduled", Icon: "🧾", Title: "Ждут счета"},
	}
	byKey := make(map[string][]paymentCycleRow, len(order))
	for _, row := range rows {
		key := paymentDigestGroupKey(row, now)
		byKey[key] = append(byKey[key], row)
	}
	out := make([]paymentDigestGroup, 0, len(order))
	for _, group := range order {
		group.Rows = byKey[group.Key]
		if len(group.Rows) == 0 {
			continue
		}
		group.Total = paymentRowsTotal(group.Rows)
		out = append(out, group)
	}
	return out
}

func paymentDigestGroupKey(row paymentCycleRow, now time.Time) string {
	switch strings.TrimSpace(row.CycleState) {
	case "sent":
		return "sent"
	default:
		return "scheduled"
	}
}

func paymentDigestClientKey(row paymentCycleRow) string {
	if id := strings.TrimSpace(row.ClientID); id != "" {
		return id
	}
	return strings.ToLower(strings.TrimSpace(row.ClientName))
}

func mergePaymentCycleRows(a, b paymentCycleRow) paymentCycleRow {
	a.TotalAmount += b.TotalAmount
	a.ServiceCount += b.ServiceCount
	a.ServiceSummary = mergePaymentSummary(a.ServiceSummary, b.ServiceSummary)
	if paymentDigestDateBefore(b.DueDate, a.DueDate) {
		a.DueDate = b.DueDate
	}
	if strings.TrimSpace(a.PaymentActionID) == "" {
		a.PaymentActionID = b.PaymentActionID
	}
	if strings.TrimSpace(a.StateLabel) == "" {
		a.StateLabel = b.StateLabel
	}
	if strings.TrimSpace(a.PaymentStatus) == "" {
		a.PaymentStatus = b.PaymentStatus
	}
	return a
}

func mergePaymentSummary(left, right string) string {
	seen := map[string]bool{}
	parts := make([]string, 0, 4)
	add := func(value string) {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			parts = append(parts, part)
		}
	}
	add(left)
	add(right)
	return strings.Join(parts, ", ")
}

func paymentRowsTotal(rows []paymentCycleRow) float64 {
	var total float64
	for _, row := range rows {
		total += row.TotalAmount
	}
	return total
}

func writePaymentDigestGroup(tg, max, plain, rowsHTML *strings.Builder, group paymentDigestGroup, now time.Time) {
	count := len(group.Rows)
	header := fmt.Sprintf("%s %s: %d %s · %s", group.Icon, group.Title, count, clientCountLabel(count), formatMoneyRUB(group.Total))
	fmt.Fprintf(tg, "<b>%s</b>\n", html.EscapeString(header))
	fmt.Fprintf(max, "%s\n", header)
	fmt.Fprintf(plain, "%s\n", header)
	fmt.Fprintf(rowsHTML, `<tr><td colspan="2" style="padding:14px 14px 8px;color:#e8ecf4;font-size:14px;font-weight:700;border-bottom:1px solid #252a36;">%s</td></tr>`, html.EscapeString(header))
	for i, row := range group.Rows {
		writePaymentDigestRow(tg, max, plain, rowsHTML, row, group, i+1, now)
	}
	tg.WriteString("\n")
	max.WriteString("\n")
	plain.WriteString("\n")
}

func writePaymentDigestRow(tg, max, plain, rowsHTML *strings.Builder, row paymentCycleRow, group paymentDigestGroup, n int, now time.Time) {
	client := firstNonEmpty(strings.TrimSpace(row.ClientName), "Клиент")
	due := formatDueDateRU(row.DueDate)
	amount := formatMoneyRUB(row.TotalAmount)
	status := paymentDigestRowStatus(row, group.Key, now)
	serviceSummary := strings.TrimSpace(row.ServiceSummary)
	duePrefix := "до"
	if paymentDigestIsOverdue(row, now) {
		duePrefix = "с"
	}

	fmt.Fprintf(tg, "%s <b>%d.</b> %s — <b>%s</b>, %s %s", group.Icon, n, html.EscapeString(client), html.EscapeString(amount), duePrefix, html.EscapeString(due))
	if row.ServiceCount > 1 {
		fmt.Fprintf(tg, " · %d усл.", row.ServiceCount)
	}
	fmt.Fprintf(tg, " · %s\n", html.EscapeString(status))
	if serviceSummary != "" {
		fmt.Fprintf(tg, "   %s\n", html.EscapeString(serviceSummary))
	}

	fmt.Fprintf(max, "%s %d. %s — %s, %s %s", group.Icon, n, client, amount, duePrefix, due)
	if row.ServiceCount > 1 {
		fmt.Fprintf(max, " · %d усл.", row.ServiceCount)
	}
	fmt.Fprintf(max, " · %s\n", status)
	if serviceSummary != "" {
		fmt.Fprintf(max, "   %s\n", serviceSummary)
	}

	fmt.Fprintf(plain, "%d. %s — %s, %s %s (%s)", n, client, amount, duePrefix, due, status)
	if serviceSummary != "" {
		fmt.Fprintf(plain, " · %s", serviceSummary)
	}
	plain.WriteString("\n")

	link := paymentActionURL(row.PaymentActionID)
	linkCell := `<span style="color:#9aa3b2;font-size:12px;">счёт в Pulse</span>`
	if link != "" {
		linkCell = fmt.Sprintf(`<a href="%s" style="color:#6ea8fe;text-decoration:none;font-weight:600;">Открыть →</a>`, html.EscapeString(link))
	}
	fmt.Fprintf(rowsHTML,
		`<tr><td style="padding:12px 14px;border-bottom:1px solid #252a36;color:#e8ecf4;font-size:14px;">%s <strong>%s</strong><br><span style="color:#9aa3b2;font-size:12px;">%s %s · %s · %s</span></td><td style="padding:12px 14px;border-bottom:1px solid #252a36;text-align:right;white-space:nowrap;">%s</td></tr>`,
		group.Icon, html.EscapeString(client), duePrefix, html.EscapeString(due), html.EscapeString(amount), html.EscapeString(status), linkCell,
	)
}

func paymentDigestRowStatus(row paymentCycleRow, groupKey string, now time.Time) string {
	if paymentDigestIsOverdue(row, now) {
		due, ok := parsePaymentDigestDate(row.DueDate)
		if ok {
			days := int(midnightUTC(now).Sub(due).Hours() / 24)
			if days > 0 {
				return fmt.Sprintf("Просрочено %d %s", days, dayLabel(days))
			}
		}
		return "Просрочено"
	}
	if paymentDigestIsDueToday(row, now) {
		return "Срок сегодня"
	}
	if label := strings.TrimSpace(row.StateLabel); label != "" {
		return label
	}
	return paymentStatusLabel(row.PaymentStatus)
}

func paymentDigestIsOverdue(row paymentCycleRow, now time.Time) bool {
	due, ok := parsePaymentDigestDate(row.DueDate)
	return ok && due.Before(midnightUTC(now))
}

func paymentDigestIsDueToday(row paymentCycleRow, now time.Time) bool {
	due, ok := parsePaymentDigestDate(row.DueDate)
	return ok && due.Equal(midnightUTC(now))
}

func dayLabel(n int) string {
	n = n % 100
	if n >= 11 && n <= 14 {
		return "дней"
	}
	switch n % 10 {
	case 1:
		return "день"
	case 2, 3, 4:
		return "дня"
	default:
		return "дней"
	}
}

func paymentDigestDateBefore(left, right string) bool {
	l, ok := parsePaymentDigestDate(left)
	if !ok {
		return false
	}
	r, ok := parsePaymentDigestDate(right)
	if !ok {
		return true
	}
	return l.Before(r)
}

func parsePaymentDigestDate(raw string) (time.Time, bool) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, false
	}
	return midnightUTC(parsed), true
}

func midnightUTC(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func paymentDigestEmailShell(projectName string, count int, total float64, eventHTML, rowsHTML string) string {
	summary := fmt.Sprintf("%d %s", count, clientCountLabel(count))
	if count > 0 {
		summary = fmt.Sprintf("%s · %s", summary, formatMoneyRUB(total))
	}
	pictureHeading := ""
	if strings.TrimSpace(eventHTML) != "" {
		pictureHeading = `<tr><td style="padding:16px 24px 8px;color:#e8ecf4;font-size:16px;font-weight:700;">Общая картина по оплатам</td></tr>`
	}
	if strings.TrimSpace(rowsHTML) == "" {
		rowsHTML = `<tr><td style="padding:14px;color:#9aa3b2;font-size:14px;">Сейчас нет клиентов в окне напоминания.</td></tr>`
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:24px;background:#0b0d12;font-family:Inter,Segoe UI,Roboto,Arial,sans-serif;color:#e8ecf4;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:640px;margin:0 auto;background:#12151c;border:1px solid #252a36;border-radius:14px;overflow:hidden;">
    <tr><td style="padding:22px 24px;background:linear-gradient(135deg,#1a2233,#12151c);border-bottom:1px solid #252a36;">
      <div style="font-size:12px;letter-spacing:.08em;text-transform:uppercase;color:#6ea8fe;margin-bottom:8px;">Riders Pulse</div>
      <div style="font-size:22px;font-weight:700;">💳 Отчет по оплатам</div>
      <div style="margin-top:8px;color:#9aa3b2;font-size:14px;">📁 %s · %s</div>
    </td></tr>
    %s
    %s
    <tr><td style="padding:0 8px;">
      <table role="presentation" width="100%%" cellspacing="0" cellpadding="0">%s</table>
    </td></tr>
    <tr><td style="padding:16px 24px 22px;color:#6b7280;font-size:12px;border-top:1px solid #252a36;">
      Операционное уведомление · <a href="https://pulse.riders.agency" style="color:#6ea8fe;">pulse.riders.agency</a>
    </td></tr>
  </table>
</body>
</html>`, html.EscapeString(projectName), html.EscapeString(summary), eventHTML, pictureHeading, rowsHTML)
}
