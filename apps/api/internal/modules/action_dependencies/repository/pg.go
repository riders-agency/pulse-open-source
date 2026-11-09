package repository

import (
	"errors"

	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

// ErrPlaintextRequired is returned when rotating a dependency secret without a new value.
var ErrPlaintextRequired = errors.New("plaintext_secret required")

type Deps struct {
	P         dbpkg.DB
	MasterHEX string
}
