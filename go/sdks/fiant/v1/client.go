package v1

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
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

// Client is the main entry point for interacting with the Fiant API
type Client struct {
	url      string
	clientID string

	privateKey jwk.Key
	//publicKey  jwk.Key // for later use

	// Thumbprint is used as the `kid` field in the jwt protected header
	publicKeyThumbprint string

	http *http.Client

	// services
	AuthService         *authService
	UsersService        *usersService
	TransactionsService *transactionsService
}

type ClientOptions func(*Client) error

func WithBaseURL(providerURL string) ClientOptions {
	return func(client *Client) error {
		if !slices.Contains(allowedURLs, providerURL) {
			return ErrInvalidURL
		}

		client.url = providerURL
		return nil
	}
}

func WithClientID(clientID string) ClientOptions {
	return func(client *Client) error {
		if clientID == "" {
			return ErrMissingClientID
		}

		client.clientID = clientID
		return nil
	}
}

func WithDerivedKeys(privateKey jwk.Key) ClientOptions {
	return func(client *Client) error {
		if privateKey == nil {
			return ErrMissingPrivateKey
		}

		// Remove `kid` otherwise lestrat-jws will not let us override the field in the protected header.
		if err := privateKey.Remove("kid"); err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToRemoveKid, err)
		}
		client.privateKey = privateKey

		publicKey, err := client.privateKey.PublicKey()
		if err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToDerivePublicKey, err)

		}
		// client.publicKey = publicKey

		publicKeyThumbprint, err := publicKey.Thumbprint(crypto.SHA256)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrFailedToDerivePublicKey, err)
		}

		client.publicKeyThumbprint = base64.RawURLEncoding.EncodeToString(publicKeyThumbprint)

		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) ClientOptions {
	return func(client *Client) error {
		if httpClient == nil {
			return ErrMissingHTTPClient
		}
		client.http = httpClient
		return nil
	}
}

func NewClient(opts ...ClientOptions) (*Client, error) {
	client := &Client{}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	client.http = &http.Client{
		Transport: &apiRoundTripper{
			url:                 client.url,
			privateKey:          client.privateKey,
			publicKeyThumbprint: client.publicKeyThumbprint,
			defaultTransport:    http.DefaultTransport,
			headers: map[string]string{
				acceptHeader:      acceptValue,
				contentTypeHeader: contentTypeValue,
				clientIDHeader:    client.clientID,
			},
		},
	}

	// services
	client.AuthService = &authService{client: client}
	client.UsersService = &usersService{client: client}
	client.TransactionsService = &transactionsService{client: client}

	return client, nil
}

type requestOption func(*http.Request)

func withHeader(key, value string) requestOption {
	return func(r *http.Request) { r.Header.Set(key, value) }
}

func (client *Client) get(ctx context.Context, path string, opts ...requestOption) (*http.Response, error) {
	endpoint, err := url.JoinPath(client.url, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// the method expects that payload has
// MarshallJSON/UnmarshalJSON implemented if it is a struct, otherwise it will be marshaled as is
// see any dto struct in the domain/dto package for an example of how to implement these methods
func (client *Client) post(ctx context.Context, path string, payload any, opts ...requestOption) (*http.Response, error) {
	endpoint, err := url.JoinPath(client.url, path)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
