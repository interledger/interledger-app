package client

import (
	"context"
	"net/http"

	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/authorisation/ops"
)

var _ authorisation.InternalClient = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) authorisation.InternalClient {
	return &client{
		b: b,
	}
}

func (c client) AddPublicKey(ctx context.Context, clientURL string, publicKey authorisation.Jwk) error {
	return ops.CreateClientPublicKey(ctx, c.b, clientURL, publicKey)
}
func (c client) Introspect(ctx context.Context, token string) (*authorisation.Grant, error) {
	return ops.Introspect(ctx, c.b, token)
}

func (c client) ListKeys(ctx context.Context, clientURL string) ([]authorisation.Jwk, error) {
	return ops.ListKeys(ctx, c.b, clientURL)
}

func (c client) VerifyRequestSig(ctx context.Context, req *http.Request, clientPaymentPointer string, requiredParts []string) bool {
	return ops.VerifyRequestSig(ctx, req, clientPaymentPointer, requiredParts)
}
