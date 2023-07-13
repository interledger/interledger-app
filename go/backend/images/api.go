package images

import "context"

type Client interface {
	GenerateTwitterIdentity(ctx context.Context, walletUrl, identifier string) ([]byte, error)
	GenerateWebsiteIdentity(ctx context.Context, walletUrl, identifier string) ([]byte, error)
	GenerateTwitterIdentityOG(ctx context.Context, walletUrl, identifier string) ([]byte, error)
}
