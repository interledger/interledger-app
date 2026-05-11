package client_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc/client"
	"gitlab.com/fynbos/backend/kyc/persona"
)

func TestGetPersonaZAIDNumber_FakeMode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// nil backends is safe: the fake path never touches the DB or Persona HTTP client.
	// Dummy Smarty credentials satisfy the prod-env guard without making real API calls.
	kycClient, err := client.NewWithPersonaConfig(nil, "dummy-id", "dummy-token", persona.Config{FakeZAID: true})
	require.NoError(t, err)

	zaIDPattern := regexp.MustCompile(`^\d{13}$`)

	for range 10 {
		idNum, err := kycClient.GetPersonaZAIDNumber(ctx, "any-wallet-id")
		require.NoError(t, err)
		assert.Regexp(t, zaIDPattern, idNum, "expected 13-digit ZA ID number")
	}
}
