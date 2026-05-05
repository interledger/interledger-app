package external

import "errors"

var (
	ErrInvalidURL      = errors.New("xago: invalid base URL")
	ErrMissingClientID = errors.New("xago: missing client ID")

	ErrMissingHTTPClient = errors.New("fiant: missing HTTP client")

	ErrUnexpectedStatusCode = errors.New("xago: unexpected status code")
	ErrDecodingResponse     = errors.New("xago: error decoding response body")
)
