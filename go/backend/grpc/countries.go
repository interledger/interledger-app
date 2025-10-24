package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/country"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetCountries(
	ctx context.Context, request *backendv1.Empty,
) (*backendv1.GetCountriesResponse, error) {
	return &backendv1.GetCountriesResponse{
		Countries: country.GetGRPCCountries(),
	}, nil
}
