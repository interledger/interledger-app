package external

import "errors"

var (
	ErrInternal               = errors.New("tabapay external: internal error.")
	ErrMultiStatus            = errors.New("tabapay external: multi status.")
	ErrBadRequest             = errors.New("tabapay external: bad request.")
	ErrUnauthorized           = errors.New("tabapay external: unauthorized.")
	ErrForbidden              = errors.New("tabapay external: forbidden.")
	ErrNotFound               = errors.New("tabapay external: not found.")
	ErrMethodNotAllowed       = errors.New("tabapay external: method not allowed.")
	ErrNotAcceptable          = errors.New("tabapay external: method not acceptable.")
	ErrConflict               = errors.New("tabapay external: conflict.")
	ErrGone                   = errors.New("tabapay external: gone.")
	ErrUnsupportedMediatype   = errors.New("tabapay external: unsupported media type.")
	ErrMisdirectedRequest     = errors.New("tabapay external: misdirected request.")
	ErrUnprocessableEntity    = errors.New("tabapay external: unprocessable entity.")
	ErrLocked                 = errors.New("tabapay external: locked.")
	ErrTooManyRequests        = errors.New("tabapay external: too many requests.")
	ErrRequestHeadersTooLarge = errors.New("tabapay external: request headers too large.")
	ErrServer                 = errors.New("tabapay external: server error.")
	ErrBadGateway             = errors.New("tabapay external: bad gateway.")
	ErrServiceUnavailable     = errors.New("tabapay external: server unavailable.")
	ErrGatewayTimeout         = errors.New("tabapay external: gateway timeout.")
)
