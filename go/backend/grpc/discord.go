package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/discord"
	"gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateDiscordAuthURL(ctx context.Context, req *backend.Empty) (*backend.CreateDiscordAuthURLResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	url, err := s.b.Discord().CreateAuthURL(ctx, discord.CreateAuthURLArgs{
		Scopes:   []string{"identify", "guilds"},
		WalletID: wallet.ID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backend.CreateDiscordAuthURLResponse{
		Url: url,
	}, nil
}
