package country_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/country"
)

func TestGetStateCode(t *testing.T) {
	lookup := []string{"california", "California", " caliFornia", " CaliforniA ", "california ", "CA", "CA "}

	for _, l := range lookup {
		code, err := country.GetStateCode(country.US, l)
		require.NoError(t, err)
		assert.Equal(t, "US-CA", code)
	}

	code, err := country.GetStateCode(country.US, "w")
	require.Error(t, err)
	assert.Empty(t, code)
}
