package ops_test

import (
	"context"
	"testing"
	"time"

	wallets_mock "gitlab.com/fynbos/backend/wallets/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	transactions_mock "gitlab.com/fynbos/backend/transactions/client/mock"
)

func TestCreateIncomingPayment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	tc := transactions_mock.NewMockClient(ctrl)
	txID := uuid.NewString()
	tc.EXPECT().CreateTransactionTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(txID, nil).AnyTimes()
	wc := wallets_mock.NewMockClient(ctrl)
	wc.EXPECT().AddAddress(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	b := ops.NewTestBackends(t, db, nil, tc, func(b *ops.TestBackends) {
		b.Wc = wc
	})

	cases := []struct {
		name    string
		ppAsset string
		args    openpayments.CreateIncomingPaymentArgs
		err     error
	}{
		{
			name: "success",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "https://fynbos.me/moneyplease",
				FromPaymentPointer: "https://fynbos.me/sendingmoney",
				IncomingAmount: &currency.Amount{
					Value:    100,
					Currency: currency.ParseCurrency("USD"),
					Scale:    2,
				},
				ExternalRef: "external",
				Description: "Desc Incoming Payment",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
		},
		{
			name: "success with created by",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "https://fynbos.me/moneypleaseapi",
				FromPaymentPointer: "https://fynbos.me/sendingmoneyapi",
				IncomingAmount: &currency.Amount{
					Value:    100,
					Currency: currency.ParseCurrency("USD"),
					Scale:    2,
				},
				ExternalRef: "external",
				Description: "Desc Incoming Payment",
				ExpiresAt:   time.Now().Add(time.Hour),
				CreatedBy:   uuid.NewString(),
			},
		},
		{
			name: "success no incoming amount",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "https://fynbos.me/moneyplease4",
				FromPaymentPointer: "https://fynbos.me/sendingmoney4",
				ExternalRef:        "external",
				ExpiresAt:          time.Now().Add(time.Hour),
			},
		},
		{
			name:    "different assets",
			ppAsset: "ZAR",
			err:     openpayments.ErrInvalidArgument,
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "https://fynbos.me/moneyplease2",
				FromPaymentPointer: "https://fynbos.me/sendingmoney2",
				IncomingAmount: &currency.Amount{
					Value:    100,
					Currency: currency.ParseCurrency("USD"),
					Scale:    2,
				},
				ExternalRef: "external",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
		},
		{
			name: "past expiry",
			err:  openpayments.ErrInvalidArgument,
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "https://fynbos.me/moneyplease3",
				FromPaymentPointer: "https://fynbos.me/sendingmoney3",
				IncomingAmount: &currency.Amount{
					Value:    100,
					Currency: currency.ParseCurrency("USD"),
					Scale:    2,
				},
				ExternalRef: "external",
				ExpiresAt:   time.Now().Add(time.Hour * -1),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recvWalletID := uuid.NewString()
			sendWalletID := uuid.NewString()

			asset := currency.USD
			assetScale := 2
			if tc.args.IncomingAmount != nil {
				asset = tc.args.IncomingAmount.Currency
				assetScale = tc.args.IncomingAmount.Scale
			}
			if tc.ppAsset != "" {
				asset = currency.ParseCurrency(tc.ppAsset)
			}
			err := ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.args.PaymentPointer,
				WalletID:   recvWalletID,
				Alias:      "Alias",
				Asset:      asset,
				AssetScale: assetScale,
			})
			require.NoError(t, err)
			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.args.FromPaymentPointer,
				WalletID:   sendWalletID,
				Alias:      "Alias",
				Asset:      asset,
				AssetScale: assetScale,
			})
			require.NoError(t, err)

			ip, err := ops.CreateIncomingPayment(ctx, b, tc.args)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.args.PaymentPointer, ip.PaymentPointer)
			assert.Equal(t, tc.args.ExternalRef, ip.ExternalRef)
			assert.Equal(t, tc.args.Description, ip.Description)
			assert.Equal(t, tc.args.CreatedBy, ip.CreatedBy)
			if tc.args.IncomingAmount != nil {
				assert.Equal(t, tc.args.IncomingAmount.Currency, ip.IncomingAmount.Currency)
				assert.Equal(t, tc.args.IncomingAmount.Scale, ip.IncomingAmount.Scale)
				assert.Equal(t, tc.args.IncomingAmount.Value, ip.IncomingAmount.Value)
			}
		})
	}
}
