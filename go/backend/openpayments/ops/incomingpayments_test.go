package ops_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/openpayments/ops"
	users_client "gitlab.com/fynbos/backend/user/client"
)

func TestCreateIncomingPayment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)

	b := ops.NewTestBackends(t, db, nil, nil)

	userClient := users_client.New(b, "fakeURL", "fakeAdminURL")

	cases := []struct {
		name    string
		ppAsset string
		args    openpayments.CreateIncomingPaymentArgs
		err     error
	}{
		{
			name: "success",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "http://fynbos.me/moneyplease",
				FromPaymentPointer: "http://fynbos.me/sendingmoney",
				IncomingAmount: &openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
				ExternalRef: "external",
				Description: "Desc Incoming Payment",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
		},
		{
			name: "success no incoming amount",
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "http://fynbos.me/moneyplease4",
				FromPaymentPointer: "http://fynbos.me/sendingmoney4",
				ExternalRef:        "external",
				ExpiresAt:          time.Now().Add(time.Hour),
			},
		},
		{
			name:    "different assets",
			ppAsset: "ZAR",
			err:     openpayments.ErrInvalidArgument,
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "http://fynbos.me/moneyplease2",
				FromPaymentPointer: "http://fynbos.me/sendingmoney2",
				IncomingAmount: &openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
				ExternalRef: "external",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
		},
		{
			name: "past expiry",
			err:  openpayments.ErrInvalidArgument,
			args: openpayments.CreateIncomingPaymentArgs{
				PaymentPointer:     "http://fynbos.me/moneyplease3",
				FromPaymentPointer: "http://fynbos.me/sendingmoney3",
				IncomingAmount: &openpayments.Amount{
					Value:      100,
					Asset:      "USD",
					AssetScale: 2,
				},
				ExternalRef: "external",
				ExpiresAt:   time.Now().Add(time.Hour * -1),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recvUserID := uuid.NewString()
			sendUserID := uuid.NewString()
			// Create Signups
			_, err := db.ExecContext(ctx, "INSERT INTO signups (id, user_id) VALUES ($1, $2), ($3, $4)", uuid.NewString(), recvUserID, uuid.NewString(), sendUserID)
			require.NoError(t, err)
			// Create Wallets
			recvWallet, err := userClient.CreateNewWallet(ctx, recvUserID, "test")
			require.NoError(t, err)
			sendWallet, err := userClient.CreateNewWallet(ctx, sendUserID, "test")
			require.NoError(t, err)

			asset := "USD"
			assetScale := 2
			if tc.args.IncomingAmount != nil {
				asset = tc.args.IncomingAmount.Asset
				assetScale = tc.args.IncomingAmount.AssetScale
			}
			if tc.ppAsset != "" {
				asset = tc.ppAsset
			}
			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.args.PaymentPointer,
				WalletID:   recvWallet.ID,
				Alias:      "Alias",
				Asset:      asset,
				AssetScale: assetScale,
			})
			require.NoError(t, err)
			err = ops.CreatePaymentPointer(ctx, b, openpayments.PaymentPointer{
				URL:        tc.args.FromPaymentPointer,
				WalletID:   sendWallet.ID,
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
			if tc.args.IncomingAmount != nil {
				assert.Equal(t, tc.args.IncomingAmount.Asset, ip.IncomingAmount.Asset)
				assert.Equal(t, tc.args.IncomingAmount.AssetScale, ip.IncomingAmount.AssetScale)
				assert.Equal(t, tc.args.IncomingAmount.Value, ip.IncomingAmount.Value)
			}
		})
	}
}
