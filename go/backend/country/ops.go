package country

import (
	"fmt"
	"sort"
	"strings"

	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

// This will return the ISO3166_2 code for the state. e.g. US-CA.
func GetStateCode(country Country, state string) (string, error) {
	_, ok := States[country][strings.TrimSpace(state)] // check for direct match
	if ok {
		return fmt.Sprintf("%s-%s", country, strings.TrimSpace(state)), nil
	}

	states, ok := States[country]
	if !ok {
		return "", fmt.Errorf("%w Unknown country=%s.", ErrNotFound, country)
	}

	for code, stateName := range states {
		if strings.EqualFold(strings.TrimSpace(state), stateName) {
			return fmt.Sprintf("%s-%s", country, code), nil
		}
	}

	return "", fmt.Errorf("%w Unknown state=%s for country=%s.", ErrNotFound, state, country)
}

func GetGRPCCountries() []*backendv1.Country {
	var ret []*backendv1.Country
	for c, details := range Details {
		ret = append(ret, &backendv1.Country{Id: c.String(), Name: details.Name})
	}

	var us, za, ca *backendv1.Country
	var filteredRet []*backendv1.Country
	for _, c := range ret {
		if c.Id == US.String() {
			us = c
		} else if c.Id == ZA.String() {
			za = c
		} else if c.Id == CA.String() {
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

	return finalRet
}
