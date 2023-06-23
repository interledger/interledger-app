package grpc

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/linkedin"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateLinkedinAuthURL(
	ctx context.Context,
	_ *backendv1.Empty,
) (*backendv1.CreateLinkedinAuthURLResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	state := uuid.NewString()
	url, err := s.b.Linkedin().CreateAuthURL(ctx, &linkedin.CreateAuthURLArgs{
		State:    state,
		Scopes:   []string{"w_member_social"},
		WalletID: wallet.ID,
	})
	if err != nil {
		return nil, InternalError("Create Linkedin Auth URL")
	}

	return &backendv1.CreateLinkedinAuthURLResponse{
		Url: url,
	}, nil
}

func (s *rpcService) LinkedinCallback(
	ctx context.Context, request *backendv1.LinkedinCallbackRequest,
) (*backendv1.LinkedinCallbackResponse, error) {
	connection, err := s.b.Linkedin().CreateConnection(ctx, &linkedin.CreateConnectionArgs{
		State:    request.State,
		AuthCode: request.Code,
	})
	if err != nil {
		return nil, InternalError("Create Linkedin Connection")
	}

	identity, err := s.b.Identities().Add(ctx, identities.AddArgs{
		WalletID:   connection.WalletID,
		Platform:   identities.PlatformLinkedin,
		Identifier: connection.Username,
	})
	if err != nil {
		if errors.Is(err, identities.ErrAlreadyExists) {
			return nil, AlreadyExistsError("Identity already exists.")
		}
		return nil, InternalError("Error adding identity.")
	}

	return &backendv1.LinkedinCallbackResponse{
		Id: identity.ID,
	}, nil
}

func (s *rpcService) VerifyLinkedin(ctx context.Context, request *backendv1.VerifyLinkedinRequest) (*backendv1.Empty, error) {
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

	err = s.b.Linkedin().PublishPublicProof(ctx, id.ID)
	if err != nil {
		return nil, InternalError("error starting workflow")
	}

	return &backendv1.Empty{}, nil
}
