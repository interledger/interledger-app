package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/limits"
	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) ListLimits(ctx context.Context, _ *pb.Empty) (*pb.ListLimitsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	ll, err := s.b.Limits().List(ctx, w.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	res := make([]*pb.ConfiguredLimit, len(ll))
	for i, l := range ll {
		res[i] = &pb.ConfiguredLimit{
			ForeignId:      l.ForeignID,
			ForeignDisplay: l.ForeignDisplay,
			ForeignType:    string(l.ForeignType),
			Daily:          l.Limit.Daily.ToPB(),
			Monthly:        l.Limit.Monthly.ToPB(),
			Overall:        l.Limit.Overall.ToPB(),
		}
	}

	return &pb.ListLimitsResponse{Limits: res}, nil
}

func (s *rpcService) UpdateClientLimits(ctx context.Context, req *pb.UpdateClientLimitsRequest) (*pb.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	err = s.b.Limits().UpdateClientLimits(ctx, w.ID, req.ClientUrl, limits.Limit{
		Daily:   currency.FromPB(req.Daily),
		Monthly: currency.FromPB(req.Monthly),
		Overall: currency.FromPB(req.Overall),
	})

	return &pb.Empty{}, toGRPCError(err)
}
