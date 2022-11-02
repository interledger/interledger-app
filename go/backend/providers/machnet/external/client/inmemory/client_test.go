package inmemory_test

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/external/client/inmemory"
)

func TestInMemoryClient(t *testing.T) {
	client := inmemory.New()

	// crud
	user, err := client.RegisterUser(context.Background(), external.User{
		FirstName:    "Pickle",
		LastName:     "Rick",
		Gender:       "male",
		AddressLine1: "Netflix",
		Email:        "its@pickle.rick",
		Type:         "SEND",
	})
	require.NoError(t, err)
	require.Equal(t, "Pickle", user.FirstName)
	require.Equal(t, "Rick", user.LastName)
	require.Equal(t, external.StatusUnverified, user.Status)

	updatedUser, err := client.UpdateUser(context.Background(), user.ID, external.User{
		FirstName: "Morty",
		Email:     "its@morty.rick",
	})
	require.NoError(t, err)
	require.Equal(t, "Morty", updatedUser.FirstName)
	require.Equal(t, "Rick", updatedUser.LastName)
	require.Equal(t, "male", updatedUser.Gender)
	require.Equal(t, "Netflix", updatedUser.AddressLine1)
	require.Equal(t, "its@morty.rick", updatedUser.Email)
	require.Equal(t, external.StatusUnverified, updatedUser.Status)

	freshUser, err := client.GetUserByID(context.Background(), updatedUser.ID)
	require.NoError(t, err)
	require.Equal(t, "Morty", freshUser.FirstName)
	require.Equal(t, "Rick", freshUser.LastName)
	require.Equal(t, "male", freshUser.Gender)
	require.Equal(t, "Netflix", freshUser.AddressLine1)
	require.Equal(t, "its@morty.rick", freshUser.Email)
	require.Equal(t, external.StatusUnverified, freshUser.Status)

	// kyc
	kycStatus, err := client.GetVerificationStatus(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, external.StatusUnverified, kycStatus.KycStatus)

	kycResponse, err := client.InitiateKYC(context.Background(), user.ID)
	require.NoError(t, err)
	require.True(t, kycResponse.Success)
	require.Equal(t, external.StatusVerified, kycResponse.Status)

	kycStatus, err = client.GetVerificationStatus(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, external.StatusVerified, kycStatus.KycStatus)
	require.Equal(t, external.StatusVerified, kycStatus.CipInfo.FirstName)

	// receive users
	receiveUser, err := client.RegisterUser(context.Background(), external.User{
		FirstName:  "James",
		Type:       "RECEIVE",
		SendUserID: user.ID,
	})
	require.NoError(t, err)
	require.Equal(t, user.ID, receiveUser.SendUserID)
	require.Equal(t, "RECEIVE", receiveUser.Type)

	receiveUsers, err := client.GetReceiveUserList(context.Background(), user.ID)
	require.NoError(t, err)
	require.Len(t, receiveUsers, 1)
	require.Equal(t, receiveUser.ID, receiveUsers[0].ID)

	// wallets
	nickname := faker.Name()
	wallet, err := client.CreateUserWallet(context.Background(), user.ID, nickname)
	require.NoError(t, err)
	require.Equal(t, nickname, wallet.NickName)
	require.Equal(t, user.ID, wallet.UserID)

	getWallet, err := client.GetUserWallet(context.Background(), user.ID, wallet.ID)
	require.NoError(t, err)
	require.Equal(t, nickname, getWallet.NickName)
	require.Equal(t, user.ID, getWallet.UserID)

	// withdraw from wallets
	_, err = client.WithdrawFromUserWallet(context.Background(), external.WithdrawFromUserWalletArgs{
		UserID:    user.ID,
		ToFundID:  uuid.NewString(),
		WalletID:  getWallet.ID,
		Amount:    10,
		FeeAmount: 0,
		Currency:  "USD",
		IPAddress: "10.10.10.10",
	})
	require.ErrorIs(t, err, external.ErrInternal)
}
