package grpc

import (
	"context"

	pb "gitlab.com/fynbos/proto/backend/v1"
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
		resp[i] = &pb.RafikiGrant{
			Id:                 g.Id,
			Client:             g.Client,
			State:              g.State,
			FinalizationReason: g.FinalizationReason,
		}
	}

	return &pb.ListRafikiGrantsResponse{Grants: resp}, nil
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

	return &pb.RafikiGrant{
		Id:                 g.Id,
		Client:             g.Client,
		State:              g.State,
		FinalizationReason: g.FinalizationReason,
	}, nil
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
