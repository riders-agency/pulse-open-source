// Package invoicing defines the bank-neutral invoice boundary. No bank is connected by default.
package invoicing

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid invoice draft")
var ErrConflict = errors.New("invoice changed or payment closed")

type Input struct {
	Revision  int    `json:"revision"`
	Kind      string `json:"kind"`
	PayerName string `json:"payer_name"`
	TaxID     string `json:"tax_id"`
	Purpose   string `json:"purpose"`
	DueDate   string `json:"due_date"`
}

type Draft struct {
	Input
	PaymentActionID string    `json:"payment_action_id"`
	AmountMinor     string    `json:"amount_minor"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Workspace struct {
	CanEdit       bool           `json:"can_edit"`
	Draft         *Draft         `json:"draft"`
	AmountMinor   string         `json:"amount_minor"`
	Currency      string         `json:"currency"`
	PaymentStatus string         `json:"payment_status"`
	PayerName     string         `json:"payer_name"`
	Providers     []Capabilities `json:"providers"`
}

func Normalize(in Input) (Input, error) {
	in.PayerName = strings.TrimSpace(in.PayerName)
	in.TaxID = strings.TrimSpace(in.TaxID)
	in.Purpose = strings.TrimSpace(in.Purpose)
	if in.Revision < 0 || in.Revision > 2147483646 || (in.Kind != "bank_invoice" && in.Kind != "payment_link") {
		return in, ErrInvalid
	}
	if in.PayerName == "" || in.Purpose == "" || utf8.RuneCountInString(in.PayerName) > 240 || utf8.RuneCountInString(in.Purpose) > 1000 || utf8.RuneCountInString(in.TaxID) > 32 {
		return in, ErrInvalid
	}
	if in.DueDate != "" {
		if _, err := time.Parse(time.DateOnly, in.DueDate); err != nil {
			return in, ErrInvalid
		}
	}
	return in, nil
}
