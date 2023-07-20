package grpc

import (
	"context"
	"sort"

	"gitlab.com/fynbos/backend/country"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetCountries(
	ctx context.Context, request *backendv1.Empty,
) (*backendv1.GetCountriesResponse, error) {

	var ret []*backendv1.Country
	for c, details := range country.Details {
		ret = append(ret, &backendv1.Country{Id: c.String(), Name: details.Name})
	}

	sort.Slice(ret, func(i, j int) bool {
		if country.Country(ret[i].Id) == country.US {
			return true
		}
		if country.Country(ret[j].Id) == country.US {
			return false
		}
		return ret[i].Name < ret[j].Name
	})

	return &backendv1.GetCountriesResponse{
		Countries: ret,
	}, nil
}
