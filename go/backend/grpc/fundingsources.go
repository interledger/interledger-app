package grpc

import (
	"context"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetFundingsources(
	ctx context.Context,
	req *backendv1.GetFundingsourcesRequest,
) (*backendv1.GetFundingsourcesResponse, error) {
	user, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}
	wallet, err := s.b.Users().GetWallet(ctx, user.ID, req.GetWalletId())
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	fundingsources, err := s.b.FundingSources().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, InternalError("Unable to get fundingsources.")
	}

	ret := make([]*backendv1.FundingSource, len(fundingsources))
	for i, fs := range fundingsources {
		ret[i] = &backendv1.FundingSource{
			Id:   fs.ID,
			Name: fs.Name,
			Mask: fs.Mask,
		}
	}

	return &backendv1.GetFundingsourcesResponse{
		Fundingsources: ret,
	}, nil
}
