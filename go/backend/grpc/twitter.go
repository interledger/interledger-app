package grpc

import (
	"context"
	"errors"
	"gitlab.com/fynbos/backend/identities"

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

	wallet, err := s.b.Users().WalletForContext(ctx)
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
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, err = s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

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

	// Publishing the proof
	proofUrl, err := s.b.Twitter().PublishTweetProof(ctx, identity, connection)
	if err != nil {
		return nil, InternalError("Error publishing tweet proof")
	}

	// Verification
	_, err = s.b.Identities().StartVerification(ctx, identity.ID, proofUrl)
	if err != nil {
		return nil, InternalError("Error starting verification")
	}

	return &backendv1.TwitterCallbackResponse{
		Id: identity.ID,
	}, nil
}
