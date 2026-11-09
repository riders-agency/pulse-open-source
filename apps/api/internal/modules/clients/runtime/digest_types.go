package runtime

import "time"

type paymentDigest struct {
	Subject      string
	TelegramHTML string
	MaxText      string
	EmailHTML    string
	PlainText    string
}

type paymentDigestEvent struct {
	Present            bool
	PreRendered        bool
	PaymentID          string
	PaymentFactID      string
	PaymentActionID    string
	ClientID           string
	ClientServiceID    string
	ClientName         string
	ServiceName        string
	ServiceSummary     string
	Source             string
	SourceLabel        string
	PaymentMethod      string
	PaymentMethodLabel string
	Amount             float64
	Currency           string
	PaidAt             string
	PeriodStart        string
	PeriodEnd          string
	PeriodLabel        string
	NextPeriodStart    string
	NextPaymentDueDate string
	PaymentURL         string
	ClientURL          string
	ServiceURL         string
	ProjectURL         string
	Links              paymentDigestLinks
}

type paymentDigestLinks struct {
	ClientLink        paymentDigestLink
	ClientServiceLink paymentDigestLink
	PaymentLink       paymentDigestLink
}

type paymentDigestLink struct {
	Label           string
	Href            string
	URL             string
	ClientID        string
	ClientServiceID string
	PaymentID       string
	PaymentActionID string
}

var paymentDigestNow = func() time.Time {
	return time.Now().UTC()
}
