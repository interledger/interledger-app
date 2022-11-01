package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/fakecash"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestFakeCashAccount(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	user := &user.User{
		ID: uuid.NewString(),
	}
	wallet, err := c.Users().CreateNewWallet(context.Background(), user.ID, "default")
	require.NoError(t, err)

	t.Run("requires authenticated user", func(st *testing.T) {
		response, err := client.LinkCashAccount(
			user_mock.ActingAsContext(t, context.Background(), nil),
			&backendv1.LinkCashAccountRequest{
				Name: "test",
			},
		)

		assert.Nil(st, response)
		assert.Error(st, err)
	})

	t.Run("creates fake cash account backed linked account", func(st *testing.T) {
		faskcashAccountID := uuid.NewString()
		linkedAccountID := uuid.NewString()
		c.fakecash.EXPECT().Create(gomock.Any(), fakecash.CreateArgs{}).Return(&fakecash.Account{
			ID: faskcashAccountID,
		}, nil)
		c.linkedaccounts.EXPECT().Create(gomock.Any(), &linkedaccounts.CreateArgs{
			WalletID:   wallet.ID,
			Name:       "test",
			Provider:   "fakecash",
			ProviderID: faskcashAccountID,
			Type:       "fakecash",
		}).Return(
			&linkedaccounts.LinkedAccount{
				ID:         linkedAccountID,
				WalletID:   wallet.ID,
				Name:       "test",
				Provider:   "fakecash",
				ProviderID: faskcashAccountID,
				Type:       "fakecash",
			},
			nil,
		)

		response, err := client.LinkCashAccount(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.LinkCashAccountRequest{
				Name: "test",
			},
		)

		require.NoError(st, err)
		assert.Equal(st, linkedAccountID, response.GetId())
		assert.Equal(st, "test", response.GetName())
		assert.Equal(st, "fakecash", response.GetType())
	})
}
