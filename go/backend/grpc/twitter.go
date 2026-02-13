package grpc

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/twitter"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateTwitterAuthURL(
	ctx context.Context,
	_ *backendv1.Empty,
) (*backendv1.CreateTwitterAuthURLResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	state := uuid.NewString()
	url, err := s.b.Twitter().CreateAuthURL(ctx, &twitter.CreateAuthURLArgs{
		State:    state,
		Scopes:   []string{"offline.access", "tweet.read", "tweet.write", "users.read"},
		WalletID: wallet.ID,
	})
	if err != nil {
		return nil, InternalError("Create Twitter Auth URL")
	}

	return &backendv1.CreateTwitterAuthURLResponse{
		Url: url,
	}, nil
}
