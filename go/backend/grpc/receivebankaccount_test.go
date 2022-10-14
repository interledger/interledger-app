package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	_user "gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetBanks(t *testing.T) {
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

	c.KYCClient.EXPECT().GetIndividualDetails(gomock.Any(), wallet.ID).Return(
		&kyc.IndividualDetails{
			WalletID:    wallet.ID,
			CountryCode: "USD",
		},
		nil,
	).Times(1)
	c.machnet.EXPECT().GetBanks(gomock.Any(), "USD").Return(
		[]machnet.Bank{
			{
				ID:   1,
				Name: "Monopoly",
				Branches: []machnet.Branch{
					{
						ID:   2,
						Name: "Board",
					},
				},
			},
			{
				ID:   2,
				Name: "Last",
				Branches: []machnet.Branch{
					{
						ID:   1,
						Name: "Place",
					},
				},
			},
		},
		nil,
	).Times(1)

	banks, err := client.ListBanks(user_mock.ActingAsContext(t, context.Background(), user), &backendv1.Empty{})
	require.NoError(t, err)

	require.Len(t, banks.GetBanks(), 2)
	assert.Equal(t, uint32(1), banks.GetBanks()[0].Id)
	assert.Equal(t, "Monopoly", banks.GetBanks()[0].Name)
	require.Len(t, banks.GetBanks()[0].GetBranches(), 1)
	assert.Equal(t, uint32(2), banks.GetBanks()[0].GetBranches()[0].Id)
	assert.Equal(t, "Board", banks.GetBanks()[0].GetBranches()[0].Name)

	assert.Equal(t, uint32(2), banks.GetBanks()[1].Id)
	assert.Equal(t, "Last", banks.GetBanks()[1].Name)
	require.Len(t, banks.GetBanks()[1].GetBranches(), 1)
	assert.Equal(t, uint32(1), banks.GetBanks()[1].GetBranches()[0].Id)
	assert.Equal(t, "Place", banks.GetBanks()[1].GetBranches()[0].Name)
}

func TestCreateReceiveBankAccount(t *testing.T) {
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

	requestArgs := backendv1.CreateReceiveBankAccountRequest{
		Name:          "Test",
		BankId:        1,
		BranchId:      2,
		AccountType:   "SAVINGS",
		AccountNumber: "123456",
	}
	linkedAccountID := uuid.NewString()
	receiveBankAccountID := uuid.NewString()
	c.machnet.EXPECT().CreateReceiveBankAccount(gomock.Any(), machnet.CreateReceiveBankAccountArgs{
		WalletID:      wallet.ID,
		AccountNumber: requestArgs.AccountNumber,
		BankID:        requestArgs.BankId,
		BranchID:      requestArgs.BranchId,
		Name:          "Test",
	}).Return(
		&machnet.ReceiveBankAccount{
			ID:            receiveBankAccountID,
			WalletID:      wallet.ID,
			AccountNumber: requestArgs.AccountNumber,
			BankID:        requestArgs.BankId,
			BranchID:      requestArgs.BranchId,
		},
		nil,
	).Times(1)
	c.linkedaccounts.EXPECT().GetByProviderID(gomock.Any(), linkedaccounts.GetByProviderIDArgs{
		Provider:   machnet.ProviderName,
		ProviderID: receiveBankAccountID,
		Type:       machnet.TypeReceiveBankAccount,
	}).Return(
		&linkedaccounts.LinkedAccount{
			ID:         linkedAccountID,
			WalletId:   wallet.ID,
			Mask:       "3456",
			Provider:   machnet.ProviderName,
			ProviderID: receiveBankAccountID,
			Type:       machnet.TypeReceiveBankAccount,
			Name:       requestArgs.Name,
		},
		nil,
	).Times(1)

	linkedAccount, err := client.CreateReceiveBankAccount(
		user_mock.ActingAsContext(t, context.Background(), user),
		&requestArgs,
	)
	require.NoError(t, err)

	assert.Equal(t, requestArgs.Name, linkedAccount.GetName())
	assert.Equal(t, linkedAccountID, linkedAccount.GetId())
	assert.Equal(t, "3456", linkedAccount.GetMask())
	assert.Equal(t, machnet.TypeReceiveBankAccount, linkedAccount.GetType())
}
