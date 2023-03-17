package external

import "errors"

var (
	ErrInternal            = errors.New("mx external: internal error")
	ErrBadRequest          = errors.New("mx external: bad request")
	ErrUnauthorized        = errors.New("mx external: unauthorized")
	ErrForbidden           = errors.New("mx external: forbidden")
	ErrNotFound            = errors.New("mx external: not found")
	ErrMethodNotAllowed    = errors.New("mx external: method not allowed")
	ErrNotAcceptable       = errors.New("mx external: method not acceptable")
	ErrConflict            = errors.New("mx external: conflict")
	ErrUnprocessableEntity = errors.New("mx external: unprocessable entity")
	ErrServer              = errors.New("mx external: server error")
	ErrServiceUnavailable  = errors.New("mx external: server unavailable")
)
