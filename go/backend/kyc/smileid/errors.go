package smileid

import "errors"

var (
	ErrInternal            = errors.New("smileid external: internal error")
	ErrBadRequest          = errors.New("smileid external: bad request")
	ErrUnauthorized        = errors.New("smileid external: unauthorized")
	ErrForbidden           = errors.New("smileid external: forbidden")
	ErrNotFound            = errors.New("smileid external: not found")
	ErrMethodNotAllowed    = errors.New("smileid external: method not allowed")
	ErrNotAcceptable       = errors.New("smileid external: method not acceptable")
	ErrConflict            = errors.New("smileid external: conflict")
	ErrUnprocessableEntity = errors.New("smileid external: unprocessable entity")
	ErrServer              = errors.New("smileid external: server error")
	ErrServiceUnavailable  = errors.New("smileid external: server unavailable")
)
