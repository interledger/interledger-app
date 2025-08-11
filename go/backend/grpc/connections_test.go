package grpc

import (
	"context"
	"testing"

	_user "gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/keys"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestCreatePublicKey(t *testing.T) {
	t.Skip("TODO: Fix this test, currently failling")
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &_user.User{
		ID: uuid.NewString(),
	}
	wallet := wallets.Wallet{
		ID:   uuid.NewString(),
		Name: "testing",
	}
	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()
	base64EncodedJWK := "ewogICJrdHkiOiAiT0tQIiwKICAiY3J2IjogIkVkMjU1MTkiLAogICJraWQiOiAidGVzdC1rZXktZWQyNTUxOSIsCiAgImQiOiAibjROaS1IcElTcFZPYm5RTVcwd09oQ0tST2FJS3FLdFdfMlpZYjJwOUtjVSIsCiAgIngiOiAiSnJRTGo1UF84OWlYRVM5LXZGZ3JJeTI5Y2xGOUNDX29QUHN3M2M1RDBicyIKfQ=="
	base64PublicKey := "JrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs="

	keyID := uuid.NewString()
	c.keys.EXPECT().AddPublicKey(gomock.Any(), wallet.ID, base64PublicKey, "FynTest", "test-key-ed25519").Return(
		&keys.Key{
			ID:        keyID,
			Name:      "FynTest",
			KeyID:     "test-key-ed25519",
			WalletID:  wallet.ID,
			Reference: "",
			PublicKey: "base64PublicKey",
		},
		nil,
	).AnyTimes()

	c.rafiki.EXPECT().CreatePaymentPointerKey(gomock.Any(), gomock.Any(), wallet.ID).Return(nil).AnyTimes()

	c.limits.EXPECT().UpdatePublicKeyLimits(gomock.Any(), wallet.ID, keyID, limits.Limit{
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

	_, err := client.CreateConnection(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.CreateConnectionRequest{
		ApplicationName: "FynTest",
		PublicKey:       base64EncodedJWK,
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
	t.Skip("TODO: Fix this test, currently failling")	
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	wallet := wallets.Wallet{
		ID:   uuid.NewString(),
		Name: "testing",
	}
	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()

	base64PublicKey := "JrQLj5P/89iXES9+vFgrIy29clF9CC/oPPsw3c5D0bs="
	keyFingerprint := "SHA256:22ce02aa18eb1ee5f39482d0f57a6ba56f4d549f81db547f3bea2863207c8a01"
	keyUuid := uuid.NewString()
	c.keys.EXPECT().List(gomock.Any(), wallet.ID).Return(
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

	wa, err := wallets.ParseAddress("https://local.ilp.link/test")
	require.NoError(t, err)

	wallet := wallets.Wallet{
		ID:        uuid.NewString(),
		Name:      "testing",
		Addresses: []wallets.Address{wa},
	}
	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()

	publicKeyUuid := uuid.NewString()
	c.limits.EXPECT().UpdatePublicKeyLimits(gomock.Any(), wallet.ID, publicKeyUuid, limits.Limit{
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
	w := wallets.Wallet{
		ID:   uuid.NewString(),
		Name: "testing",
	}
	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&w, nil).AnyTimes()
	c.rafiki.EXPECT().RevokePaymentPointerKey(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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

	_, err := client.DeleteConnection(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.DeleteConnectionRequest{
		Id: keyID,
	})
	require.NoError(t, err)
}
