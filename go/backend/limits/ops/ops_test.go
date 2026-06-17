package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/db"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/limits"
	"github.com/interledger/interledger-app/go/backend/limits/ops"
	"github.com/interledger/interledger-app/go/backend/transactions"
	tx_client "github.com/interledger/interledger-app/go/backend/transactions/client"
	users_mock "github.com/interledger/interledger-app/go/backend/user/client/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExceedsKYCLimits(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbc := db.MigrateTestDB(t, ctx)
	fee := currency.FromFloat64(2, currency.USD)

	b := NewTestBackends(t, dbc, users_mock.NewMock())
	txClient := tx_client.New(b)
	cases := []struct {
		name               string
		tx                 *transactions.CreateTransactionArgs
		amnt               currency.Amount
		expectExceedsLimit bool
		expectLimitType    limits.LimitType
		kycLevel           kyc.Status
	}{
		{
			name:               "exceeds single tx limit",
			amnt:               currency.FromFloat64(11_000, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeTransaction,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name:               "under single tx limit no previous transactions",
			amnt:               currency.FromFloat64(249, currency.USD),
			expectExceedsLimit: false,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name: "does not exceed daily defaults",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://ilp.link/alice2",
				Destination: "https://ilp.link/bob2",
				Amount:      currency.FromFloat64(2_900, currency.USD),
				ProviderFee: &fee,
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:               currency.FromFloat64(98, currency.USD),
			expectExceedsLimit: false,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name: "does exceed daily limits",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://ilp.link/alice2",
				Destination: "https://ilp.link/bob2",
				Amount:      currency.FromFloat64(2_900, currency.USD),
				ProviderFee: &fee,
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:               currency.FromFloat64(99, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeDaily,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name:               "pending kyc exceeds limits",
			amnt:               currency.FromFloat64(1, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeTransaction,
			kycLevel:           kyc.StatusPending,
		},
		{
			name:               "denied kyc exceeds limits",
			amnt:               currency.FromFloat64(1, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeTransaction,
			kycLevel:           kyc.StatusDenied,
		},
		{
			name:               "unknown kyc exceeds limits",
			amnt:               currency.FromFloat64(1, currency.USD),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeTransaction,
			kycLevel:           kyc.StatusUnknown,
		},
		{
			name: "exceeds ZAR yearly limits",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://ilp.link/alice2",
				Destination: "https://ilp.link/bob2",
				Amount:      currency.FromFloat64(20_000, currency.ZAR),
				ProviderFee: &fee,
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:               currency.FromFloat64(99, currency.ZAR),
			expectExceedsLimit: true,
			expectLimitType:    limits.LimitTypeYearly,
			kycLevel:           kyc.StatusLevel1,
		},
		{
			name: "No ZAR limits",
			tx: &transactions.CreateTransactionArgs{
				State:       transactions.StateCompleted,
				Source:      "https://ilp.link/alice2",
				Destination: "https://ilp.link/bob2",
				Amount:      currency.FromFloat64(200_000, currency.ZAR),
				ProviderFee: &fee,
				Provider:    "test",
				ForeignType: transactions.TransactionTypeOpenOutgoingPayment,
			},
			amnt:               currency.FromFloat64(99, currency.ZAR),
			expectExceedsLimit: false,
			kycLevel:           kyc.StatusLevel2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			walletID := uuid.NewString()
			b.kyc.EXPECT().GetKYCStatus(gomock.Any(), walletID).Return(tc.kycLevel, nil).AnyTimes()

			if tc.tx != nil {
				tc.tx.WalletID = walletID

				_, err := txClient.CreateTransaction(ctx, *tc.tx)
				require.NoError(t, err)
			}

			exceeds, limitType, err := ops.ExceedsKYCLimits(ctx, b, walletID, tc.amnt)
			require.NoError(t, err)
			assert.Equal(t, tc.expectExceedsLimit, exceeds)
			assert.Equal(t, tc.expectLimitType, limitType)
		})
	}
}
