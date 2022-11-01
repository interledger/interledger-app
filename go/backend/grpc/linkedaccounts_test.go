package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/linkedaccounts"
	_user "gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetLinkedAccounts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	user := &_user.User{
		ID: uuid.NewString(),
	}

	t.Run("requires authenticated user", func(st *testing.T) {

		response, err := client.GetLinkedAccounts(
			user_mock.ActingAsContext(t, context.Background(), nil),
			&backendv1.Empty{},
		)

		assert.Nil(st, response)
		assert.Error(st, err)
	})

	t.Run("returns linked accounts", func(st *testing.T) {
		wallet, err := c.Users().CreateNewWallet(ctx, user.ID, "default")
		if err != nil {
			t.Fatal(err)
		}
		walletID := wallet.ID
		expectedLinkedAccounts := []linkedaccounts.LinkedAccount{
			{
				ID:       uuid.NewString(),
				WalletID: walletID,
				Name:     "test1",
				Mask:     "abc",
			},
			{
				ID:       uuid.NewString(),
				WalletID: walletID,
				Name:     "test2",
				Mask:     "cba",
			},
		}
		c.linkedaccounts.EXPECT().ListByWalletId(gomock.Any(), walletID).Return(expectedLinkedAccounts, nil).Times(1)

		response, err := client.GetLinkedAccounts(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.Empty{},
		)
		if err != nil {
			st.Fatal(err)
		}

		assert.Len(st, response.GetLinkedAccounts(), 2)
		assert.Equal(st, response.GetLinkedAccounts()[0].Id, expectedLinkedAccounts[0].ID)
		assert.Equal(st, response.GetLinkedAccounts()[0].Name, expectedLinkedAccounts[0].Name)
		assert.Equal(st, response.GetLinkedAccounts()[0].Mask, expectedLinkedAccounts[0].Mask)
		assert.Equal(st, response.GetLinkedAccounts()[1].Id, expectedLinkedAccounts[1].ID)
		assert.Equal(st, response.GetLinkedAccounts()[1].Name, expectedLinkedAccounts[1].Name)
		assert.Equal(st, response.GetLinkedAccounts()[1].Mask, expectedLinkedAccounts[1].Mask)
	})
}
