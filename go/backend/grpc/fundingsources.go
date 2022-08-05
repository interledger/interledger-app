package grpc

import (
	"context"

	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetFundingsources(
	ctx context.Context,
	req *backendv1.Empty,
) (*backendv1.GetFundingsourcesResponse, error) {
	user, err := s.userService.ForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	acc, err := s.accountsService.GetByIdentityID(ctx, user.ID)
	if err != nil {
		return nil, InternalError("Unable to get account.")
	}

	fundingsources, err := s.fundingSourceService.GetByAccountId(ctx, acc.IdentityID)
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
