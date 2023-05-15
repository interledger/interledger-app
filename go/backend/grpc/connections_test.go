package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/keys"
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
	w, err := c.Users().CreateNewWallet(context.Background(), user.CreateWalletArgs{
		UserID: u.ID,
		Name:   "Marko Polo",
	})
	require.NoError(t, err)

	pemEncodedPublicKey := `-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAJrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs=
-----END PUBLIC KEY-----`
	base64PublicKey := "JrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs="

	keyID := uuid.NewString()
	c.keys.EXPECT().AddPublicKey(gomock.Any(), w.ID, base64PublicKey, "FynTest").Return(
		&keys.Key{
			ID:        keyID,
			Name:      "FynTest",
			WalletID:  w.ID,
			Reference: "",
			PublicKey: "base64PublicKey",
		},
		nil,
	).AnyTimes()

	c.limits.EXPECT().UpdatePublicKeyLimits(gomock.Any(), w.ID, keyID, limits.Limit{
		Daily: currency.Amount{
			Value:    10,
			Currency: currency.Currency("USD"),
			Scale:    2,
		},
		Monthly: currency.Amount{
			Value:    100,
			Currency: currency.Currency("USD"),
			Scale:    2,
		},
		Overall: currency.Amount{
			Value:    1000,
			Currency: currency.Currency("USD"),
			Scale:    2,
		},
	}).AnyTimes()

	_, err = client.CreateConnection(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.CreateConnectionRequest{
		ApplicationName: "FynTest",
		PublicKey:       pemEncodedPublicKey,
		DailyLimit: &backendv1.Amount{
			Asset:      "USD",
			AssetScale: 2,
			Amount:     10,
		},
		MonthlyLimit: &backendv1.Amount{
			Asset:      "USD",
			AssetScale: 2,
			Amount:     100,
		},
		OverallLimit: &backendv1.Amount{
			Asset:      "USD",
			AssetScale: 2,
			Amount:     1000,
		},
	})
	require.NoError(t, err)
}

func TestGetAndListPublicKeys(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	w, err := c.Users().CreateNewWallet(context.Background(), user.CreateWalletArgs{
		UserID: u.ID,
		Name:   "Marko Polo",
	})
	require.NoError(t, err)

	base64PublicKey := "JrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs="
	keyFingerprint := "SHA256:22ce02aa18eb1ee5f39482d0f57a6ba56f4d549f81db547f3bea2863207c8a01"
	keyUuid := uuid.NewString()
	c.keys.EXPECT().List(gomock.Any(), w.ID).Return(
		[]keys.Key{
			{
				ID:        keyUuid,
				Reference: "",
				PublicKey: base64PublicKey,
				Name:      "FynTest",
				Type:      keys.NonCustodial,
			},
		},
		nil,
	).AnyTimes()

	listRpc, err := client.ListConnections(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, err)
	require.Len(t, listRpc.GetKeys(), 1)
	assert.Equal(t, keyFingerprint, listRpc.GetKeys()[0].PublicKeyFingerprint)
	assert.Equal(t, "FynTest", listRpc.GetKeys()[0].ApplicationName)

	getRpc, err := client.GetConnection(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.GetConnectionRequest{
		Id: keyUuid,
	})
	require.NoError(t, err)
	assert.Equal(t, keyFingerprint, getRpc.GetPublicKeyFingerprint())
	assert.Equal(t, "FynTest", getRpc.GetApplicationName())
}

func TestUpdatePublicKeyLimits(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	w, err := c.Users().CreateNewWallet(context.Background(), user.CreateWalletArgs{
		UserID: u.ID,
		Name:   "Marko Polo",
	})
	require.NoError(t, err)

	ppURL := "https://local.fynbos.me/test"
	c.OPClient.EXPECT().GetWalletPaymentPointer(gomock.Any(), w.ID).Return(&openpayments.PaymentPointer{
		ID:       uuid.NewString(),
		WalletID: w.ID,
		URL:      ppURL,
	}, nil).AnyTimes()

	publicKeyUuid := uuid.NewString()
	c.limits.EXPECT().UpdatePublicKeyLimits(gomock.Any(), w.ID, publicKeyUuid, limits.Limit{
		Daily: currency.Amount{
			Value:    10,
			Currency: "USD",
			Scale:    2,
		},
		Monthly: currency.Amount{
			Value:    100,
			Currency: "USD",
			Scale:    2,
		},
		Overall: currency.Amount{
			Value:    1000,
			Currency: "USD",
			Scale:    2,
		},
	}).Return(nil).AnyTimes()

	_, err = client.UpdateConnectionLimits(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.UpdateConnectionLimitsRequest{
		Id: publicKeyUuid,
		Daily: &backendv1.Amount{
			Asset:      "USD",
			AssetScale: 2,
			Amount:     10,
		},
		Monthly: &backendv1.Amount{
			Asset:      "USD",
			AssetScale: 2,
			Amount:     100,
		},
		Overall: &backendv1.Amount{
			Asset:      "USD",
			AssetScale: 2,
			Amount:     1000,
		},
	})
	require.NoError(t, err)
}

func TestDeletePublicKey(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	w, err := c.Users().CreateNewWallet(context.Background(), user.CreateWalletArgs{
		UserID: u.ID,
		Name:   "Marko Polo",
	})
	require.NoError(t, err)

	base64PublicKey := "JrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs="
	keyID := uuid.NewString()
	c.keys.EXPECT().List(gomock.Any(), w.ID).Return(
		[]keys.Key{
			{
				ID:        keyID,
				Reference: "",
				PublicKey: base64PublicKey,
				Name:      "FynTest",
				Type:      keys.NonCustodial,
			},
		},
		nil,
	).AnyTimes()

	c.keys.EXPECT().DeletePublicKey(gomock.Any(), keyID).Return(nil).AnyTimes()

	_, err = client.DeleteConnection(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.DeleteConnectionRequest{
		Id: keyID,
	})
	require.NoError(t, err)
}
