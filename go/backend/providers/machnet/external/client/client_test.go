package client_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/external/client"
)

func TestMachnetClientIntegration(t *testing.T) {
	clientID := os.Getenv("MACHNET_CLIENT_ID")
	clientSecret := os.Getenv("MACHNET_CLIENT_SECRET")
	if clientID == "" && clientSecret == "" {
		t.Skip("Skipping machnet external client integration set as credentials aren't set.")
	}
	t.Setenv("FYNBOS_ENV", "sandbox")

	client := client.New(clientID, clientSecret)
	createdUser, err := client.RegisterUser(context.Background(), external.User{
		FirstName:    "Tenzin",
		LastName:     "Norgay",
		Email:        faker.Email(),
		Gender:       "female",
		DateOfBirth:  "2000-01-01",
		AddressLine1: "500 8 El Camino Real Santa Clara",
		MobilePhone:  "9879879870",
		City:         "Clara",
		Zipcode:      "95053",
		State:        "CA",
		Country:      "US",
		IPAddress:    "73.85.79.9",
		Business:     false,
		Type:         external.TypeSendUser,
	})
	require.NoError(t, err)

	user, err := client.GetUserByID(context.Background(), createdUser.ID)
	require.NoError(t, err)
	require.Equal(t, "Tenzin", user.FirstName)

	_, err = client.UpdateUser(context.Background(), user.ID, external.User{
		Gender:       "female",
		Country:      "US",
		State:        "CA",
		AddressLine1: "500 8 El Camino Real Santa Clara",
	})
	require.NoError(t, err)

	_, err = client.InitiateKYC(context.Background(), user.ID)
	require.NoError(t, err)

	verificationStatus, err := client.GetVerificationStatus(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, verificationStatus.UserID)

	if verificationStatus.KycStatus == external.StatusVerified {
		widgetToken, err := client.GetFundingAccountWidgetToken(context.Background(), user.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, widgetToken)
	} else {
		log.Println("Unable to get widget token as user is not verified.")
	}

	banks, err := client.GetBanks(context.Background(), "GH")
	require.NoError(t, err)
	require.Len(t, banks, 1)
	assert.Equal(t, "Zenith Bank", banks[0].Name)
	assert.Equal(t, uint32(37), banks[0].ID)
	require.Len(t, banks[0].Branches, 3)
	assert.Equal(t, uint32(37), banks[0].Branches[0].ID)
	assert.Equal(t, "Zenith Bank", banks[0].Branches[0].Name)
}
