package smileid_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/kyc/smileid"
	"gitlab.com/fynbos/env"
	"gopkg.in/stretchr/testify.v1/require"
)

func TestClient(t *testing.T) {
	env.SetEnv(t, "local")
	partnerID := os.Getenv("SMILE_ID_PARTNER_ID")
	apiKey := os.Getenv("SMILE_ID_API_KEY")
	if partnerID == "" || apiKey == "" {
		t.Skip("Credentials not set")
	}

	client := smileid.New(partnerID, apiKey)

	walletID := uuid.NewString()
	jobID := uuid.NewString()
	token, err := client.GetToken(context.Background(), walletID, jobID, smileid.EnhancedKYCProduct)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}
