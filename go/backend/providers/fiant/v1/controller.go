package v1

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"fmt"

	"net/http"
	"net/url"
	"slices"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

var allowedURLs = []string{
	"https://api.platform.fiant.io/v1/",
	"https://api.staging.fiant.io/v1/",
}

/*
note(bradu):
some of the fields in the Controller are not definitely needed
but they are included for clarity and future use
*/

// Controller is the main entry point for interacting with the Fiant API
type Controller struct {
	url      string
	clientID string

	privateKey jwk.Key
	//publicKey  jwk.Key // for later use

	// Thumbprint is used as the `kid` field in the jwt protected header
	publicKeyThumbprint string

	http *http.Client

	// handlers
	Users        *userHandler
	Transactions *transactionHandler
}

type Options func(*Controller) error

func WithBaseURL(providerURL string) Options {
	return func(ctrl *Controller) error {
		if !slices.Contains(allowedURLs, providerURL) {
			return ErrInvalidURL
		}

		ctrl.url = providerURL
		return nil
	}
}

func WithClientID(clientID string) Options {
	return func(ctrl *Controller) error {
		if clientID == "" {
			return ErrMissingClientID
		}

		ctrl.clientID = clientID
		return nil
	}
}

func WithDerivedKeys(privateKey jwk.Key) Options {
	return func(ctrl *Controller) error {
		if privateKey == nil {
			return ErrMissingPrivateKey
		}

		// Remove `kid` otherwise lestrat-jws will not let us override the field in the protected header.
		if err := privateKey.Remove("kid"); err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToRemoveKid, err)
		}
		ctrl.privateKey = privateKey

		publicKey, err := ctrl.privateKey.PublicKey()
		if err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToDerivePublicKey, err)

		}
		// ctrl.publicKey = publicKey

		publicKeyThumbprint, err := publicKey.Thumbprint(crypto.SHA256)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToComputeThumbprint, err)
		}

		ctrl.publicKeyThumbprint = base64.RawURLEncoding.EncodeToString(publicKeyThumbprint)

		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) Options {
	return func(ctrl *Controller) error {
		if httpClient == nil {
			return ErrMissingHTTPClient
		}
		ctrl.http = httpClient
		return nil
	}
}

func NewController(opts ...Options) (*Controller, error) {
	ctrl := &Controller{}

	for _, opt := range opts {
		if err := opt(ctrl); err != nil {
			return nil, err
		}
	}

	if ctrl.http == nil {
		ctrl.http = &http.Client{
			Transport: &apiRoundTripper{
				url:                 ctrl.url,
				privateKey:          ctrl.privateKey,
				publicKeyThumbprint: ctrl.publicKeyThumbprint,
				defaultTransport:    http.DefaultTransport,
				headers: map[string]string{
					acceptHeader:      acceptValue,
					contentTypeHeader: contentTypeValue,
					clientIDHeader:    ctrl.clientID,
				},
			},
		}
	}

	// initialize handlers
	ctrl.Users = &userHandler{
		ctrl: ctrl,
		path: "users", //https://api.fiant.io/v1/users
	}

	ctrl.Transactions = &transactionHandler{
		ctrl: ctrl,
		path: "transactions", //https://api.fiant.io/v1/transactions
	}

	return ctrl, nil
}

func (ctrl *Controller) get(ctx context.Context, path string) (*http.Response, error) {
	endpoint, err := url.JoinPath(ctrl.url, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {
		return nil, err
	}

	resp, err := ctrl.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (ctrl *Controller) post(ctx context.Context, path string, payload []byte) (*http.Response, error) {
	endpoint, err := url.JoinPath(ctrl.url, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resp, err := ctrl.http.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
