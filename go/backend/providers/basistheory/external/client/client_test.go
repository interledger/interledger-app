package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/basistheory/external"
	"gitlab.com/fynbos/backend/providers/basistheory/external/client"
	"gotest.tools/assert"
)

func TestClient(t *testing.T) {
	if os.Getenv("BASISTHEORY_API_KEY") == "" {
		t.Skip()
	}

	client := client.New(os.Getenv("BASISTHEORY_API_KEY"))
	token, err := client.GetToken(context.Background(), "e828446a-8b11-4b01-98c5-b3ae312d2d07")
	require.NoError(t, err)

	assert.Equal(t, "e828446a-8b11-4b01-98c5-b3ae312d2d07", *token.Id)

	cardData, err := external.ExtractCardDataFrom(token)
	require.NoError(t, err)
	assert.Equal(t, "XXXXXXXXXXXX0002", cardData.TokenizedNumber)
	assert.Equal(t, "02", cardData.ExpirationMonth)
	assert.Equal(t, "2024", cardData.ExpirationYear)
}
