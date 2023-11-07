package client_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/pacioli"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/identities"

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
	SenderReferralAmount    currency.Amount
	ReceiverReferralAmount  currency.Amount
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

	b.user.MapUserWallet(context.Background(), uuid.NewString(), wallets.WebMonetizationWalletID)
	sendWalletID, sendLinkedAccount, sendBalance, sendBank := createTestWallet(t, b)
	receiveWalletID, receiveLinkedAccount, receiveBalance, _ := createTestWallet(t, b)
	webMonetizaiontLinkedAccount, err := b.LinkedAccounts().GetDefaultSend(ctx, wallets.WebMonetizationWalletID, currency.USD)
	require.NoError(t, err)

	// adding dummy transaction so referrals don't run
	_, err = b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID: receiveWalletID,
		Provider: transactions.Provider(tabapay.ProviderName),
		Amount: currency.Amount{
			Value:    10,
			Currency: currency.USD,
		},
		ForeignType: transactions.TransactionTypeIncoming,
		State:       transactions.StateCompleted,
		Destination: "https://local.fynbos.me/test",
	})
	require.NoError(t, err)

	cases := []struct {
		Name        string
		Args        payments.CreateArgs
		Assertions  Assertions
		AddIdentity bool
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
				SenderAmount:    currency.FromUInt64(777, currency.ParseCurrency("USD")),
				ReceiverAmount:  currency.FromUInt64(777, currency.ParseCurrency("USD")),
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
				SenderAmount:    currency.FromUInt64(666, currency.ParseCurrency("USD")),
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
				PaymentState:         payments.StateFailed,
				SendTransactionState: transactions.StateFailed,
			},
		},
		{
			Name:        "Requires account linking",
			AddIdentity: true,
			Args: payments.CreateArgs{
				Sender: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: sendWalletID,
				},
				SenderAccount: sendLinkedAccount,
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeSlack,
					Identifier: "fynbos / DevTest",
				},
				SenderAmount:   currency.FromUInt64(10, currency.ParseCurrency("USD")),
				ReceiverAmount: currency.FromUInt64(10, currency.ParseCurrency("USD")),
				IPAddress:      "192.36.8.4",
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
			Name: "Golden path Web monetization",
			Args: payments.CreateArgs{
				Sender: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: wallets.WebMonetizationWalletID,
				},
				SenderAccount: webMonetizaiontLinkedAccount.ID,
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: receiveWalletID,
				},
				ReceiverAccount: receiveLinkedAccount,
				SenderAmount:    currency.FromUInt64(10, currency.ParseCurrency("USD")),
				ReceiverAmount:  currency.FromUInt64(10, currency.ParseCurrency("USD")),
				IPAddress:       "192.36.8.4",
				Type:            payments.TypeWebMonetization,
			},
			Assertions: Assertions{
				PaymentState:         payments.StateCompleted,
				SendTransactionState: transactions.StateCompleted,
				SendTransfers: []AssertTransfer{
					{
						TransferType: transactions.TransferTypeDebitWebMonetization,
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
			Name: "Golden path Xago withdrawal",
			Args: payments.CreateArgs{
				Sender: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: sendWalletID,
				},
				SenderAccount: sendBalance,
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: sendWalletID,
				},
				ReceiverAccount: sendBank,
				SenderAmount:    currency.FromUInt64(10, currency.ParseCurrency("ZAR")),
				ReceiverAmount:  currency.FromUInt64(10, currency.ParseCurrency("ZAR")),
				IPAddress:       "192.36.8.4",
				Type:            payments.TypeWithdrawal,
			},
			Assertions: Assertions{
				PaymentState:         payments.StateCompleted,
				SendTransactionState: transactions.StateCompleted,
				SendTransfers: []AssertTransfer{
					{
						TransferType: transactions.TransferTypeDebitBalance,
						State:        transactions.StateCompleted,
					},
				},
			},
		},
		{
			Name: "Golden path xago wallets",
			Args: payments.CreateArgs{
				Sender: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: sendWalletID,
				},
				SenderAccount: sendBalance,
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: receiveWalletID,
				},
				ReceiverAccount: receiveBalance,
				SenderAmount:    currency.FromUInt64(10000, currency.ParseCurrency("ZAR")),
				ReceiverAmount:  currency.FromUInt64(10000, currency.ParseCurrency("ZAR")),
				IPAddress:       "192.36.8.4",
			},
			Assertions: Assertions{
				PaymentState:         payments.StateCompleted,
				SendTransactionState: transactions.StateCompleted,
				SendTransfers: []AssertTransfer{
					{
						TransferType: transactions.TransferTypeDebitBalance,
						State:        transactions.StateCompleted,
					},
				},
				ReceiveTransfers: []AssertTransfer{
					{
						TransferType: transactions.TransferTypeCreditBalance,
						State:        transactions.StateCompleted,
					},
				},
				ReceiveTransactionState: transactions.StateCompleted,
			},
		},
		/*
			Temporal test environment doesn't accurately throw Max Retry errors so cannot test the failure case
			{
				Name: "Xago withdrawal fails",
				Args: payments.CreateArgs{
					Sender: payments.Identity{
						Type:       payments.IdentityTypeWalletID,
						Identifier: sendWalletID,
					},
					SenderAccount: sendBalance,
					Receiver: payments.Identity{
						Type:       payments.IdentityTypeWalletID,
						Identifier: sendWalletID,
					},
					ReceiverAccount: sendBank,
					SenderAmount:    currency.FromUInt64(666, currency.ParseCurrency("ZAR")),
					ReceiverAmount:  currency.FromUInt64(666, currency.ParseCurrency("ZAR")),
					IPAddress:       "192.36.8.4",
					Type:            payments.TypeWithdrawal,
				},
				Assertions: Assertions{
					PaymentState:         payments.StateFailed,
					SendTransactionState: transactions.StateFailed,
					SendTransfers: []AssertTransfer{
						{
							TransferType: transactions.TransferTypeDebitBalance,
							State:        transactions.StateCompleted,
						},
						{
							TransferType: transactions.TransferTypeCreditBalance,
							State:        transactions.StateCompleted,
						},
					},
				},
			},*/
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
				OTP:       "123456",
			})
			require.NoError(st, err)

			if tc.AddIdentity {
				b.env.RegisterDelayedCallback(func() {

					var platform identities.Platform
					switch tc.Args.Receiver.Type {
					case payments.IdentityTypeSlack:
						platform = identities.PlatformSlack
					case payments.IdentityTypeDiscord:
						platform = identities.PlatformDiscord
					case payments.IdentityTypeTwitter:
						platform = identities.PlatformTwitter
					}

					id, err := b.Identities().Add(ctx, identities.AddArgs{
						WalletID:   receiveWalletID,
						Platform:   platform,
						Identifier: tc.Args.Receiver.Identifier,
					})
					require.NoError(st, err)
					err = b.Identities().UpdateState(ctx, id.ID, identities.StateVerified, "")
					require.NoError(st, err)
				}, time.Hour)

			}

			p, requiredActions, err := pc.Confirm(ctx, p.ID)
			require.NoError(st, err)
			require.Empty(st, requiredActions)

			for {
				// Just so we don't spin on IsWorkflowCompleted
				time.Sleep(100 * time.Millisecond)

				if b.env.IsWorkflowCompleted() {
					break
				}
			}
			require.NoError(st, b.env.GetWorkflowError())

			p, err = pc.Lookup(ctx, p.ID)
			require.NoError(st, err)
			assert.Equal(st, tc.Assertions.PaymentState, p.State)

			var sendTransaction *transactions.Transaction
			if p.Type == payments.TypeWebMonetization {
				sendTransaction, err = b.Transactions().GetTransaction(ctx, wallets.WebMonetizationWalletID, p.SendTransactionID)
			} else {
				sendTransaction, err = b.Transactions().GetTransaction(ctx, sendWalletID, p.SendTransactionID)
			}
			require.NoError(st, err)
			assert.True(st, strings.HasPrefix(sendTransaction.Destination, "https://local.fynbos.me/"))
			assert.True(st, strings.HasPrefix(sendTransaction.Source, "https://local.fynbos.me/"))
			assert.Equal(st, tc.Assertions.SendTransactionState, sendTransaction.State)
			sendTransfers := []AssertTransfer{}

			sendXfers, err := b.Transactions().ListTransfers(ctx, sendTransaction.ID)
			require.NoError(st, err)

			for _, xfer := range sendXfers {
				sendTransfers = append(sendTransfers, AssertTransfer{TransferType: xfer.Type, State: xfer.State})
			}
			assert.ElementsMatch(st, tc.Assertions.SendTransfers, sendTransfers)

			if tc.Assertions.ReceiveTransactionState != "" {
				recvTransaction, err := b.Transactions().GetTransaction(ctx, receiveWalletID, p.ReceiveTransactionID)
				require.NoError(st, err)
				assert.True(st, strings.HasPrefix(recvTransaction.Destination, "https://local.fynbos.me/"))
				assert.True(st, strings.HasPrefix(recvTransaction.Source, "https://local.fynbos.me/"))
				assert.Equal(st, tc.Assertions.ReceiveTransactionState, recvTransaction.State)
				recvTransfers := []AssertTransfer{}

				recvXfers, err := b.Transactions().ListTransfers(ctx, recvTransaction.ID)
				require.NoError(st, err)

				for _, xfer := range recvXfers {
					recvTransfers = append(recvTransfers, AssertTransfer{TransferType: xfer.Type, State: xfer.State})
				}
				assert.ElementsMatch(st, tc.Assertions.ReceiveTransfers, recvTransfers)
			} else {
				assert.Empty(st, p.ReceiveTransactionID)
			}
		})
	}
}

func TestReferrals(t *testing.T) {
	env.SetEnv(t, "local")
	ctx := context.Background()
	b := NewTestBackends(t)

	pc := client.New(b)

	b.user.MapUserWallet(context.Background(), uuid.NewString(), wallets.ReferralsWalletID)
	referralsWallet, err := b.Wallets().Get(ctx, wallets.ReferralsWalletID)
	require.NoError(t, err)

	cases := []struct {
		Name                   string
		Args                   payments.CreateArgs
		Assertions             Assertions
		AddIdentity            bool
		AddReceiverTransaction bool
	}{
		// {
		// 	Name: "Creates $20.00 referral for new user with linked identity",
		// 	Assertions: Assertions{
		// 		PaymentState:         payments.StateCompleted,
		// 		SendTransactionState: transactions.StateCompleted,
		// 		SendTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeDebitCard,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 		ReceiveTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeCreditCard,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 		ReceiveTransactionState: transactions.StateCompleted,
		// 		SenderReferralAmount:    currency.FromUInt64(20_00, currency.USD),
		// 		ReceiverReferralAmount:  currency.FromUInt64(20_00, currency.USD),
		// 	},
		// 	AddIdentity: true,
		// },
		// {
		// 	Name: "Creates $10.00 referral for new user without linked identity",
		// 	Assertions: Assertions{
		// 		PaymentState:         payments.StateCompleted,
		// 		SendTransactionState: transactions.StateCompleted,
		// 		SendTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeDebitCard,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 		ReceiveTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeCreditCard,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 		ReceiveTransactionState: transactions.StateCompleted,
		// 		SenderReferralAmount:    currency.FromUInt64(10_00, currency.USD),
		// 		ReceiverReferralAmount:  currency.FromUInt64(10_00, currency.USD),
		// 	},
		// },
		{
			Name: "Only creates referrals for new user",
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
			AddReceiverTransaction: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(st *testing.T) {
			b.RestoreTemporalEnv()
			sendWalletID, sendLinkedAccount, _, _ := createTestWallet(t, b)
			receiveWalletID, receiveLinkedAccount, _, _ := createTestWallet(t, b)
			if tc.AddIdentity {
				id, err := b.Identities().Add(ctx, identities.AddArgs{
					WalletID:   receiveWalletID,
					Platform:   identities.PlatformDiscord,
					Identifier: "devdiscord",
				})
				require.NoError(st, err)
				err = b.Identities().UpdateState(ctx, id.ID, identities.StateVerified, "")
				require.NoError(st, err)
			}

			if tc.AddReceiverTransaction {
				_, err := b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
					WalletID: receiveWalletID,
					Provider: transactions.Provider(tabapay.ProviderName),
					Amount: currency.Amount{
						Value:    10,
						Currency: currency.USD,
					},
					ForeignType: transactions.TransactionTypeIncoming,
					State:       transactions.StateCompleted,
					Destination: "https://local.fynbos.me/test",
				})
				require.NoError(st, err)
			}

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
				IPAddress:       "192.36.8.4",
			})
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
				OTP:       "123456",
			})
			require.NoError(st, err)

			p, requiredActions, err := pc.Confirm(ctx, p.ID)
			require.NoError(st, err)
			require.Empty(st, requiredActions)

			for {
				// Just so we don't spin on IsWorkflowCompleted
				time.Sleep(100 * time.Millisecond)

				if b.env.IsWorkflowCompleted() {
					break
				}
			}
			require.NoError(st, b.env.GetWorkflowError())

			p, err = pc.Lookup(ctx, p.ID)
			require.NoError(st, err)
			assert.Equal(st, tc.Assertions.PaymentState, p.State)

			var sendTransaction *transactions.Transaction
			if p.Type == payments.TypeWebMonetization {
				sendTransaction, err = b.Transactions().GetTransaction(ctx, wallets.WebMonetizationWalletID, p.SendTransactionID)
			} else {
				sendTransaction, err = b.Transactions().GetTransaction(ctx, sendWalletID, p.SendTransactionID)
			}
			require.NoError(st, err)
			assert.True(st, strings.HasPrefix(sendTransaction.Destination, "https://local.fynbos.me/"))
			assert.True(st, strings.HasPrefix(sendTransaction.Source, "https://local.fynbos.me/"))
			assert.Equal(st, tc.Assertions.SendTransactionState, sendTransaction.State)
			sendTransfers := []AssertTransfer{}
			sendXfers, err := b.Transactions().ListTransfers(ctx, sendTransaction.ID)
			require.NoError(st, err)
			for _, xfer := range sendXfers {
				sendTransfers = append(sendTransfers, AssertTransfer{TransferType: xfer.Type, State: xfer.State})
			}
			assert.ElementsMatch(st, tc.Assertions.SendTransfers, sendTransfers)

			if tc.Assertions.ReceiveTransactionState != "" {
				recvTransaction, err := b.Transactions().GetTransaction(ctx, receiveWalletID, p.ReceiveTransactionID)
				require.NoError(st, err)
				assert.True(st, strings.HasPrefix(recvTransaction.Destination, "https://local.fynbos.me/"))
				assert.True(st, strings.HasPrefix(recvTransaction.Source, "https://local.fynbos.me/"))
				assert.Equal(st, tc.Assertions.ReceiveTransactionState, recvTransaction.State)
				recvTransfers := []AssertTransfer{}
				recvXfers, err := b.Transactions().ListTransfers(ctx, recvTransaction.ID)
				require.NoError(st, err)
				for _, xfer := range recvXfers {
					recvTransfers = append(recvTransfers, AssertTransfer{TransferType: xfer.Type, State: xfer.State})
				}
				require.NoError(st, err)
				assert.ElementsMatch(st, tc.Assertions.ReceiveTransfers, recvTransfers)
			} else {
				assert.Empty(st, p.ReceiveTransactionID)
			}

			receiveTxs, err := b.Transactions().ListCompleted(ctx, db.Pagination{}, sendWalletID)
			require.NoError(st, err)
			var senderReferral *transactions.Transaction
			for _, tx := range receiveTxs {
				if tx.Source == referralsWallet.AddressString() {
					senderReferral = &tx
					break
				}
			}
			if tc.Assertions.SenderReferralAmount.Value > 0 {
				require.NotNil(st, senderReferral)
				ts, err := b.Transactions().ListTransfers(ctx, senderReferral.ID)
				require.NoError(st, err)
				require.Len(st, ts, 1)
				assert.Equal(st, tc.Assertions.SenderReferralAmount, ts[0].Amount)
			} else {
				assert.Nil(st, senderReferral)
			}

			recevierTxs, err := b.Transactions().ListCompleted(ctx, db.Pagination{}, receiveWalletID)
			require.NoError(st, err)
			var recevierReferral *transactions.Transaction
			for _, tx := range recevierTxs {
				if tx.Source == referralsWallet.AddressString() {
					recevierReferral = &tx
					break
				}
			}
			if tc.Assertions.ReceiverReferralAmount.Value > 0 {
				require.NotNil(st, recevierReferral)
				ts, err := b.Transactions().ListTransfers(ctx, recevierReferral.ID)
				require.NoError(st, err)
				require.Len(st, ts, 1)
				assert.Equal(st, tc.Assertions.ReceiverReferralAmount, ts[0].Amount)
			} else {
				assert.Nil(st, recevierReferral)
			}
		})
	}
}

func TestPayAnyone(t *testing.T) {
	env.SetEnv(t, "local")
	ctx := context.Background()
	b := NewTestBackends(t)
	b.RestoreTemporalEnv()

	pc := client.New(b)

	sendWalletID, sendLinkedAccount := createTestWallet(t, b)

	payment, err := pc.Create(ctx, payments.CreateArgs{
		Sender: payments.Identity{
			Type:       payments.IdentityTypeWalletID,
			Identifier: sendWalletID,
		},
		SenderAccount: sendLinkedAccount,
		Receiver: payments.Identity{
			Type:       payments.IdentityTypeUnknown,
			Identifier: "Justin",
		},
		SenderAmount: currency.FromUInt64(100, currency.ParseCurrency("USD")),
		IPAddress:    "192.36.8.4",
	})
	require.NoError(t, err)

	// generate 3DS session
	threeDSSession, err := b.tabapay.Init3DS(ctx, tabapay.Init3DSArgs{
		Amount:  currency.FromUInt64(100, currency.ParseCurrency("USD")),
		OrderID: payment.ID,
		CardID:  sendLinkedAccount,
	})
	require.NoError(t, err)

	payment, err = pc.Update(ctx, payments.UpdateArgs{
		ID:        payment.ID,
		ThreeDSID: threeDSSession.ID,
		OTP:       "123456",
	})
	require.NoError(t, err)

	link, err := b.Payments().CreatePaymentLink(ctx, payment.ID)
	require.NoError(t, err)
	var receiverLinkAccountID string
	b.env.RegisterDelayedCallback(func() {
		link, err = b.Payments().ConsumePaymentLink(ctx, payments.ConsumePaymentLinkArgs{
			ID:        link.ID,
			FirstName: "Justin",
			LastName:  "Time",
			Email:     "justin@test.com",
			IpAddress: "10.0.10.10",
		})
		require.NoError(t, err)

		la, err := b.LinkedAccounts().Create(context.Background(), &linkedaccounts.CreateArgs{
			WalletID:        link.ReceiverWalletID,
			Name:            "default",
			Provider:        tabapay.ProviderName,
			ProviderID:      uuid.NewString(),
			CanSend:         true,
			CanReceive:      true,
			Type:            tabapay.TypeCard,
			SendCurrency:    currency.USD,
			ReceiveCurrency: currency.USD,
		})
		require.NoError(t, err)
		receiverLinkAccountID = la.ID

		link, err = b.Payments().CompletePaymentLink(ctx, link.ID, la.ID)
		require.NoError(t, err)
	}, time.Hour)

	payment, requireActions, err := pc.Confirm(ctx, payment.ID)
	require.NoError(t, err)
	require.Empty(t, requireActions)

	for {
		// Just so we don't spin on IsWorkflowCompleted
		time.Sleep(100 * time.Millisecond)

		if b.env.IsWorkflowCompleted() {
			break
		}
	}
	require.NoError(t, b.env.GetWorkflowError())

	require.False(t, link.CompletedAt.IsZero())

	payment, err = b.Payments().Lookup(ctx, payment.ID)
	require.NoError(t, err)
	assert.Equal(t, payments.StateCompleted, payment.State)
	assert.NotEmpty(t, link.ReceiverWalletID)
	assert.Equal(t, link.ReceiverWalletID, payment.Receiver.WalletID)

	assert.NotEmpty(t, link.ReceiverLinkedAccountID)
	assert.Equal(t, receiverLinkAccountID, payment.ReceiverAccount)
}

/*
Seeds a user:
- user client returns user for userID
- user client returns list of users
- wallet is created for userID
- linked card is created
- tabapay account is created
- xago account and balance account is created
*/
func createTestWallet(t *testing.T, b *TestBackends) (string, string, string, string) {
	userID := uuid.NewString()
	address, err := wallets.ParseAddress(fmt.Sprintf("%s/%s", env.OpenPaymentsURL(), faker.FirstName()))
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
		WalletID:        wallet.ID,
		Name:            "default",
		Provider:        tabapay.ProviderName,
		ProviderID:      uuid.NewString(),
		CanSend:         true,
		CanReceive:      true,
		Type:            tabapay.TypeCard,
		SendCurrency:    currency.USD,
		ReceiveCurrency: currency.USD,
	})
	require.NoError(t, err)

	xBalance, err := b.LinkedAccounts().Create(context.Background(), &linkedaccounts.CreateArgs{
		WalletID:        wallet.ID,
		Name:            "xago_balance",
		Provider:        xago.ProviderName,
		ProviderID:      uuid.NewString(),
		Type:            xago.AccTypeBalance,
		CanSend:         true,
		CanReceive:      true,
		State:           linkedaccounts.Verified,
		SendCountry:     country.ZA,
		SendCurrency:    currency.ZAR,
		ReceiveCountry:  country.ZA,
		ReceiveCurrency: currency.ZAR,
	})
	require.NoError(t, err)

	accs, err := b.pac.ConfigureAccounts(context.Background(), []pacioli.ConfigureAccountArgs{
		{
			ID:       xBalance.ID,
			LedgerID: xago.LedgerIDZAR,
			Code:     1,
		},
	})
	require.NoError(t, err)
	require.Len(t, accs, 0)

	tr, err := b.pac.CreateTransfers(context.Background(), []pacioli.CreateTransferArgs{
		{
			ID:              uuid.NewString(),
			Amount:          100000,
			DebitAccountID:  xago.ZAROpsAccount,
			CreditAccountID: xBalance.ID,
			Pending:         false,
			Code:            1,
			Timeout:         0,
			Ledger:          xago.LedgerIDZAR,
		},
	})
	require.NoError(t, err)
	require.Len(t, tr, 0)

	xBank, err := b.LinkedAccounts().Create(context.Background(), &linkedaccounts.CreateArgs{
		WalletID:        wallet.ID,
		Name:            "xago_bank",
		Provider:        xago.ProviderName,
		ProviderID:      uuid.NewString(),
		Type:            xago.AccTypeBank,
		CanReceive:      true,
		State:           linkedaccounts.Verified,
		SendCountry:     country.ZA,
		SendCurrency:    currency.ZAR,
		ReceiveCountry:  country.ZA,
		ReceiveCurrency: currency.ZAR,
	})
	require.NoError(t, err)

	return wallet.ID, la.ID, xBalance.ID, xBank.ID
}
