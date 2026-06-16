package grpc

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/rafiki"

	"github.com/interledger/interledger-app/go/backend/db"

	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
)

func (s *rpcService) ListRafikiGrants(ctx context.Context, _ *pb.Empty) (*pb.ListRafikiGrantsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	gl, err := s.b.Rafiki().ListGrants(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := make([]*pb.RafikiGrant, len(gl))
	for i, g := range gl {
		resp[i] = transformRafikiGrant(g)
	}

	return &pb.ListRafikiGrantsResponse{Grants: resp}, nil
}

func transformRafikiGrant(g rafiki.Grant) *pb.RafikiGrant {
	access := make([]*pb.RafikiAccess, len(g.Access))
	for i, a := range g.Access {
		access[i] = &pb.RafikiAccess{
			Id:         a.ID,
			Identifier: a.Identifier,
			Type:       a.Type,
			Actions:    a.Actions,
			Limits: &pb.RafikiLimits{
				Receiver:              a.Limits.Receiver,
				Interval:              a.Limits.Interval,
				DebitAmount:           a.Limits.DebitAmount.ToPB(),
				ReceiveAmount:         a.Limits.ReceiveAmount.ToPB(),
				FormattedDebitAmount:  a.Limits.DebitAmount.Format(),
				FormattdReceiveAmount: a.Limits.ReceiveAmount.Format(),
			},
		}
	}
	return &pb.RafikiGrant{
		Id:                 g.Id,
		Client:             g.Client,
		State:              g.State,
		FinalizationReason: g.FinalizationReason,
		CreatedAt:          g.CreatedAt,
		Access:             access,
	}
}

func (s *rpcService) GetRafikiGrant(ctx context.Context, req *pb.GetRafikiGrantRequest) (*pb.RafikiGrant, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, err = s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	g, err := s.b.Rafiki().GetGrant(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformRafikiGrant(*g), nil
}

func (s *rpcService) RevokeRafikiGrant(ctx context.Context, req *pb.RevokeRafikiGrantRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	_, err = s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	err = s.b.Rafiki().RevokeGrant(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.Empty{}, nil
}

func (s *rpcService) ListPendingWebMonetization(ctx context.Context, _ *pb.Empty) (*pb.ListTransactionsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	txs, err := s.b.Rafiki().ListPendingTransactions(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformTransactions(txs, db.Pagination{PageSize: len(txs)}), nil
}
