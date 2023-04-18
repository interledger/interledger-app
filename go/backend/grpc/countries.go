package grpc

import (
	"context"

	"gitlab.com/fynbos/backend/country"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetCountries(
	ctx context.Context, request *backendv1.Empty,
) (*backendv1.GetCountriesResponse, error) {
	var ret []*backendv1.Country
	for country, details := range country.Details {
		ret = append(ret, &backendv1.Country{Id: country.String(), Name: details.Name})
	}

	return &backendv1.GetCountriesResponse{
		Countries: ret,
	}, nil
}
