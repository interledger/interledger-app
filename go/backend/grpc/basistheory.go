package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) CreateCard(
	ctx context.Context, req *backend.CreateCardRequest,
) (*backend.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	card, err := s.b.BasisTheory().CreateCard(ctx, req.GetTokenID(), w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	_, err = s.b.Tabapay().CreateCard(ctx, tabapay.CreateCardArgs{
		WalletID:          w.ID,
		BasisTheoryCardID: card.ID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backend.Empty{}, nil
}
