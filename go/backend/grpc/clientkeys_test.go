package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestCreatePublicKey(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	w, err := c.Users().CreateNewWallet(context.Background(), u.ID, "Marko Polo")
	require.NoError(t, err)

	ppURL := "https://local.fynbos.me/test"
	c.OPClient.EXPECT().GetWalletPaymentPointer(gomock.Any(), w.ID).Return(&openpayments.PaymentPointer{
		ID:       uuid.NewString(),
		WalletID: w.ID,
		URL:      ppURL,
	}, nil).AnyTimes()

	c.authorisation.EXPECT().AddPublicKey(gomock.Any(), ppURL, authorisation.Jwk{
		Kty: "OKP",
		Kid: "FynTest",
		Alg: "EdDSA",
		Crv: "Ed25519",
		X:   "le key",
		Use: "sign",
	}).AnyTimes()

	c.limits.EXPECT().UpdateClientLimits(gomock.Any(), w.ID, ppURL, limits.Limit{
		Daily: currency.Amount{
			Value:    10,
			Currency: currency.Currency("USD"),
		},
		Monthly: currency.Amount{
			Value:    100,
			Currency: currency.Currency("USD"),
		},
		Overall: currency.Amount{
			Value:    1000,
			Currency: currency.Currency("USD"),
		},
	}).AnyTimes()

	_, err = client.CreatePublicKey(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.CreatePublicKeyRequest{
		ApplicationName: "FynTest",
		PublicKey:       "le key",
		DailyLimit: &backendv1.PublicKeyLimit{
			Currency: "USD",
			Amount:   10,
		},
		MonthlyLimit: &backendv1.PublicKeyLimit{
			Currency: "USD",
			Amount:   100,
		},
		OverallLimit: &backendv1.PublicKeyLimit{
			Currency: "USD",
			Amount:   1000,
		},
	})
	require.NoError(t, err)
}
