package repository

import (
	"errors"

	depscontracts "riders.agency/pulse/api/internal/modules/action_dependencies/contracts"

	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

var (
	ErrInvalidIdentifiers  = errors.New("invalid client identifiers")
	ErrDaDataNotConfigured = depscontracts.ErrDaDataNotConfigured
	ErrPaymentClosed       = errors.New("payment action is already closed")
	ErrPaymentNoService    = errors.New("client service is required to record a payment")
	ErrIdempotencyRequired = errors.New("idempotency_key is required")
)

type Clients struct {
	P         dbpkg.DB
	MasterHEX string
}
