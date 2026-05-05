package external

import (
	"fmt"
	"net/http"
)

// type urls struct {
// 	base     string
// 	identity string
// }

type apiRoundTripper struct {
	// url      urls
	loginURL string // requests to this URL skip Bearer injection

	tokenProvider *tokenProvider

	headers          map[string]string
	defaultTransport http.RoundTripper
}

func (art *apiRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())

	for k, v := range art.headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}

	// The login endpoint authenticates via credentials in the body, not a Bearer token.
	if req.URL.String() == art.loginURL {
		fmt.Println("loginURL")
		return art.defaultTransport.RoundTrip(req)
	}

	// Inject a cached token proactively if one is available, avoiding an
	// unnecessary 401 round-trip when the token is still fresh.
	if token := art.tokenProvider.getCached(); token != "" {
		fmt.Println("token", token)
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := art.defaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	// 401: force-refresh the token and retry the request once.
	resp.Body.Close()

	token, err := art.tokenProvider.get(req.Context(), true)
	if err != nil {
		return nil, err
	}

	retryReq := req.Clone(req.Context())
	if req.GetBody != nil {
		retryReq.Body, err = req.GetBody()
		if err != nil {
			return nil, err
		}
	}
	retryReq.Header.Set("Authorization", "Bearer "+token)

	return art.defaultTransport.RoundTrip(retryReq)
}
