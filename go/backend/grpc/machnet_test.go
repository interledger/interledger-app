package grpc

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	_user "gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetMachnetWidgetToken(t *testing.T) {
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
	wallet, err := c.Users().CreateNewWallet(ctx, user.ID, "default")
	require.NoError(t, err)

	t.Run("requires authenticated user", func(st *testing.T) {
		rpc, err := client.GetMachnetWidgetToken(
			user_mock.ActingAsContext(t, context.Background(), nil),
			&backendv1.Empty{},
		)

		assert.Nil(st, rpc)
		assert.Error(st, err)
	})

	t.Run("returns token", func(st *testing.T) {
		machnetUserID := uuid.NewString()
		c.machnet.EXPECT().GetWidgetToken(gomock.Any(), wallet.ID).Return(&machnet.WidgetToken{
			Value:            "machnet-widget-token",
			ExpiresInMinutes: 15,
			UserID:           machnetUserID,
		}, nil).Times(1)

		rpc, err := client.GetMachnetWidgetToken(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.Empty{},
		)
		require.NoError(st, err)
		assert.Equal(st, "machnet-widget-token", rpc.GetValue())
		assert.Equal(st, int64(15), rpc.GetExpiresInMinutes())
		assert.Equal(st, machnetUserID, rpc.GetUserId())
	})
}

func TestHasSendUser(t *testing.T) {
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
	wallet, err := c.Users().CreateNewWallet(ctx, user.ID, "default")
	require.NoError(t, err)

	t.Run("returns true if user exists", func(st *testing.T) {
		machnetUserID := uuid.NewString()
		c.machnet.EXPECT().GetUserByWalletID(gomock.Any(), wallet.ID).Return(&machnet.User{
			ID:        machnetUserID,
			WalletID:  wallet.ID,
			CreatedAt: "",
			UpdatedAt: "",
			KYCStatus: machnet.KYCStatusInProgress,
		}, nil).Times(1)

		rpc, err := client.HasSendUser(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.Empty{},
		)
		require.NoError(st, err)

		assert.True(t, rpc.HasSendUser)
	})

	t.Run("returns false if no user exists", func(st *testing.T) {
		c.machnet.EXPECT().GetUserByWalletID(gomock.Any(), wallet.ID).Return(nil, machnet.ErrNotFound).Times(1)

		rpc, err := client.HasSendUser(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.Empty{},
		)
		require.NoError(st, err)

		assert.False(t, rpc.HasSendUser)
	})
}

func TestCreateSendUser(t *testing.T) {
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
	wallet, err := c.Users().CreateNewWallet(ctx, user.ID, "default")
	require.NoError(t, err)

	t.Run("returns if workflow runs correctly", func(st *testing.T) {
		await := func(ctx context.Context, out interface{}) error {
			return nil
		}
		c.machnet.EXPECT().CreateSendUser(gomock.Any(), wallet.ID).Return(await, nil).Times(1)

		_, err := client.CreateSendUser(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.Empty{},
		)
		require.NoError(st, err)
	})

	t.Run("returns error if workflow fails", func(st *testing.T) {
		await := func(ctx context.Context, out interface{}) error {
			return machnet.ErrInternal
		}
		c.machnet.EXPECT().CreateSendUser(gomock.Any(), wallet.ID).Return(await, nil).Times(1)

		_, err := client.CreateSendUser(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.Empty{},
		)
		require.Error(st, err)
	})

}

func TestCreateWallet(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	user := &_user.User{
		ID: uuid.NewString(),
	}
	wallet, err := c.Users().CreateNewWallet(context.Background(), user.ID, "default")
	require.NoError(t, err)

	t.Run("requires authenticated user", func(st *testing.T) {
		rpc, err := client.GetMachnetWidgetToken(
			user_mock.ActingAsContext(t, context.Background(), nil),
			&backendv1.Empty{},
		)
		require.NotNil(st, err)
		assert.Nil(st, rpc)
	})

	t.Run("creates linked account", func(st *testing.T) {
		name := faker.Name()
		sendUserID := uuid.NewString()
		c.machnet.EXPECT().GetUserByWalletID(gomock.Any(), wallet.ID).Return(
			&machnet.User{
				ID:       sendUserID,
				WalletID: wallet.ID,
			},
			nil,
		).Times(1)
		linkedAccountID := uuid.NewString()
		externalWalletID := uuid.NewString()
		c.machnet.EXPECT().CreateWallet(gomock.Any(), machnet.CreateWalletArgs{
			Nickname:   name,
			SendUserID: sendUserID,
		}).Return(
			&linkedaccounts.LinkedAccount{
				ID:         linkedAccountID,
				WalletID:   wallet.ID,
				Name:       name,
				Mask:       "Fynbos Cash",
				Provider:   machnet.ProviderName,
				ProviderID: externalWalletID,
				Type:       machnet.TypeWallet,
			},
			nil,
		).Times(1)

		rpc, err := client.CreateWallet(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.CreateWalletRequest{
				Nickname: name,
			})
		require.NoError(st, err)

		assert.Equal(st, linkedAccountID, rpc.GetId())
		assert.Equal(st, "Fynbos Cash", rpc.GetMask())
		assert.Equal(st, name, rpc.GetName())
		assert.Equal(st, machnet.TypeWallet, rpc.GetType())
	})

	t.Run("returns already exists if user has a machnet wallet with a different name", func(st *testing.T) {
		name := faker.Name()
		sendUserID := uuid.NewString()
		c.machnet.EXPECT().GetUserByWalletID(gomock.Any(), wallet.ID).Return(
			&machnet.User{
				ID:       sendUserID,
				WalletID: wallet.ID,
			},
			nil,
		).Times(1)
		c.machnet.EXPECT().CreateWallet(gomock.Any(), machnet.CreateWalletArgs{
			Nickname:   name,
			SendUserID: sendUserID,
		}).Return(nil, machnet.ErrUserHasExistingWallet).Times(1)

		_, err := client.CreateWallet(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.CreateWalletRequest{
				Nickname: name,
			})
		require.Error(st, err)
		grpcStatus, ok := status.FromError(err)
		require.True(st, ok)
		assert.Equal(st, codes.AlreadyExists, grpcStatus.Code())
	})
}

func TestRpcService_GetWalletBalance(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	user := &_user.User{
		ID: uuid.NewString(),
	}
	wallet, err := c.Users().CreateNewWallet(context.Background(), user.ID, "default")
	require.NoError(t, err)

	t.Run("requires authenticated user", func(st *testing.T) {
		rpc, err := client.GetWalletBalance(
			user_mock.ActingAsContext(t, context.Background(), nil),
			&backendv1.Empty{},
		)
		require.NotNil(st, err)
		assert.Nil(st, rpc)
	})

	t.Run("get wallet balance", func(st *testing.T) {
		walletProviderID := uuid.NewString()
		c.linkedaccounts.EXPECT().ListByWalletId(gomock.Any(), wallet.ID).Return(
			[]linkedaccounts.LinkedAccount{
				{
					Provider:   machnet.ProviderName,
					ProviderID: walletProviderID,
					Type:       machnet.TypeWallet,
				},
			}, nil)
		c.machnet.EXPECT().GetWallet(gomock.Any(), walletProviderID).Return(
			&machnet.Wallet{
				AvailableBalance: 100,
				Balance:          110,
			},
			nil,
		).Times(1)

		rpc, err := client.GetWalletBalance(
			user_mock.ActingAsContext(t, context.Background(), user),
			&backendv1.Empty{})
		require.NoError(st, err)

		assert.Equal(st, uint64(100), rpc.Available)
		assert.Equal(st, uint64(110), rpc.Balance)
	})
}

func TestWithdrawFromMachnetWallet(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	user := &_user.User{
		ID: uuid.NewString(),
	}
	wallet, err := c.Users().CreateNewWallet(context.Background(), user.ID, "default")
	require.NoError(t, err)

	t.Run("requires authenticated user", func(st *testing.T) {
		rpc, err := client.GetMachnetWidgetToken(
			user_mock.ActingAsContext(t, context.Background(), nil),
			&backendv1.Empty{},
		)
		require.NotNil(st, err)
		assert.Nil(st, rpc)
	})

	t.Run("creates withdrawal", func(st *testing.T) {
		toLinkedAccountID := uuid.NewString()
		c.linkedaccounts.EXPECT().Get(gomock.Any(), toLinkedAccountID).Return(&linkedaccounts.LinkedAccount{
			ID:       toLinkedAccountID,
			WalletID: wallet.ID,
		}, nil).Times(1)

		walletLinkedAccountID := uuid.NewString()
		c.linkedaccounts.EXPECT().ListByWalletId(gomock.Any(), wallet.ID).Return(
			[]linkedaccounts.LinkedAccount{
				{ID: walletLinkedAccountID, WalletID: wallet.ID, Provider: machnet.ProviderName, Type: machnet.TypeWallet},
			},
			nil,
		).Times(1)

		withdrawalID := uuid.NewString()
		c.machnet.EXPECT().WithdrawFromWallet(gomock.Any(), machnet.WithdrawFromWalletArgs{
			WalletLinkedAccountID: walletLinkedAccountID,
			Amount:                1000,
			ToLinkedAccountID:     toLinkedAccountID,
			IpAddress:             "10.10.10.10",
		}).Return(&machnet.WalletWithdrawal{
			ID:                withdrawalID,
			Amount:            1000,
			ToLinkedAccountID: toLinkedAccountID,
			Status:            "PROCESSING",
		}, nil).Times(1)

		withdrawal, err := client.WithdrawFromMachnetWallet(
			user_mock.ActingAsContext(st, context.Background(), user),
			&backendv1.WithdrawFromMachnetWalletRequest{
				ToLinkedAccountId: toLinkedAccountID,
				Amount:            1000,
				IpAddress:         "10.10.10.10",
			},
		)
		require.NoError(st, err)
		assert.Equal(st, "PROCESSING", withdrawal.Status)
		assert.Equal(st, toLinkedAccountID, withdrawal.ToLinkedAccountId)
		assert.Equal(st, uint64(1000), withdrawal.Amount)
		assert.Equal(st, withdrawalID, withdrawal.Id)
	})

	t.Run("validates request", func(st *testing.T) {
		_, err := client.WithdrawFromMachnetWallet(
			user_mock.ActingAsContext(st, context.Background(), user),
			&backendv1.WithdrawFromMachnetWalletRequest{
				ToLinkedAccountId: "asd",
				Amount:            0,
			},
		)
		require.Error(st, err)

		grpcStatus, ok := status.FromError(err)
		require.True(st, ok)
		errorFields := []string{}
		for _, detail := range grpcStatus.Details() {
			for _, violation := range detail.(*errdetails.BadRequest).FieldViolations {
				errorFields = append(errorFields, violation.Field)
			}
		}
		assert.EqualValues(t, errorFields, []string{"ToLinkedAccount", "Amount", "IpAddress"})
	})
}
