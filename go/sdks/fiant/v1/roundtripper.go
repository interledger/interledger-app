package v1

import (
	"time"

	"net/http"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

type apiRoundTripper struct {
	url string
	// clientID string

	privateKey jwk.Key
	//publicKey jwk.Key // for later use

	publicKeyThumbprint string

	headers map[string]string

	defaultTransport http.RoundTripper
}

func (art *apiRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))

	for k, v := range art.headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}

	signature, err := art.signature(req)
	if err != nil {
		return nil, err
	}

	req.Header.Add(signatureHeader, signature)

	return art.defaultTransport.RoundTrip(req)
}
