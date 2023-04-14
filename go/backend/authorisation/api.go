package authorisation

import (
	"context"
	"net/http"
)

type InternalClient interface {
	Introspect(ctx context.Context, token string) (*Grant, error)
	VerifyRequestSig(ctx context.Context, req *http.Request, clientPaymentPointer string, requiredParts []string) bool
	LookupClient(ctx context.Context, url string) (*Client, error)
}
