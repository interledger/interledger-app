package client_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/providers/pti"

	"gitlab.com/fynbos/pacioli"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/providers/xago"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/payments/client"
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
	env.SetEnv(t, "test")
	ctx := context.Background()
	b := NewTestBackends(t)

	pc := client.New(b)

	b.user.MapUserWallet(context.Background(), uuid.NewString(), wallets.WebMonetizationWalletID)
	sendWallet := createTestWallet(t, b)
	recvWallet := createTestWallet(t, b)

	cases := []struct {
		Name        string
		Args        payments.CreateArgs
		Assertions  Assertions
		AddIdentity bool
		Skip        bool
	}{
		// {
		// 	Skip:        true, // PTI needs both parties signed up so we cannot reserve funds pre emptively.
		// 	Name:        "Requires account linking",
		// 	AddIdentity: true,
		// 	Args: payments.CreateArgs{
		// 		Sender: payments.Identity{
		// 			Type:       payments.IdentityTypeWalletID,
		// 			Identifier: sendWallet.walletID,
		// 		},
		// 		SenderAccount: sendWallet.ptiUSDLinkedAcc,
		// 		Receiver: payments.Identity{
		// 			Type:       payments.IdentityTypeSlack,
		// 			Identifier: "interledger / DevTest",
		// 		},
		// 		SenderAmount:   currency.FromUInt64(10, currency.ParseCurrency("USD")),
		// 		ReceiverAmount: currency.FromUInt64(10, currency.ParseCurrency("USD")),
		// 		IPAddress:      "192.36.8.4",
		// 	},
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
		// 	},
		// },
		// {
		// 	Name: "Golden path Web monetization",
		// 	Args: payments.CreateArgs{
		// 		Sender: payments.Identity{
		// 			Type:       payments.IdentityTypeWalletID,
		// 			Identifier: sendWallet.walletID,
		// 		},
		// 		SenderAccount: sendWallet.ptiUSDLinkedAcc,
		// 		Receiver: payments.Identity{
		// 			Type:       payments.IdentityTypeWalletID,
		// 			Identifier: recvWallet.walletID,
		// 		},
		// 		ReceiverAccount: recvWallet.ptiUSDLinkedAcc,
		// 		SenderAmount:    currency.FromUInt64(10, currency.ParseCurrency("USD")),
		// 		ReceiverAmount:  currency.FromUInt64(10, currency.ParseCurrency("USD")),
		// 		IPAddress:       "192.36.8.4",
		// 		Type:            payments.TypeWebMonetization,
		// 	},
		// 	Assertions: Assertions{
		// 		PaymentState:         payments.StateCompleted,
		// 		SendTransactionState: transactions.StateCompleted,
		// 		SendTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeDebitWebMonetization,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 		ReceiveTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeCreditBalance,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 		ReceiveTransactionState: transactions.StateCompleted,
		// 	},
		// },
		{
			Name: "Golden path Xago withdrawal",
			Args: payments.CreateArgs{
				Sender: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: sendWallet.walletID,
				},
				SenderAccount: sendWallet.xagoZARLinkedAcc,
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: sendWallet.walletID,
				},
				ReceiverAccount: sendWallet.xagoBankLinkedAcc,
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
					{
						TransferType: transactions.TransferTypeCreditBankAccount,
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
					Identifier: sendWallet.walletID,
				},
				SenderAccount: sendWallet.xagoZARLinkedAcc,
				Receiver: payments.Identity{
					Type:       payments.IdentityTypeWalletID,
					Identifier: recvWallet.walletID,
				},
				ReceiverAccount: recvWallet.xagoZARLinkedAcc,
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
		// {
		// 	Name: "Golden path pti deposit",
		// 	Args: payments.CreateArgs{
		// 		Type: payments.TypeDeposit,
		// 		Sender: payments.Identity{
		// 			Type:       payments.IdentityTypeWalletID,
		// 			Identifier: sendWallet.walletID,
		// 		},
		// 		SenderAccount: sendWallet.ptiCardLinkedAcc,
		// 		Receiver: payments.Identity{
		// 			Type:       payments.IdentityTypeWalletID,
		// 			Identifier: sendWallet.walletID,
		// 		},
		// 		ReceiverAccount: sendWallet.ptiUSDLinkedAcc,
		// 		SenderAmount:    currency.FromUInt64(10000, currency.USD),
		// 		ReceiverAmount:  currency.FromUInt64(10000, currency.USD),
		// 		IPAddress:       "192.36.8.4",
		// 	},
		// 	Assertions: Assertions{
		// 		PaymentState:         payments.StateCompleted,
		// 		SendTransactionState: transactions.StateCompleted,
		// 		SendTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeDebitCard,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 			{
		// 				TransferType: transactions.TransferTypeCreditBalance,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name: "Golden path pti withdrawal",
		// 	Args: payments.CreateArgs{
		// 		Type: payments.TypeWithdrawal,
		// 		Sender: payments.Identity{
		// 			Type:       payments.IdentityTypeWalletID,
		// 			Identifier: sendWallet.walletID,
		// 		},
		// 		SenderAccount: sendWallet.ptiUSDLinkedAcc,
		// 		Receiver: payments.Identity{
		// 			Type:       payments.IdentityTypeWalletID,
		// 			Identifier: sendWallet.walletID,
		// 		},
		// 		ReceiverAccount: sendWallet.ptiBankLinkedAcc,
		// 		SenderAmount:    currency.FromUInt64(10000, currency.USD),
		// 		ReceiverAmount:  currency.FromUInt64(10000, currency.USD),
		// 		IPAddress:       "192.36.8.4",
		// 	},
		// 	Assertions: Assertions{
		// 		PaymentState:         payments.StateCompleted,
		// 		SendTransactionState: transactions.StateCompleted,
		// 		SendTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeDebitBalance,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 			{
		// 				TransferType: transactions.TransferTypeCreditBankAccount,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name: "Golden path pti wallets",
		// 	Args: payments.CreateArgs{
		// 		Sender: payments.Identity{
		// 			Type:       payments.IdentityTypeWalletID,
		// 			Identifier: sendWallet.walletID,
		// 		},
		// 		SenderAccount: sendWallet.ptiUSDLinkedAcc,
		// 		Receiver: payments.Identity{
		// 			Type:       payments.IdentityTypeWalletID,
		// 			Identifier: recvWallet.walletID,
		// 		},
		// 		ReceiverAccount: recvWallet.ptiUSDLinkedAcc,
		// 		SenderAmount:    currency.FromUInt64(10000, currency.USD),
		// 		ReceiverAmount:  currency.FromUInt64(10000, currency.USD),
		// 		IPAddress:       "192.36.8.4",
		// 	},
		// 	Assertions: Assertions{
		// 		PaymentState:         payments.StateCompleted,
		// 		SendTransactionState: transactions.StateCompleted,
		// 		SendTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeDebitBalance,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 		ReceiveTransfers: []AssertTransfer{
		// 			{
		// 				TransferType: transactions.TransferTypeCreditBalance,
		// 				State:        transactions.StateCompleted,
		// 			},
		// 		},
		// 		ReceiveTransactionState: transactions.StateCompleted,
		// 	},
		// },
		/*
			Temporal test environment doesn't accurately throw Max Retry errors so cannot test the failure case
			{
				Name: "Xago withdrawal fails",
				Args: payments.CreateArgs{
					Sender: payments.Identity{
						Type:       payments.IdentityTypeWalletID,
						Identifier: sendWallet.walletID,
					},
					SenderAccount: sendWallet.xagoZARLinkedAcc,
					Receiver: payments.Identity{
						Type:       payments.IdentityTypeWalletID,
						Identifier: sendWallet.walletID,
					},
					ReceiverAccount: sendWallet.xagoBankLinkedAcc,
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
			if tc.Skip {
				st.Skip("Skipping sub test", tc.Name)
			}
			b.RestoreTemporalEnv()
			p, err := pc.Create(ctx, tc.Args)
			require.NoError(st, err)

			p, err = pc.Update(ctx, payments.UpdateArgs{
				ID:  p.ID,
				OTP: "123456",
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

			sendTransaction, err := b.Transactions().GetTransaction(ctx, sendWallet.walletID, p.SendTransactionID)
			require.NoError(st, err)
			assert.True(st, strings.HasPrefix(sendTransaction.Destination, "https://local.ilp.link/"))
			assert.True(st, strings.HasPrefix(sendTransaction.Source, "https://local.ilp.link/"))
			assert.Equal(st, tc.Assertions.SendTransactionState, sendTransaction.State)
			sendTransfers := []AssertTransfer{}

			sendXfers, err := b.Transactions().ListTransfers(ctx, sendTransaction.ID)
			require.NoError(st, err)

			for _, xfer := range sendXfers {
				sendTransfers = append(sendTransfers, AssertTransfer{TransferType: xfer.Type, State: xfer.State})
			}
			assert.ElementsMatch(st, tc.Assertions.SendTransfers, sendTransfers)

			if tc.Assertions.ReceiveTransactionState != "" {
				recvTransaction, err := b.Transactions().GetTransaction(ctx, recvWallet.walletID, p.ReceiveTransactionID)
				require.NoError(st, err)
				assert.True(st, strings.HasPrefix(recvTransaction.Destination, "https://local.ilp.link/"))
				assert.True(st, strings.HasPrefix(recvTransaction.Source, "https://local.ilp.link/"))
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

type testWallet struct {
	userID            string
	walletID          string
	xagoZARLinkedAcc  string
	xagoBankLinkedAcc string
	// ptiCardLinkedAcc  string
	// ptiBankLinkedAcc  string
	// ptiUSDLinkedAcc   string
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
func createTestWallet(t *testing.T, b *TestBackends) testWallet {
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

	// PTI user
	// _, err = b.DB().Exec(" INSERT INTO pti_users(external_id, wallet_id, status, assessment_status) VALUES ($1, $2, 'confirmed', 'confirmed')", uuid.NewString(), walletID)
	// require.NoError(t, err)

	// ptiCard, err := b.LinkedAccounts().Create(context.Background(), &linkedaccounts.CreateArgs{
	// 	WalletID:        wallet.ID,
	// 	Name:            "pti card",
	// 	Provider:        pti.ProviderName,
	// 	ProviderID:      uuid.NewString(),
	// 	Type:            pti.TypeCard,
	// 	CanSend:         false,
	// 	CanReceive:      false,
	// 	State:           linkedaccounts.Verified,
	// 	SendCountry:     country.US,
	// 	SendCurrency:    currency.USD,
	// 	ReceiveCountry:  country.US,
	// 	ReceiveCurrency: currency.USD,
	// })
	// require.NoError(t, err)

	// ptiBankAccount, err := b.LinkedAccounts().Create(context.Background(), &linkedaccounts.CreateArgs{
	// 	WalletID:        wallet.ID,
	// 	Name:            "pti bank",
	// 	Provider:        pti.ProviderName,
	// 	ProviderID:      uuid.NewString(),
	// 	Type:            pti.TypeBank,
	// 	CanSend:         true,
	// 	CanReceive:      true,
	// 	State:           linkedaccounts.Verified,
	// 	SendCountry:     country.US,
	// 	SendCurrency:    currency.USD,
	// 	ReceiveCountry:  country.US,
	// 	ReceiveCurrency: currency.USD,
	// })
	// require.NoError(t, err)

	ptiBal, err := b.LinkedAccounts().Create(context.Background(), &linkedaccounts.CreateArgs{
		WalletID:        wallet.ID,
		Name:            "pti balance",
		Provider:        pti.ProviderName,
		ProviderID:      uuid.NewString(),
		Type:            pti.AccTypeBalance,
		CanSend:         true,
		CanReceive:      true,
		State:           linkedaccounts.Verified,
		SendCountry:     country.US,
		SendCurrency:    currency.USD,
		ReceiveCountry:  country.US,
		ReceiveCurrency: currency.USD,
	})
	require.NoError(t, err)

	accs, err = b.pac.ConfigureAccounts(context.Background(), []pacioli.ConfigureAccountArgs{
		{
			ID:       ptiBal.ID,
			LedgerID: pti.LedgerIDUSD,
			Code:     1,
		},
	})
	require.NoError(t, err)
	require.Len(t, accs, 0)

	tr, err = b.pac.CreateTransfers(context.Background(), []pacioli.CreateTransferArgs{
		{
			ID:              uuid.NewString(),
			Amount:          100000,
			DebitAccountID:  pti.USDOpsAccount,
			CreditAccountID: ptiBal.ID,
			Pending:         false,
			Code:            1,
			Timeout:         0,
			Ledger:          pti.LedgerIDUSD,
		},
	})
	require.NoError(t, err)
	require.Len(t, tr, 0)

	return testWallet{
		userID:            userID,
		walletID:          walletID,
		xagoZARLinkedAcc:  xBalance.ID,
		xagoBankLinkedAcc: xBank.ID,
		// ptiCardLinkedAcc:  ptiCard.ID,
		// ptiBankLinkedAcc: ptiBankAccount.ID,
		// ptiUSDLinkedAcc: ptiBal.ID,
	}
}
