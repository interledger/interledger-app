package external

import "errors"

var (
	ErrInternal               = errors.New("pti external: internal error")
	ErrMultiStatus            = errors.New("pti external: multi status")
	ErrBadRequest             = errors.New("pti external: bad request")
	ErrUnauthorized           = errors.New("pti external: unauthorized")
	ErrForbidden              = errors.New("pti external: forbidden")
	ErrNotFound               = errors.New("pti external: not found")
	ErrMethodNotAllowed       = errors.New("pti external: method not allowed")
	ErrNotAcceptable          = errors.New("pti external: method not acceptable")
	ErrConflict               = errors.New("pti external: conflict")
	ErrGone                   = errors.New("pti external: gone")
	ErrUnsupportedMediatype   = errors.New("pti external: unsupported media type")
	ErrMisdirectedRequest     = errors.New("pti external: misdirected request")
	ErrUnprocessableEntity    = errors.New("pti external: unprocessable entity")
	ErrLocked                 = errors.New("pti external: locked")
	ErrTooManyRequests        = errors.New("pti external: too many requests")
	ErrRequestHeadersTooLarge = errors.New("pti external: request headers too large")
	ErrServer                 = errors.New("pti external: server error")
	ErrBadGateway             = errors.New("pti external: bad gateway")
	ErrServiceUnavailable     = errors.New("pti external: server unavailable")
	ErrGatewayTimeout         = errors.New("pti external: gateway timeout")
	ErrInvalidSignature       = errors.New("pti external: invalid signature")
)
