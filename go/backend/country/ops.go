package country

import (
	"fmt"
	"strings"
)

// This will return the ISO3166_2 code for the state. e.g. US-CA.
func GetStateCode(country Country, state string) (string, error) {
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
