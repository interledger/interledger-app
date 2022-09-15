package agreements

import "context"

type Client interface {
	Sign(ctx context.Context, args *SignArgs) error
	GetSignatures(ctx context.Context, identityID string) ([]Signature, error)
	Get(ctx context.Context, id string) (*Agreement, error)
}
