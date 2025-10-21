package external

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Options func(*client) error

func WithBaseURL(baseURL string) Options {
	return func(c *client) error {
		if baseURL == "" {
			return fmt.Errorf("%w: missing BaseURL", ErrInternal)
		}

		if _, err := url.Parse(baseURL); err != nil {
			return fmt.Errorf("%w: invalid BaseURL", ErrInternal)
		}

		c.baseURL = baseURL
		return nil
	}
}

func WithOTELLHTTPClient() Options {
	return func(c *client) error {
		c.http = otelhttp.DefaultClient
		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) Options {
	return func(c *client) error {
		if httpClient == nil {
			return fmt.Errorf("%w: missing HTTPClient", ErrInternal)
		}
		c.http = httpClient
		return nil
	}
}

func WithClientID(clientID string) Options {
	return func(c *client) error {
		if clientID == "" {
			return fmt.Errorf("%w: missing ClientID", ErrInternal)
		}

		if err := uuid.Validate(clientID); err != nil {
			return fmt.Errorf("%w: invalid ClientID not uuid", ErrInternal)
		}

		c.clientID = clientID
		return nil
	}
}

func WithDerivedKeys(privateKey jwk.Key) Options {
	return func(c *client) error {
		if privateKey == nil {
			return fmt.Errorf("%w: missing PrivateKey", ErrInternal)
		}

		// Remove `kid` otherwise lestrat-jws will not let us override the field in the protected header.
		err := privateKey.Remove("kid")
		if err != nil {
			return fmt.Errorf("%w Failed to remove `kid` field from jwk %s", ErrInternal, err)
		}
		c.privateKey = privateKey

		publicKey, err := c.privateKey.PublicKey()

		if err != nil {
			return fmt.Errorf("%w: cannot derive PublicKey from PrivateKey", ErrInternal)
		}
		c.publicKey = publicKey

		publicKeyThumbprint, err := publicKey.Thumbprint(crypto.SHA256)
		if err != nil {
			return fmt.Errorf("%w: cannot compute PublicKey thumbprint", ErrInternal)
		}

		encodedThumb := base64.RawURLEncoding.EncodeToString(publicKeyThumbprint)

		c.publicKeyThumbprint = encodedThumb

		return nil
	}
}

func NewWithOptions(opts ...Options) (Client, error) {
	client := &client{}
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}
