package grpc

import (
	"context"

	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetCountries(
	ctx context.Context, request *backendv1.Empty,
) (*backendv1.GetCountriesResponse, error) {
	countries, err := s.b.Countries().GetAll(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}

	ret := make([]*backendv1.Country, len(countries))
	for i, country := range countries {
		ret[i] = &backendv1.Country{Id: country.Alpha_2, Name: country.Name}
	}

	return &backendv1.GetCountriesResponse{
		Countries: ret,
	}, nil
}
