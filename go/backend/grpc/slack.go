package grpc

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/slack"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateSlackAuthURL(ctx context.Context, _ *pb.Empty) (*pb.CreateSlackAuthURLResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	url, err := s.b.Slack().CreateAuthURL(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateSlackAuthURLResponse{Url: url}, nil
}

func (s *rpcService) SlackCallback(ctx context.Context, req *pb.SlackCallbackRequest) (*pb.SlackCallbackResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, err = s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	con, err := s.b.Slack().CreateConnection(ctx, slack.CreateConnectionArgs{
		AuthCode: req.Code,
		State:    req.State,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	id, err := s.b.Identities().Add(ctx, identities.AddArgs{
		WalletID:   con.WalletID,
		Platform:   identities.PlatformSlack,
		Identifier: fmt.Sprintf("%s / %s", con.TeamName, con.Username),
	})

	if err != nil {
		if errors.Is(err, identities.ErrAlreadyExists) {
			return nil, AlreadyExistsError("Identity already exists.")
		}
		return nil, InternalError("Error adding identity.")
	}

	err = s.b.Identities().UpdateState(ctx, id.ID, identities.StateVerified, "")
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.SlackCallbackResponse{
		Id: id.ID,
	}, nil
}
