package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/client"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
)

type Assertions struct {
	PaymentState            payments.State
	SendTransactionState    transactions.State
	SendTransfers           []AssertTransfer
	ReceiveTransactionState transactions.State
	ReceiveTransfers        []AssertTransfer
}

type AssertTransfer struct {
	TransferType transactions.TransferType
	State        transactions.State
}

func TestClient(t *testing.T) {
	env.SetEnv(t, "local")
	ctx := context.Background()
	b := NewTestBackends(t)

	pc := client.New(b)
	sendWalletID, sendLinkedAccount := createTestWallet(t, b)
	receiveWalletID, receiveLinkedAccount := createTestWallet(t, b)

	cases := []struct {
		Name       string
		Args       payments.CreateArgs
		Assertions Assertions
	}{
		{
			Name: "Golden path",
			Args: payments.CreateArgs{
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
				IPAddress:       "192.36.8.4",
			},
			Assertions: Assertions{
				PaymentState:         payments.StateCompleted,
				SendTransactionState: transactions.StateCompleted,
				SendTransfers: []AssertTransfer{
					{
						TransferType: transactions.TransferTypeDebitCard,
						State:        transactions.StateCompleted,
					},
				},
				ReceiveTransfers: []AssertTransfer{
					{
						TransferType: transactions.TransferTypeCreditCard,
						State:        transactions.StateCompleted,
					},
				},
				ReceiveTransactionState: transactions.StateCompleted,
			},
		},
		{
			Name: "Pull from card fails",
			Args: payments.CreateArgs{
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
				SenderAmount:    currency.FromUInt64(666, currency.ParseCurrency("USD")),
				ReceiverAmount:  currency.FromUInt64(10, currency.ParseCurrency("USD")),
				IPAddress:       "192.36.8.4",
			},
			Assertions: Assertions{
				PaymentState:         payments.StateFailed,
				SendTransactionState: transactions.StateFailed,
				SendTransfers:        []AssertTransfer{},
				ReceiveTransfers:     []AssertTransfer{},
			},
		},
		{
			Name: "Push to card fails",
			Args: payments.CreateArgs{
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
				ReceiverAmount:  currency.FromUInt64(666, currency.ParseCurrency("USD")),
				IPAddress:       "192.36.8.4",
			},
			Assertions: Assertions{
				PaymentState:         payments.StateFailed,
				SendTransactionState: transactions.StateFailed,
				SendTransfers: []AssertTransfer{
					{
						TransferType: transactions.TransferTypeDebitCard,
						State:        transactions.StateCompleted,
					},
					{
						TransferType: transactions.TransferTypeCreditCard,
						State:        transactions.StateCompleted,
					},
				},
				ReceiveTransfers: []AssertTransfer{},
			},
		},
		{
			Name: "Compliance fails",
			Args: payments.CreateArgs{
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
				SenderAmount:    currency.FromUInt64(1222, currency.ParseCurrency("USD")),
				ReceiverAmount:  currency.FromUInt64(1222, currency.ParseCurrency("USD")),
				IPAddress:       "192.36.8.4",
			},
			Assertions: Assertions{
				PaymentState: payments.StateFailed,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(st *testing.T) {
			b.RestoreTemporalEnv()
			p, err := pc.Create(ctx, tc.Args)
			require.NoError(st, err)
			// generate 3DS session
			threeDSSession, err := b.tabapay.Init3DS(ctx, tabapay.Init3DSArgs{
				Amount:  tc.Args.SenderAmount,
				OrderID: p.ID,
				CardID:  sendLinkedAccount,
			})
			require.NoError(st, err)

			p, err = pc.Update(ctx, payments.UpdateArgs{
				ID:        p.ID,
				ThreeDSID: threeDSSession.ID,
			})
			require.NoError(st, err)

			p, requiredActions, err := pc.Confirm(ctx, p.ID)
			require.NoError(st, err)
			require.Empty(st, requiredActions)

			for {
				if b.env.IsWorkflowCompleted() {
					break
				}
				// Just so we don't spin on IsWorkflowCompleted
				time.Sleep(100 * time.Millisecond)
			}
			require.NoError(st, b.env.GetWorkflowError())

			p, err = pc.Lookup(ctx, p.ID)
			require.NoError(st, err)
			assert.Equal(st, tc.Assertions.PaymentState, p.State)

			if tc.Assertions.SendTransactionState != "" {
				sendTransaction, err := b.Transactions().GetTransaction(ctx, sendWalletID, p.SendTransactionID)
				require.NoError(st, err)
				assert.Equal(st, tc.Assertions.SendTransactionState, sendTransaction.State)
				sendTransfers := []AssertTransfer{}
				for _, tr := range sendTransaction.Transfers {
					sendTransfers = append(sendTransfers, AssertTransfer{TransferType: tr.Type, State: tr.State})
				}
				assert.ElementsMatch(st, tc.Assertions.SendTransfers, sendTransfers)
			} else {
				assert.Empty(st, p.SendTransactionID)
			}

			if tc.Assertions.ReceiveTransactionState != "" {
				recvTransaction, err := b.Transactions().GetTransaction(ctx, receiveWalletID, p.ReceiveTransactionID)
				require.NoError(st, err)
				assert.Equal(st, tc.Assertions.ReceiveTransactionState, recvTransaction.State)
				recvTransfers := []AssertTransfer{}
				for _, tr := range recvTransaction.Transfers {
					recvTransfers = append(recvTransfers, AssertTransfer{TransferType: tr.Type, State: tr.State})
				}
				assert.ElementsMatch(st, tc.Assertions.ReceiveTransfers, recvTransfers)
			} else {
				assert.Empty(st, p.ReceiveTransactionID)
			}
		})
	}
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
