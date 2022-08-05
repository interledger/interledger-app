package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/fundingsources"
	_user "gitlab.com/fynbos/backend/user"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetFundingsources(t *testing.T) {
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
	accountID := uuid.NewString()
	c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), gomock.Any()).Return(
		&accounts.Account{
			ID:         accountID,
			IdentityID: user.ID,
		},
		nil,
	).AnyTimes()

	t.Run("requires authenticated user", func(st *testing.T) {
		response, err := client.GetFundingsources(
			_user.ActingAsContext(t, context.Background(), nil),
			&backendv1.Empty{},
		)

		assert.Nil(st, response)
		assert.Error(st, err)
	})

	t.Run("returns fundingsources", func(st *testing.T) {
		expectedFundingsources := []fundingsources.FundingSource{
			{
				ID:        uuid.NewString(),
				AccountID: accountID,
				Name:      "test1",
				Mask:      "abc",
			},
			{
				ID:        uuid.NewString(),
				AccountID: accountID,
				Name:      "test2",
				Mask:      "cba",
			},
		}
		c.FundingsourceService.EXPECT().GetByAccountId(gomock.Any(), accountID).Return(expectedFundingsources, nil).Times(1)

		response, err := client.GetFundingsources(
			_user.ActingAsContext(t, context.Background(), user),
			&backendv1.Empty{},
		)
		if err != nil {
			st.Fatal(err)
		}

		assert.Len(st, response.GetFundingsources(), 2)
		assert.Equal(st, response.GetFundingsources()[0].Id, expectedFundingsources[0].ID)
		assert.Equal(st, response.GetFundingsources()[0].Name, expectedFundingsources[0].Name)
		assert.Equal(st, response.GetFundingsources()[0].Mask, expectedFundingsources[0].Mask)
		assert.Equal(st, response.GetFundingsources()[1].Id, expectedFundingsources[1].ID)
		assert.Equal(st, response.GetFundingsources()[1].Name, expectedFundingsources[1].Name)
		assert.Equal(st, response.GetFundingsources()[1].Mask, expectedFundingsources[1].Mask)
	})
}
