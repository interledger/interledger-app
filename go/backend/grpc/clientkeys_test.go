package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

	pemEncodedPublicKey := `-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAJrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs=
-----END PUBLIC KEY-----`
	base64PublicKey := "JrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs="

	keyUuid := uuid.NewString()
	c.authorisation.EXPECT().AddPublicKey(gomock.Any(), ppURL, authorisation.Jwk{
		Kty: "OKP",
		Kid: "FynTest",
		Alg: "EdDSA",
		Crv: "Ed25519",
		X:   base64PublicKey,
		Use: "sign",
	}).Return(&authorisation.Jwk{
		ID:  keyUuid,
		Kty: "OKP",
		Kid: "FynTest",
		Alg: "EdDSA",
		Crv: "Ed25519",
		X:   base64PublicKey,
		Use: "sign",
	}, nil).AnyTimes()

	c.limits.EXPECT().UpdatePublicKeyLimits(gomock.Any(), w.ID, keyUuid, limits.Limit{
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
	w, err := c.Users().CreateNewWallet(context.Background(), u.ID, "Marko Polo")
	require.NoError(t, err)

	base64PublicKey := "JrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs="
	keyFingerprint := "SHA256:22ce02aa18eb1ee5f39482d0f57a6ba56f4d549f81db547f3bea2863207c8a01"
	ppURL := "https://local.fynbos.me/test"
	c.OPClient.EXPECT().GetWalletPaymentPointer(gomock.Any(), w.ID).Return(&openpayments.PaymentPointer{
		ID:       uuid.NewString(),
		WalletID: w.ID,
		URL:      ppURL,
	}, nil).AnyTimes()

	keyUuid := uuid.NewString()
	c.authorisation.EXPECT().ListKeys(gomock.Any(), ppURL).Return(
		[]authorisation.Jwk{
			{
				ID:  keyUuid,
				Kty: "OKP",
				Alg: "EdDSA",
				Crv: "Ed25519",
				X:   base64PublicKey,
				Use: "sign",
				Kid: "FynTest",
			},
		},
		nil,
	).AnyTimes()
	clientID := uuid.NewString()
	c.authorisation.EXPECT().LookupClient(gomock.Any(), ppURL).Return(
		&authorisation.Client{
			ID: clientID,
		},
		nil,
	).AnyTimes()
	c.authorisation.EXPECT().GetPublicKeyByID(gomock.Any(), keyUuid).Return(
		&authorisation.Jwk{
			ClientID: clientID,
			ID:       keyUuid,
			Kty:      "OKP",
			Alg:      "EdDSA",
			Crv:      "Ed25519",
			X:        base64PublicKey,
			Use:      "sign",
			Kid:      "FynTest",
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
	w, err := c.Users().CreateNewWallet(context.Background(), u.ID, "Marko Polo")
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
	w, err := c.Users().CreateNewWallet(context.Background(), u.ID, "Marko Polo")
	require.NoError(t, err)

	ppURL := "https://local.fynbos.me/test"
	c.OPClient.EXPECT().GetWalletPaymentPointer(gomock.Any(), w.ID).Return(&openpayments.PaymentPointer{
		ID:       uuid.NewString(),
		WalletID: w.ID,
		URL:      ppURL,
	}, nil).AnyTimes()

	clientID := uuid.NewString()
	c.authorisation.EXPECT().LookupClient(gomock.Any(), ppURL).Return(
		&authorisation.Client{
			ID: clientID,
		},
		nil,
	).AnyTimes()
	publicKeyUuid := uuid.NewString()
	c.authorisation.EXPECT().GetPublicKeyByID(gomock.Any(), publicKeyUuid).Return(
		&authorisation.Jwk{
			ClientID: clientID,
			ID:       publicKeyUuid,
			Kty:      "OKP",
			Alg:      "EdDSA",
			Crv:      "Ed25519",
			X:        "le key",
			Use:      "sign",
			Kid:      "FynTest",
		},
		nil,
	).AnyTimes()

	c.authorisation.EXPECT().DeletePublicKey(gomock.Any(), publicKeyUuid).Return(nil).AnyTimes()

	_, err = client.DeleteConnection(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.DeleteConnectionRequest{
		Id: publicKeyUuid,
	})
	require.NoError(t, err)
}
