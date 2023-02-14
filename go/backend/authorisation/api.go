package authorisation

import "context"

type InternalClient interface {
	AddPublicKey(ctx context.Context, clientURL string, key Jwk) error
	Introspect(ctx context.Context, token string) error
	ListKeys(ctx context.Context, clientURl string) ([]Jwk, error)
	BaseURL() string
}
