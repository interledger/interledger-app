package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/discord"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/proto/backend/v1"
)

func (r *rpcService) CreateDiscordAuthURL(ctx context.Context, req *backend.Empty) (*backend.CreateDiscordAuthURLResponse, error) {
	_, err := r.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := r.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	url, err := r.b.Discord().CreateAuthURL(ctx, discord.CreateAuthURLArgs{
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

func (r *rpcService) DiscordCallback(ctx context.Context, req *backend.DiscordCallbackRequest) (*backend.DiscordCallbackResponse, error) {
	connection, err := r.b.Discord().CreateConnection(ctx, discord.CreateConnectionArgs{
		AuthCode: req.GetCode(),
		State:    req.GetState(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	id, err := r.b.Identities().Add(ctx, identities.AddArgs{
		WalletID:   connection.WalletID,
		Platform:   identities.PlatformDiscord,
		Identifier: connection.Username,
	})

	if err != nil {
		if errors.Is(err, identities.ErrAlreadyExists) {
			return nil, AlreadyExistsError("Identity already exists.")
		}
		return nil, InternalError("Error adding identity.")
	}

	err = r.b.Identities().UpdateState(ctx, id.ID, identities.StateVerified, "")
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backend.DiscordCallbackResponse{
		Id: id.ID,
	}, nil
}
