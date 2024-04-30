package external

import "errors"

var (
	ErrInternal               = errors.New("gatehub external: internal error.")
	ErrMultiStatus            = errors.New("gatehub external: multi status.")
	ErrBadRequest             = errors.New("gatehub external: bad request.")
	ErrUnauthorized           = errors.New("gatehub external: unauthorized.")
	ErrForbidden              = errors.New("gatehub external: forbidden.")
	ErrNotFound               = errors.New("gatehub external: not found.")
	ErrMethodNotAllowed       = errors.New("gatehub external: method not allowed.")
	ErrNotAcceptable          = errors.New("gatehub external: method not acceptable.")
	ErrConflict               = errors.New("gatehub external: conflict.")
	ErrGone                   = errors.New("gatehub external: gone.")
	ErrUnsupportedMediatype   = errors.New("gatehub external: unsupported media type.")
	ErrMisdirectedRequest     = errors.New("gatehub external: misdirected request.")
	ErrUnprocessableEntity    = errors.New("gatehub external: unprocessable entity.")
	ErrLocked                 = errors.New("gatehub external: locked.")
	ErrTooManyRequests        = errors.New("gatehub external: too many requests.")
	ErrRequestHeadersTooLarge = errors.New("gatehub external: request headers too large.")
	ErrServer                 = errors.New("gatehub external: server error.")
	ErrBadGateway             = errors.New("gatehub external: bad gateway.")
	ErrServiceUnavailable     = errors.New("gatehub external: server unavailable.")
	ErrGatewayTimeout         = errors.New("gatehub external: gateway timeout.")
	ErrInvalidSignature       = errors.New("gatehub external: invalid signature.")
)
