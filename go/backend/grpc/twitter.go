package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/identities"
	"github.com/interledger/interledger-app/go/backend/twitter"
	backendv1 "github.com/interledger/interledger-app/go/proto/backend/v1"
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

func (s *rpcService) TwitterCallback(
	ctx context.Context, request *backendv1.TwitterCallbackRequest,
) (*backendv1.TwitterCallbackResponse, error) {
	connection, err := s.b.Twitter().CreateConnection(ctx, &twitter.CreateConnectionArgs{
		State:    request.State,
		AuthCode: request.Code,
	})
	if err != nil {
		return nil, InternalError("Create Twitter Connection")
	}

	identity, err := s.b.Identities().Add(ctx, identities.AddArgs{
		WalletID:   connection.WalletID,
		Platform:   identities.PlatformTwitter,
		Identifier: connection.Username,
	})
	if err != nil {
		if errors.Is(err, identities.ErrAlreadyExists) {
			return nil, AlreadyExistsError("Identity already exists.")
		}
		return nil, InternalError("Error adding identity.")
	}

	return &backendv1.TwitterCallbackResponse{
		Id: identity.ID,
	}, nil
}
