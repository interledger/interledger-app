package dev_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/external/client/dev"
)

func TestDevClient(t *testing.T) {
	client := dev.New()

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

	err = client.InitiateKyc(context.Background(), user.ID)
	require.NoError(t, err)

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
}
