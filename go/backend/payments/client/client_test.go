package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/client"
	"gopkg.in/stretchr/testify.v1/require"
)

func TestClient(t *testing.T) {
	ctx := context.Background()
	b := &TestBackends{
		db: db.MigrateTestDB(t, ctx),
	}
	pc := client.New(b)
	sendWalletID, sendLinkedAccount := createTestWallet(t, b)
	receiveWalletID, receiveLinkedAccount := createTestWallet(t, b)

	p, err := pc.Create(ctx, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: sendWalletID,
		},
		SenderAccount: sendLinkedAccount,
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: receiveWalletID,
		},
		ReceiverAccount: receiveLinkedAccount,
		SenderAmount:    currency.FromUInt64(10, currency.ParseCurrency("USD")),
		ReceiverAmount:  currency.FromUInt64(10, currency.ParseCurrency("USD")),
	})
	require.NoError(t, err)

	p, err = pc.Update(ctx, payments.UpdateArgs{
		ID:        p.ID,
		ThreeDSID: "123",
	})
	require.NoError(t, err)

	p, requiredActions, err := pc.Confirm(ctx, p.ID)
	require.NoError(t, err)
	assert.Empty(t, requiredActions)
}

/*
Seeds a user:
- user client returns user for userID
- user client returns list of users
- wallet is created for userID
- linked card is created
- tabapay account is created
*/
func createTestWallet(t *testing.T, b *TestBackends) (string, string) {
	return "", ""
}
