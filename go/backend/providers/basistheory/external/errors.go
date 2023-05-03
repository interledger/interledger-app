package external

import "errors"

var (
	ErrInternal               = errors.New("basistheory external: internal error.")
	ErrMultiStatus            = errors.New("basistheory external: multi status.")
	ErrBadRequest             = errors.New("basistheory external: bad request.")
	ErrUnauthorized           = errors.New("basistheory external: unauthorized.")
	ErrForbidden              = errors.New("basistheory external: forbidden.")
	ErrNotFound               = errors.New("basistheory external: not found.")
	ErrMethodNotAllowed       = errors.New("basistheory external: method not allowed.")
	ErrNotAcceptable          = errors.New("basistheory external: method not acceptable.")
	ErrConflict               = errors.New("basistheory external: conflict.")
	ErrGone                   = errors.New("basistheory external: gone.")
	ErrUnsupportedMediatype   = errors.New("basistheory external: unsupported media type.")
	ErrMisdirectedRequest     = errors.New("basistheory external: misdirected request.")
	ErrUnprocessableEntity    = errors.New("basistheory external: unprocessable entity.")
	ErrLocked                 = errors.New("basistheory external: locked.")
	ErrTooManyRequests        = errors.New("basistheory external: too many requests.")
	ErrRequestHeadersTooLarge = errors.New("basistheory external: request headers too large.")
	ErrServer                 = errors.New("basistheory external: server error.")
	ErrBadGateway             = errors.New("basistheory external: bad gateway.")
	ErrServiceUnavailable     = errors.New("basistheory external: server unavailable.")
	ErrGatewayTimeout         = errors.New("basistheory external: gateway timeout.")
)
