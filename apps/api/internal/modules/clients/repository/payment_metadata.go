package repository

import (
	"fmt"
	"strings"
)

func normalizePaymentDocumentKind(value string) string {
	switch strings.TrimSpace(value) {
	case "invoice", "act", "receipt", "reconciliation":
		return strings.TrimSpace(value)
	default:
		return "file"
	}
}

func normalizePaymentMethod(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "legal_invoice", nil
	}
	switch value {
	case "cash", "legal_invoice", "bank_transfer", "card", "other":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported payment_method %q", value)
	}
}

func paymentMethodPublicLabel(method string) string {
	switch strings.TrimSpace(method) {
	case "cash":
		return "наличные"
	case "legal_invoice":
		return "по счету с юрлица"
	case "bank_transfer":
		return "банковский перевод"
	case "card":
		return "карта"
	case "other":
		return "другое"
	default:
		if strings.TrimSpace(method) == "" {
			return "по счету с юрлица"
		}
		return strings.TrimSpace(method)
	}
}
