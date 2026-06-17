package grpc

import (
	"context"

	"github.com/interledger/interledger-app/go/backend/country"
	backendv1 "github.com/interledger/interledger-app/go/proto/backend/v1"
)

func (s *rpcService) GetCountries(
	ctx context.Context, request *backendv1.Empty,
) (*backendv1.GetCountriesResponse, error) {
	return &backendv1.GetCountriesResponse{
		Countries: country.GetGRPCCountries(),
	}, nil
}
