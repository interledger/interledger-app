package external_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/env"
)

func TestNewClientValidation(t *testing.T) {
	validParams := struct {
		appID                  string
		secret                 string
		cardAppID              string
		gatewayID              string
		cardAccountProductCode string
		vaultID                string
		onOffRampClientID      string
		onboardingClientID     string
		exchangeClientID       string
		baseURL                string
		onboardingBaseURL      string
		onOffRampBaseURL       string
		organizationID         string
	}{
		appID:                  "test-app-id",
		secret:                 "test-secret",
		cardAppID:              "test-card-app-id",
		gatewayID:              "test-gateway-id",
		cardAccountProductCode: "test-product-code",
		vaultID:                "test-vault-id",
		onOffRampClientID:      "test-onoff-ramp-client-id",
		onboardingClientID:     "test-onboarding-client-id",
		exchangeClientID:       "test-exchange-client-id",
		baseURL:                "https://api.example.com",
		onboardingBaseURL:      "https://onboarding.example.com",
		onOffRampBaseURL:       "https://ramp.example.com",
		organizationID:         "test-organization-id",
	}

	tests := []struct {
		name                   string
		appID                  string
		secret                 string
		cardAppID              string
		gatewayID              string
		cardAccountProductCode string
		vaultID                string
		onOffRampClientID      string
		onboardingClientID     string
		exchangeClientID       string
		baseURL                string
		onboardingBaseURL      string
		onOffRampBaseURL       string
		organizationID         string
		expectNil              bool
	}{
		{
			name:                   "valid parameters",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                validParams.vaultID,
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         validParams.organizationID,
			expectNil:              false,
		},
		{
			name:                   "missing cardAccountProductCode",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: "",
			vaultID:                validParams.vaultID,
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         validParams.organizationID,
			expectNil:              true,
		},
		{
			name:                   "missing gatewayID",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              "",
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                validParams.vaultID,
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         validParams.organizationID,
			expectNil:              true,
		},
		{
			name:                   "missing vaultID",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                "",
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         validParams.organizationID,
			expectNil:              true,
		},
		{
			name:                   "missing onOffRampClientID",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                validParams.vaultID,
			onOffRampClientID:      "",
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         validParams.organizationID,
			expectNil:              true,
		},
		{
			name:                   "missing onboardingClientID",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                validParams.vaultID,
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     "",
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         validParams.organizationID,
			expectNil:              true,
		},
		{
			name:                   "missing exchangeClientID",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                validParams.vaultID,
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       "",
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         validParams.organizationID,
			expectNil:              true,
		},
		{
			name:                   "missing baseURL",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                validParams.vaultID,
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                "",
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         validParams.organizationID,
			expectNil:              true,
		},
		{
			name:                   "missing onboardingBaseURL",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                validParams.vaultID,
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      "",
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         validParams.organizationID,
			expectNil:              true,
		},
		{
			name:                   "missing onOffRampBaseURL",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                validParams.vaultID,
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       "",
			organizationID:         validParams.organizationID,
			expectNil:              true,
		},
		{
			name:                   "missing organizationID",
			appID:                  validParams.appID,
			secret:                 validParams.secret,
			cardAppID:              validParams.cardAppID,
			gatewayID:              validParams.gatewayID,
			cardAccountProductCode: validParams.cardAccountProductCode,
			vaultID:                validParams.vaultID,
			onOffRampClientID:      validParams.onOffRampClientID,
			onboardingClientID:     validParams.onboardingClientID,
			exchangeClientID:       validParams.exchangeClientID,
			baseURL:                validParams.baseURL,
			onboardingBaseURL:      validParams.onboardingBaseURL,
			onOffRampBaseURL:       validParams.onOffRampBaseURL,
			organizationID:         "",
			expectNil:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := external.NewClient(
				tt.appID,
				tt.secret,
				tt.cardAppID,
				tt.gatewayID,
				tt.cardAccountProductCode,
				tt.vaultID,
				tt.onOffRampClientID,
				tt.onboardingClientID,
				tt.exchangeClientID,
				tt.baseURL,
				tt.onboardingBaseURL,
				tt.onOffRampBaseURL,
				tt.organizationID,
				nil,
			)

			if tt.expectNil {
				assert.Nil(t, c, "expected client to be nil for missing parameter")
			} else {
				assert.NotNil(t, c, "expected client to be created with all parameters")
			}
		})
	}
}

func TestClient(t *testing.T) {
	env.SetEnv(t, "local")
	appID := os.Getenv("GATEHUB_APP_ID")
	secret := os.Getenv("GATEHUB_SECRET")
	cardAppID := os.Getenv("GATEHUB_CARD_APP_ID")
	gatewayID := os.Getenv("GATEHUB_GATEWAY_ID")
	cardAccountProductCode := os.Getenv("GATEHUB_CARD_ACCOUNT_PRODUCT_CODE")
	vaultID := os.Getenv("GATEHUB_PAYWISER_EURO_VAULT_ID")
	onOffRampClientID := os.Getenv("GATEHUB_ON_OFF_RAMP_CLIENT_ID")
	onboardingClientID := os.Getenv("GATEHUB_ONBOARDING_CLIENT_ID")
	exchangeClientID := os.Getenv("GATEHUB_EXCHANGE_CLIENT_ID")
	apiBaseURL := os.Getenv("GATEHUB_API_BASE_URL")
	onboardingBaseURL := os.Getenv("GATEHUB_ONBOARDING_BASE_URL")
	onOffRampBaseURL := os.Getenv("GATEHUB_ON_OFF_RAMP_BASE_URL")
	organizationID := os.Getenv("GATEHUB_ORGANIZATION_ID")

	if appID == "" || secret == "" || cardAppID == "" || gatewayID == "" || cardAccountProductCode == "" || vaultID == "" || onOffRampClientID == "" || onboardingClientID == "" || exchangeClientID == "" || apiBaseURL == "" || onboardingBaseURL == "" || onOffRampBaseURL == "" {
		t.SkipNow()
	}
	c := external.NewClient(appID, secret, cardAppID, gatewayID, cardAccountProductCode, vaultID, onOffRampClientID, onboardingClientID, exchangeClientID, apiBaseURL, onboardingBaseURL, onOffRampBaseURL, organizationID, nil)
	if c == nil {
		t.SkipNow()
	}

	ctx := context.Background()
	sendingExternalUserID := "66f1427e-43e4-48a0-9692-190c24d75058"
	// trx, err := c.CreateTransaction(ctx, external.CreateTransactionRequest{
	// 	SendingUserID:    sendingExternalUserID,
	// 	SendingAddress:   "107720301",
	// 	ReceivingAddress: "506541309", // belongs to "19227839-caa1-458f-a5ec-a3f03aa3e0e5"
	// 	Amount:           5.00,
	// 	VaultID:          "a09a0a2c-1a3a-44c5-a1b9-603a6eea9341",
	// 	Message:          "test transfer",
	// 	Type:             external.TransactionTypeHosted,
	// })
	// require.NoError(t, err)

	trx, err := c.GetTransaction(ctx, sendingExternalUserID, "b063d802-f0d7-463c-bc14-4310b6e313f4")
	require.NoError(t, err)

	temp, err := json.MarshalIndent(trx, "", " ")
	require.NoError(t, err)
	fmt.Printf("transaction: %s\n", temp)
}

func TestUser(t *testing.T) {
	env.SetEnv(t, "local")
	appID := os.Getenv("GATEHUB_APP_ID")
	secret := os.Getenv("GATEHUB_SECRET")
	cardAppID := os.Getenv("GATEHUB_CARD_APP_ID")
	gatewayID := os.Getenv("GATEHUB_GATEWAY_ID")
	cardAccountProductCode := os.Getenv("GATEHUB_CARD_ACCOUNT_PRODUCT_CODE")
	vaultID := os.Getenv("GATEHUB_PAYWISER_EURO_VAULT_ID")
	onOffRampClientID := os.Getenv("GATEHUB_ON_OFF_RAMP_CLIENT_ID")
	onboardingClientID := os.Getenv("GATEHUB_ONBOARDING_CLIENT_ID")
	exchangeClientID := os.Getenv("GATEHUB_EXCHANGE_CLIENT_ID")
	apiBaseURL := os.Getenv("GATEHUB_API_BASE_URL")
	onboardingBaseURL := os.Getenv("GATEHUB_ONBOARDING_BASE_URL")
	onOffRampBaseURL := os.Getenv("GATEHUB_ON_OFF_RAMP_BASE_URL")
	organizationID := os.Getenv("GATEHUB_ORGANIZATION_ID")

	if appID == "" || secret == "" || cardAppID == "" || gatewayID == "" || cardAccountProductCode == "" || vaultID == "" || onOffRampClientID == "" || onboardingClientID == "" || exchangeClientID == "" || apiBaseURL == "" || onboardingBaseURL == "" || onOffRampBaseURL == "" {
		t.SkipNow()
	}
	c := external.NewClient(appID, secret, cardAppID, gatewayID, cardAccountProductCode, vaultID, onOffRampClientID, onboardingClientID, exchangeClientID, apiBaseURL, onboardingBaseURL, onOffRampBaseURL, organizationID, nil)
	if c == nil {
		t.SkipNow()
	}
	ctx := context.Background()

	userID := "66f1427e-43e4-48a0-9692-190c24d75058"
	u, err := c.GetUser(ctx, userID)
	require.NoError(t, err)

	temp, err := json.MarshalIndent(u, "", " ")
	require.NoError(t, err)
	fmt.Printf("fetched user: %s\n", temp)
}
