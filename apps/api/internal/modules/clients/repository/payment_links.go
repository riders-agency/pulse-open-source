package repository

import (
	"os"
	"strconv"
	"strings"
)

func clientWorkspaceHref(clientID string) string {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "/clients"
	}
	return "/clients/" + clientID
}

func clientServiceWorkspaceHref(clientID, serviceID string) string {
	clientID = strings.TrimSpace(clientID)
	serviceID = strings.TrimSpace(serviceID)
	if clientID == "" {
		return "/clients"
	}
	if serviceID == "" {
		return "/clients/" + clientID + "?tab=services"
	}
	return "/clients/" + clientID + "?tab=services&client_service_id=" + serviceID
}

func paymentWorkspaceHref(clientID, paymentActionID, paymentFactID string) string {
	clientID = strings.TrimSpace(clientID)
	paymentActionID = strings.TrimSpace(paymentActionID)
	paymentFactID = strings.TrimSpace(paymentFactID)
	if clientID == "" {
		return "/clients"
	}
	href := "/clients/" + clientID + "?tab=payments"
	if paymentActionID != "" {
		href += "&payment_action_id=" + paymentActionID
	}
	if paymentFactID != "" {
		href += "&payment_id=" + paymentFactID
	}
	return href
}

func paymentLinks(clientID, clientName, serviceID, serviceName, paymentActionID, paymentFactID string, amount float64, currency, periodStart, periodEnd string) map[string]any {
	period := paymentPeriodLabel(periodStart, periodEnd)
	paymentLabel := strings.TrimSpace("Оплата " + formatPaymentAmount(amount, currency))
	if period != "" {
		paymentLabel += " за " + period
	}
	clientHref := clientWorkspaceHref(clientID)
	serviceHref := clientServiceWorkspaceHref(clientID, serviceID)
	paymentHref := paymentWorkspaceHref(clientID, paymentActionID, paymentFactID)
	return map[string]any{
		"client_link": map[string]any{
			"label":     firstNonEmpty(clientName, "Клиент"),
			"client_id": clientID,
			"href":      clientHref,
			"url":       pulsePublicURL(clientHref),
		},
		"client_service_link": map[string]any{
			"label":             firstNonEmpty(serviceName, "Услуга"),
			"client_id":         clientID,
			"client_service_id": serviceID,
			"href":              serviceHref,
			"url":               pulsePublicURL(serviceHref),
		},
		"payment_link": map[string]any{
			"label":             paymentLabel,
			"client_id":         clientID,
			"client_service_id": serviceID,
			"payment_id":        firstNonEmpty(paymentFactID, paymentActionID),
			"payment_action_id": paymentActionID,
			"href":              paymentHref,
			"url":               pulsePublicURL(paymentHref),
		},
	}
}

func formatPaymentAmount(amount float64, currency string) string {
	s := strconv.FormatFloat(amount, 'f', 0, 64)
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append([]string{s}, parts...)
	}
	suffix := strings.TrimSpace(currency)
	if suffix == "" || suffix == "RUB" {
		suffix = "₽"
	}
	return strings.Join(parts, " ") + " " + suffix
}

func paymentActionPublicURL(paymentActionID string) string {
	return pulsePublicURL("/payment-actions/" + strings.TrimSpace(paymentActionID))
}

func pulsePublicURL(path string) string {
	base := strings.TrimRight(strings.TrimSpace(firstNonEmpty(
		os.Getenv("RIDERS_PULSE_PUBLIC_URL"),
		os.Getenv("NEXT_PUBLIC_APP_URL"),
		os.Getenv("PUBLIC_APP_URL"),
		"https://pulse.riders.agency",
	)), "/")
	path = "/" + strings.TrimLeft(strings.TrimSpace(path), "/")
	return base + path
}

func paymentSourcePublicLabel(source string) string {
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
