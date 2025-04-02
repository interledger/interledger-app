package grpc

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) ListFeatures(ctx context.Context, _ *pb.Empty) (*pb.Features, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	feat, err := s.b.Features().Features(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Features{
		SendEnabled:              feat.SendEnabled,
		ReceiveEnabled:           feat.ReceiveEnabled,
		LinkedAccountsEnabled:    feat.LinkedAccEnabled,
		CardsEnabled:             feat.CardsEnabled,
		BanksEnabled:             feat.BanksEnabled,
		IdentitiesEnabled:        feat.IdentitiesEnabled,
		TwitterEnabled:           feat.TwitterEnabled,
		AddCardsEnabled:          feat.AddCardsEnabled,
		InteracEnabled:           feat.InteraccEnabled,
		ManageWalletCardsEnabled: feat.ManageWalletCardsEnabled,
	}, nil
}
