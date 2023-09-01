package images

import "context"

type Client interface {
	GenerateTwitterIdentity(ctx context.Context, walletUrl, identifier string) ([]byte, error)
	GenerateTwitterIdentityOG(ctx context.Context, walletUrl, identifier string) ([]byte, error)
	GenerateDomainIdentity(ctx context.Context, walletUrl, identifier string) ([]byte, error)
	GenerateDomainIdentityOG(ctx context.Context, walletUrl, identifier string) ([]byte, error)
	GenerateDiscordIdentity(ctx context.Context, walletUrl, identifier string) ([]byte, error)
	GenerateDiscordIdentityOG(ctx context.Context, walletUrl, identifier string) ([]byte, error)
}
