package gatehub

// Test Constants for Gatehub Sandbox Environment
//
// These constants are used in unit and integration tests.
// They represent the sandbox environment configuration.

const (
	// Sandbox Client IDs for Gatehub widgets and services
	TestOnOffRampClientID  = "f8119dfd-e563-44ee-9ae2-1e60a4fce74f"
	TestOnboardingClientID = "4df24d1b-5796-4eec-951b-21699d61b970"
	TestExchangeClientID   = "4e28d4df-22d7-414c-97a3-d71956df29ba"

	// Sandbox API Base URLs
	TestAPIBaseURL        = "https://api.sandbox.gatehub.net"
	TestOnboardingBaseURL = "https://onboarding.sandbox.gatehub.net"
	TestOnOffRampBaseURL  = "https://managed-ramp.sandbox.gatehub.net"

	// EUR Operations Account ID
	TestEUROpsAccount = "1854f171-eafa-4e30-bf66-7dbfe167ccfa"

	// EUR Ledger ID (spells "ghubeur" on a Nokia 3320 keyboard)
	TestEUROpsLedgerID uint32 = 4482387

	// Sandbox test credentials
	TestAppID                  = "sandbox-test-app-id"
	TestSecret                 = "sandbox-test-secret"
	TestCardAppID              = "sandbox-test-card-app-id"
	TestGatewayID              = "test-gateway-id"
	TestCardAccountProductCode = "PWSR_DEBP_2404"
	TestPaywiserEuroVaultID    = "a09a0a2c-1a3a-44c5-a1b9-603a6eea9341"
	TestSendingUserID          = "test-sending-user-id"
	TestSendingUserAddress     = "rN7n7otQDd6FczFgLdVZGMbpKRtFVfT4hb"
	TestWebhookSecret          = "test-webhook-secret"
	TestFallbackWebhookURL     = "http://localhost:3000/webhooks/gatehub"
)
