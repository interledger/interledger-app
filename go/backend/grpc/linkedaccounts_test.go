package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	_user "github.com/interledger/interledger-app/go/backend/user"
	user_mock "github.com/interledger/interledger-app/go/backend/user/client/mock"
	"github.com/interledger/interledger-app/go/backend/wallets"
	backendv1 "github.com/interledger/interledger-app/go/proto/backend/v1"
	"github.com/stretchr/testify/assert"
)

func TestGetLinkedAccounts(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &_user.User{
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
		wallet := wallets.Wallet{
			ID:   uuid.NewString(),
			Name: "testing",
		}
		c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil).AnyTimes()
		c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil).AnyTimes()
		walletID := wallet.ID
		expectedLinkedAccounts := []linkedaccounts.LinkedAccount{
			{
				ID:       uuid.NewString(),
				WalletID: walletID,
				Name:     "test1",
				Mask:     "abc",
				Nickname: "nicky",
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
			user_mock.ActingAsContext(t, context.Background(), u),
			&backendv1.Empty{},
		)
		if err != nil {
			st.Fatal(err)
		}

		assert.Len(st, response.GetLinkedAccounts(), 2)
		assert.Equal(st, response.GetLinkedAccounts()[0].Id, expectedLinkedAccounts[0].ID)
		assert.Equal(st, response.GetLinkedAccounts()[0].Name, expectedLinkedAccounts[0].Name)
		assert.Equal(st, response.GetLinkedAccounts()[0].Mask, expectedLinkedAccounts[0].Mask)
		assert.Equal(st, response.GetLinkedAccounts()[0].Nickname, expectedLinkedAccounts[0].Nickname)
		assert.Equal(st, response.GetLinkedAccounts()[0].Title, expectedLinkedAccounts[0].Nickname)
		assert.Equal(st, response.GetLinkedAccounts()[1].Id, expectedLinkedAccounts[1].ID)
		assert.Equal(st, response.GetLinkedAccounts()[1].Name, expectedLinkedAccounts[1].Name)
		assert.Equal(st, response.GetLinkedAccounts()[1].Mask, expectedLinkedAccounts[1].Mask)
		assert.Equal(st, response.GetLinkedAccounts()[1].Nickname, expectedLinkedAccounts[1].Nickname)
		assert.Equal(st, response.GetLinkedAccounts()[1].Title, expectedLinkedAccounts[1].Mask)
	})
}

func TestGetLinkedAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &_user.User{
		ID: uuid.NewString(),
	}

	t.Run("returns linked account", func(st *testing.T) {
		accID := uuid.NewString()
		wallet := wallets.Wallet{
			ID:   uuid.NewString(),
			Name: "test",
		}
		expectedLinkedAccount := &linkedaccounts.LinkedAccount{
			ID:       accID,
			WalletID: wallet.ID,
			Name:     "test1",
			Mask:     "abc",
		}
		c.linkedaccounts.EXPECT().Get(gomock.Any(), accID).Return(expectedLinkedAccount, nil).Times(1)
		c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{wallet}, nil)
		c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&wallet, nil)

		response, err := client.GetLinkedAccount(
			user_mock.ActingAsContext(t, context.Background(), u),
			&backendv1.GetLinkedAccountRequest{Id: accID},
		)
		require.NoError(st, err)

		assert.Equal(st, response.Id, expectedLinkedAccount.ID)
		assert.Equal(st, response.Name, expectedLinkedAccount.Name)
		assert.Equal(st, response.Mask, expectedLinkedAccount.Mask)
	})
}
