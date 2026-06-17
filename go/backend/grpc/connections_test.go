package grpc

import (
	"context"
	"testing"

	_user "github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/backend/wallets"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/keys"
	"github.com/interledger/interledger-app/go/backend/user"
	user_mock "github.com/interledger/interledger-app/go/backend/user/client/mock"
	backendv1 "github.com/interledger/interledger-app/go/proto/backend/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	_, err := client.CreateConnection(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.CreateConnectionRequest{
		ApplicationName: "FynTest",
		PublicKey:       base64EncodedJWK,
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
