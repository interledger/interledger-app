package http

import (
	"net/http"
)

var _ http.RoundTripper = &Transport{}

func NewTransport(base http.RoundTripper, b Backends) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &Transport{
		b:    b,
		base: base,
	}
}

type Transport struct {
	b    Backends
	base http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.base.RoundTrip(req)
	if err == nil {
		Log(t.b, req, res)
	}

	return res, err
}
