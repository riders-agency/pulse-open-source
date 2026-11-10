package errs

import "errors"

// ErrForbidden indicates the actor cannot access this tenant resource.
var ErrForbidden = errors.New("forbidden")
