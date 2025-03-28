package admin

import (
	"context"

	"gitlab.com/fynbos/backend/features"

	pb "gitlab.com/fynbos/proto/backend/admin/v1"
)

func (s *AdminRpcService) GetWalletFeatures(ctx context.Context, req *pb.GetWalletFeaturesRequest) (*pb.Features, error) {
	feat, err := s.b.Features().Features(ctx, req.WalletID)
	if err != nil {
		return nil, err
	}

	return &pb.Features{
		SendEnabled:           feat.SendEnabled,
		ReceiveEnabled:        feat.ReceiveEnabled,
		LinkedAccountsEnabled: feat.LinkedAccEnabled,
		CardsEnabled:          feat.CardsEnabled,
		BanksEnabled:          feat.BanksEnabled,
		IdentitiesEnabled:     feat.IdentitiesEnabled,
		TwitterEnabled:        feat.TwitterEnabled,
		AddCardsEnabled:       feat.AddCardsEnabled,
		ManageCardsEnabled:    feat.ManageCardsEnabled,
		WalletID:              req.WalletID,
	}, nil
}

func (s *AdminRpcService) SetWalletFeatures(ctx context.Context, req *pb.Features) (*pb.Features, error) {
	feat, err := s.b.Features().SetFeatures(ctx, req.WalletID, features.WalletFeatures{
		SendEnabled:        req.SendEnabled,
		ReceiveEnabled:     req.ReceiveEnabled,
		LinkedAccEnabled:   req.LinkedAccountsEnabled,
		CardsEnabled:       req.CardsEnabled,
		BanksEnabled:       req.BanksEnabled,
		IdentitiesEnabled:  req.IdentitiesEnabled,
		TwitterEnabled:     req.TwitterEnabled,
		AddCardsEnabled:    req.AddCardsEnabled,
		ManageCardsEnabled: req.ManageCardsEnabled,
	})
	if err != nil {
		return nil, err
	}

	return &pb.Features{
		SendEnabled:           feat.SendEnabled,
		ReceiveEnabled:        feat.ReceiveEnabled,
		LinkedAccountsEnabled: feat.LinkedAccEnabled,
		CardsEnabled:          feat.CardsEnabled,
		BanksEnabled:          feat.BanksEnabled,
		IdentitiesEnabled:     feat.IdentitiesEnabled,
		TwitterEnabled:        feat.TwitterEnabled,
		AddCardsEnabled:       feat.AddCardsEnabled,
		ManageCardsEnabled:    feat.ManageCardsEnabled,
		WalletID:              req.WalletID,
	}, nil
}
