package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/client"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gotest.tools/assert"
)

func TestClient(t *testing.T) {
	env.SetEnv(t, "local")
	ctx := context.Background()
	b := NewTestBackends(t)

	sendWalletID, sendLinkedAccount := createTestWallet(t, b)
	receiveWalletID, receiveLinkedAccount := createTestWallet(t, b)

	pc := client.New(b)
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

	// generate 3DS session
	threeDSSession, err := b.tabapay.Init3DS(ctx, tabapay.Init3DSArgs{
		Amount:  p.SenderAmount,
		OrderID: p.ID,
		CardID:  sendLinkedAccount,
	})
	require.NoError(t, err)

	p, err = pc.Update(ctx, payments.UpdateArgs{
		ID:        p.ID,
		ThreeDSID: threeDSSession.ID,
	})
	require.NoError(t, err)

	p, requiredActions, err := pc.Confirm(ctx, p.ID)
	require.NoError(t, err)
	require.Empty(t, requiredActions)

	for {
		if b.env.IsWorkflowCompleted() {
			break
		}
	}
	require.NoError(t, b.env.GetWorkflowError())

	payment, err := pc.Lookup(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, payments.StateCompleted, payment.State)
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
	userID := uuid.NewString()
	address, err := wallets.ParseAddress(fmt.Sprintf("https://fynbos.test/%s", faker.FirstName()))
	if err != nil {
		t.Fatal(err)
	}

	walletID := uuid.NewString()
	b.user.MapUserWallet(context.Background(), userID, walletID)
	wallet, err := b.Wallets().Create(context.Background(), wallets.CreateArgs{
		UserID:    userID,
		ID:        walletID,
		Addresses: []wallets.Address{address},
	})
	require.NoError(t, err)
	require.Equal(t, walletID, wallet.ID)

	la, err := b.LinkedAccounts().Create(context.Background(), &linkedaccounts.CreateArgs{
		WalletID:   wallet.ID,
		Name:       "default",
		Provider:   tabapay.ProviderName,
		ProviderID: uuid.NewString(),
		CanSend:    true,
		CanReceive: true,
		Type:       tabapay.TypeCard,
	})
	require.NoError(t, err)

	return wallet.ID, la.ID
}
