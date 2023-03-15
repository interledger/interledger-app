package authorisation

import (
	"context"
	"net/http"
)

type InternalClient interface {
	AddPublicKey(ctx context.Context, clientURL string, key Jwk) (*Jwk, error)
	Introspect(ctx context.Context, token string) (*Grant, error)
	ListKeys(ctx context.Context, clientURl string) ([]Jwk, error)
	GetPublicKeyByID(ctx context.Context, id string) (*Jwk, error)
	DeletePublicKey(ctx context.Context, id string) error
	VerifyRequestSig(ctx context.Context, req *http.Request, clientPaymentPointer string, requiredParts []string) bool
	LookupClient(ctx context.Context, url string) (*Client, error)
}
