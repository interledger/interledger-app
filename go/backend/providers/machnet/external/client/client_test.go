package client_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/external/client"
)

func TestFormatIPAddress(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "ipv4",
			input:  "127.0.0.1",
			output: "127.0.0.1",
		},
		{
			name:   "ipv6",
			input:  "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			output: "0.0.0.0",
		},
		{
			name:   "random string",
			input:  "Ladida",
			output: "0.0.0.0",
		},
		{
			name:   "empty string",
			input:  "",
			output: "0.0.0.0",
		},
		{
			name:   "invalid ip",
			input:  "266.266.266.365",
			output: "0.0.0.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := client.FormatIPAddress(tc.input)
			assert.Equal(t, tc.output, ip)
		})
	}
}

func TestMachnetClientIntegration(t *testing.T) {
	envFile := os.Getenv("ENV_FILE")
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			t.Fatal("Error loading .env file")
		}
	}
	clientID := os.Getenv("MACHNET_CLIENT_ID")
	clientSecret := os.Getenv("MACHNET_CLIENT_SECRET")
	if clientID == "" && clientSecret == "" {
		t.Skip("Skipping machnet external client integration set as credentials aren't set.")
	}

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

	banks, err := client.GetBanks(context.Background(), "GH")
	require.NoError(t, err)
	require.Len(t, banks, 1)
	assert.Equal(t, "Zenith Bank", banks[0].Name)
	assert.Equal(t, uint32(37), banks[0].ID)
	require.Len(t, banks[0].Branches, 3)
	assert.Equal(t, uint32(37), banks[0].Branches[0].ID)
	assert.Equal(t, "Zenith Bank", banks[0].Branches[0].Name)

	retry := 0
	kycStatus := ""
	for {
		if retry > 5 {
			break
		}
		verificationStatus, err := client.GetVerificationStatus(context.Background(), user.ID)
		require.NoError(t, err)
		require.Equal(t, user.ID, verificationStatus.UserID)

		kycStatus = verificationStatus.KycStatus
		if kycStatus == external.StatusVerified {
			break
		}

		log.Println("KycStatus=", kycStatus, "retrying in 3 seconds")
		time.Sleep(3 * time.Second)
		retry++
	}

	if kycStatus != external.StatusVerified {
		log.Println("Rest of test requires VERIFIED kyc status. KycStatus=", kycStatus)
		return
	}

	widgetToken, err := client.GetFundingAccountWidgetToken(context.Background(), user.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, widgetToken)

	// Create and get user wallet
	walletNickName := faker.Name()
	wallet, err := client.CreateUserWallet(context.Background(), user.ID, walletNickName)
	require.NoError(t, err)
	assert.NotEqual(t, "", wallet.ID)
	assert.Equal(t, walletNickName, wallet.NickName)
	assert.Equal(t, user.ID, wallet.UserID)
	assert.Equal(t, external.StatusVerified, wallet.VerificationStatus)
	assert.Equal(t, external.TypeWallet, wallet.FundingSourceType)
	assert.Equal(t, float64(0), wallet.Balance.Balance)
	assert.Equal(t, float64(0), wallet.Balance.AvailableBalance)

	getWallet, err := client.GetUserWallet(context.Background(), user.ID, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, wallet.ID, getWallet.ID)
	assert.Equal(t, walletNickName, getWallet.NickName)
	assert.Equal(t, user.ID, getWallet.UserID)
	assert.Equal(t, external.StatusVerified, getWallet.VerificationStatus)
	assert.Equal(t, external.TypeWallet, getWallet.FundingSourceType)
	assert.Equal(t, float64(0), getWallet.Balance.Balance)
	assert.Equal(t, float64(0), getWallet.Balance.AvailableBalance)
}
