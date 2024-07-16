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

	var us, za, ca *backendv1.Country
	var filteredRet []*backendv1.Country
	for _, c := range ret {
		if c.Id == country.US.String() {
			us = c
		} else if c.Id == country.ZA.String() {
			za = c
		} else if c.Id == country.CA.String() {
			ca = c
		} else {
			filteredRet = append(filteredRet, c)
		}
	}

	// Sort the remaining countries
	sort.Slice(filteredRet, func(i, j int) bool {
		return filteredRet[i].Name < filteredRet[j].Name
	})

	// Prepend US, CA and ZA in deterministic order
	finalRet := []*backendv1.Country{us, za, ca}
	finalRet = append(finalRet, filteredRet...)

	return &backendv1.GetCountriesResponse{
		Countries: finalRet,
	}, nil
}
