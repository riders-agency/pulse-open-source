package repository

import (
	"errors"

	dbpkg "riders.agency/pulse/api/internal/platform/db"
)

var (
	ErrBadInput       = errors.New("bad_input")
	ErrNotFound       = errors.New("not_found")
	ErrVisitorOffline = errors.New("visitor_offline")
)

type SiteChat struct {
	P dbpkg.DB
}
