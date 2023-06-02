package grpc

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/twitter"
	"gitlab.com/fynbos/log"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"go.uber.org/zap"
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
	connection, err := s.b.Twitter().CreateConnection(ctx, &twitter.CreateConnectionArgs{
		State:    request.State,
		AuthCode: request.Code,
	})
	if err != nil {
		log.Error("creating twitter connection", zap.Error(err))
		return nil, InternalError("Create Twitter Connection")
	}

	identity, err := s.b.Identities().Add(ctx, identities.AddArgs{
		WalletID:   connection.WalletID,
		Platform:   identities.PlatformTwitter,
		Identifier: connection.Username,
	})
	if err != nil {
		log.Error("adding identity", zap.Error(err))
		if errors.Is(err, identities.ErrAlreadyExists) {
			return nil, AlreadyExistsError("Identity already exists.")
		}
		return nil, InternalError("Error adding identity.")
	}

	return &backendv1.TwitterCallbackResponse{
		Id: identity.ID,
	}, nil
}

func (s *rpcService) VerifyTwitter(ctx context.Context, request *backendv1.VerifyTwitterRequest) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	id, err := s.b.Identities().Get(ctx, request.GetIdentityId())
	if err != nil {
		return nil, InternalError("error getting identity")
	}

	if id.WalletID != wallet.ID {
		return nil, NotFoundError("unknown identity")
	}

	if id.State == identities.StateVerified {
		return &backendv1.Empty{}, nil
	}

	err = s.b.Twitter().PublishTweetProof(ctx, id.ID)
	if err != nil {
		return nil, InternalError("error starting workflow")
	}

	return &backendv1.Empty{}, nil
}
