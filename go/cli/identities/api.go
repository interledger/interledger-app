package identities

import "context"

type Client interface {
	Verify(ctx context.Context, args *VerifyArgs) error
}
