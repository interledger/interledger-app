package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/machnet"
	_user "gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
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
