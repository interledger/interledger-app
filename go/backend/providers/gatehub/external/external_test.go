package external_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/env"
)

func TestClient(t *testing.T) {
	env.SetEnv(t, "local")
	if os.Getenv("GATEHUB_APP_ID") == "" || os.Getenv("GATEHUB_SECRET") == "" || os.Getenv("GATEHUB_APP_ID") == "" {
		t.SkipNow()
	}
	c := external.NewClient(os.Getenv("GATEHUB_APP_ID"), os.Getenv("GATEHUB_SECRET"), os.Getenv("GATEHUB_GATEWAY_ID"), nil)

	ctx := context.Background()
	sendingExternalUserID := "66f1427e-43e4-48a0-9692-190c24d75058"
	// trx, err := c.CreateTransaction(ctx, external.CreateTransactionRequest{
	// 	SendingUserID:    sendingExternalUserID,
	// 	SendingAddress:   "107720301",
	// 	ReceivingAddress: "506541309", // belongs to "19227839-caa1-458f-a5ec-a3f03aa3e0e5"
	// 	Amount:           5.00,
	// 	VaultID:          "a09a0a2c-1a3a-44c5-a1b9-603a6eea9341",
	// 	Message:          "test transfer",
	// 	Type:             external.TransactionTypeHosted,
	// })
	// require.NoError(t, err)

	trx, err := c.GetTransaction(ctx, sendingExternalUserID, "b063d802-f0d7-463c-bc14-4310b6e313f4")
	require.NoError(t, err)

	temp, err := json.MarshalIndent(trx, "", " ")
	require.NoError(t, err)
	fmt.Printf("transaction: %s\n", temp)
}

func TestUser(t *testing.T) {
	env.SetEnv(t, "local")
	if os.Getenv("GATEHUB_APP_ID") == "" || os.Getenv("GATEHUB_SECRET") == "" {
		t.SkipNow()
	}
	c := external.NewClient(os.Getenv("GATEHUB_APP_ID"), os.Getenv("GATEHUB_SECRET"), os.Getenv("GATEHUB_GATEWAY_ID"), nil)
	ctx := context.Background()

	userID := "66f1427e-43e4-48a0-9692-190c24d75058"
	u, err := c.GetUser(ctx, userID)
	require.NoError(t, err)

	temp, err := json.MarshalIndent(u, "", " ")
	require.NoError(t, err)
	fmt.Printf("fetched user: %s\n", temp)
}
