package invoicing

import "context"

// Capabilities describes an installed adapter, not a configured bank account.
type Capabilities struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	BankInvoice bool   `json:"bank_invoice"`
	PaymentLink bool   `json:"payment_link"`
}

// IssueRequest is built on the server after approval and bank-specific validation.
// Credentials stay inside the adapter and are never part of this contract.
type IssueRequest struct {
	Draft          Draft
	ConnectionID   string
	IssuerID       string
	IdempotencyKey string
}

type Receipt struct {
	Reference  string
	Status     string
	PaymentURL string
}

// Provider implementations must reconcile an uncertain Create result using Lookup.
// A transport timeout is not proof that a bank did not create an invoice.
// No adapter is shipped until its sandbox and authentication have been verified.
type Provider interface {
	Capabilities() Capabilities
	Create(context.Context, IssueRequest) (Receipt, error)
	Lookup(context.Context, string, string) (Receipt, error) // connection, idempotency key
}

type Registry struct{ providers []Provider }

func NewRegistry(providers ...Provider) Registry {
	return Registry{providers: append([]Provider(nil), providers...)}
}

func (r Registry) Capabilities() []Capabilities {
	out := make([]Capabilities, 0, len(r.providers))
	for _, provider := range r.providers {
		out = append(out, provider.Capabilities())
	}
	return out
}
