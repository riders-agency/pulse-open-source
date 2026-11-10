package runtime

import (
	"fmt"
	"html"
	"strings"
)

func buildPaymentDigest(rows []paymentCycleRow, projectName, projectID string, event paymentDigestEvent) paymentDigest {
	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		projectName = "Проект"
	}
	now := paymentDigestNow()
	groupedRows := coalescePaymentRows(rows, now)
	groups := paymentDigestGroups(groupedRows, now)
	count := len(groupedRows)
	total := paymentRowsTotal(groupedRows)
	listURL := projectPaymentsURL(projectID)
	subject := fmt.Sprintf("💳 Отчет по оплатам · %d %s · %s", count, clientCountLabel(count), projectName)
	if count == 0 && !event.Present {
		return paymentDigest{
			Subject:      subject,
			TelegramHTML: "<b>💳 Отчет по оплатам</b>\n\nСейчас нет клиентов в окне напоминания.",
			MaxText:      "💳 Отчет по оплатам\n\nСейчас нет клиентов в окне напоминания.",
			EmailHTML:    paymentDigestEmailShell(projectName, 0, 0, "", "<p style=\"margin:0;color:#9aa3b2;\">Сейчас нет клиентов в окне напоминания.</p>"),
			PlainText:    "Отчет по оплатам\n\nСейчас нет клиентов в окне напоминания.",
		}
	}

	eventHTML := ""
	var tg strings.Builder
	var max strings.Builder
	var plain strings.Builder
	if event.Present && !event.PreRendered {
		eventSubject := paymentEventSubject(event)
		subject = eventSubject + " · " + subject
		fmt.Fprintf(&tg, "%s\n\n", paymentEventTelegram(event))
		fmt.Fprintf(&max, "%s\n\n", paymentEventPlain(event))
		fmt.Fprintf(&plain, "%s\n\n", paymentEventPlain(event))
		eventHTML = paymentEventEmailHTML(event)
	}
	fmt.Fprintf(&tg, "<b>💳 %s</b>\n", ternaryString(event.Present, "Общая картина по оплатам", "Отчет по оплатам"))
	fmt.Fprintf(&tg, "📁 <b>%s</b>\n", html.EscapeString(projectName))
	fmt.Fprintf(&tg, "👥 <b>%d</b> %s · 💰 <b>%s</b>\n\n", count, clientCountLabel(count), formatMoneyRUB(total))

	fmt.Fprintf(&max, "💳 %s\n", ternaryString(event.Present, "Общая картина по оплатам", "Отчет по оплатам"))
	fmt.Fprintf(&max, "📁 %s\n", projectName)
	fmt.Fprintf(&max, "👥 %d %s · 💰 %s\n\n", count, clientCountLabel(count), formatMoneyRUB(total))

	fmt.Fprintf(&plain, "%s\nПроект: %s\nКлиентов: %d · Сумма: %s\n\n",
		ternaryString(event.Present, "Общая картина по оплатам", "Отчет по оплатам"), projectName, count, formatMoneyRUB(total))

	var rowsHTML strings.Builder
	if len(groups) == 0 {
		tg.WriteString("Сейчас нет клиентов в окне напоминания.\n")
		max.WriteString("Сейчас нет клиентов в окне напоминания.\n")
		plain.WriteString("Сейчас нет клиентов в окне напоминания.\n")
	} else {
		for _, group := range groups {
			writePaymentDigestGroup(&tg, &max, &plain, &rowsHTML, group, now)
		}
	}

	fmt.Fprintf(&tg, "\n📋 <a href=\"%s\">Все клиенты в Pulse</a>", html.EscapeString(listURL))
	fmt.Fprintf(&tg, "\n<i>⚡ Riders Pulse · сводка по проекту</i>")
	fmt.Fprintf(&max, "\n📋 Все клиенты: %s", listURL)
	max.WriteString("\n⚡ Riders Pulse")
	fmt.Fprintf(&plain, "\nВсе клиенты: %s\n—\nRiders Pulse", listURL)

	body := rowsHTML.String()
	emailHTML := paymentDigestEmailShell(projectName, count, total, eventHTML, body)

	return paymentDigest{
		Subject:      subject,
		TelegramHTML: strings.TrimSpace(tg.String()),
		MaxText:      strings.TrimSpace(max.String()),
		EmailHTML:    emailHTML,
		PlainText:    strings.TrimSpace(plain.String()),
	}
}
