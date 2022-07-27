package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/deposits"
	"gitlab.com/fynbos/backend/user"
	_user "gitlab.com/fynbos/backend/user"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestInitiateDeposit(t *testing.T) {
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
	accountID := uuid.NewString()
	c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), gomock.Any()).Return(
		&accounts.Account{
			ID:         accountID,
			IdentityID: user.ID,
		},
		nil,
	).AnyTimes()
	depositID := uuid.NewString()
	fundingsourceID := uuid.NewString()
	c.DepositService.EXPECT().InitiateDeposit(gomock.Any(), &deposits.InitiateDepositArgs{
		IdentityID:      user.ID,
		AccountID:       accountID,
		FundingSourceID: fundingsourceID,
		Amount:          1000,
	}).Return(
		&deposits.Deposit{
			ID: depositID,
		},
		nil,
	).AnyTimes()

	cases := []struct {
		Name            string
		ExpectedError   string
		FundingsourceID string
		User            *_user.User
	}{
		{
			Name:            "Requires authenicated user",
			User:            nil,
			FundingsourceID: fundingsourceID,
			ExpectedError:   "rpc error: code = Unauthenticated desc = Unauthenticated: no user.",
		},
		{
			Name:            "Initiates deposit",
			User:            user,
			ExpectedError:   "",
			FundingsourceID: fundingsourceID,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(st *testing.T) {
			response, err := client.InitiateDeposit(
				_user.ActingAsContext(st, context.Background(), tc.User),
				&backendv1.InitiateDepositRequest{
					FundingsourceId: tc.FundingsourceID,
					Amount:          1000,
				})

			if tc.ExpectedError == "" {
				assert.NoError(st, err, tc.Name)
				assert.Equal(st, depositID, response.GetDepositId())
			} else {
				assert.Equal(st, "", response.GetDepositId())
				assert.Equal(st, tc.ExpectedError, err.Error())
			}
		})
	}
}
