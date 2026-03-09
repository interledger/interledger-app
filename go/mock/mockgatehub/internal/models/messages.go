package models

import "time"

// GetUserResponse represents the user state response from /id/v1/users/{userID}.
type GetUserResponse struct {
	ID            string             `json:"id"`
	Email         string             `json:"email"`
	Activated     bool               `json:"activated"`
	Managed       bool               `json:"managed"`
	Role          string             `json:"role"`
	Features      []string           `json:"features"`
	KYCState      string             `json:"kyc_state"`
	RiskLevel     string             `json:"risk_level"`
	CreatedAt     time.Time          `json:"created_at"`
	Profile       UserProfile        `json:"profile"`
	Verifications []UserVerification `json:"verifications"`
}

// UserProfile represents the profile payload returned by GateHub user state.
type UserProfile struct {
	UUID               string `json:"uuid"`
	BirthDay           int    `json:"birth_day"`
	BirthMonth         int    `json:"birth_month"`
	BirthYear          int    `json:"birth_year"`
	Gender             string `json:"gender"`
	FirstName          string `json:"first_name"`
	MiddleName         string `json:"middle_name"`
	LastName           string `json:"last_name"`
	Citizenship        string `json:"citizenship"`
	AddressPostalCode  string `json:"address_postal_code"`
	AddressSubdivision string `json:"address_subdivision"`
	AddressCountryCode string `json:"address_country_code"`
	AddressCity        string `json:"address_city"`
	AddressStreet1     string `json:"address_street1"`
	AddressStreet2     string `json:"address_street2"`
	BirthCity          string `json:"birth_city"`
	BirthCountryCode   string `json:"birth_country_code"`
	TaxResidency       string `json:"tax_residency"`
	ExpectedVolume     string `json:"expected_volume"`
}

// UserVerification represents a verification entry in the user response.
type UserVerification struct {
	UUID         string `json:"uuid"`
	Status       int    `json:"status"`
	State        int    `json:"state"`
	ProviderType string `json:"provider_type"`
}

// UserWalletsResponse represents /core/v1/users/{userUuid} response.
type UserWalletsResponse struct {
	Wallets []UserWalletResponse `json:"wallets"`
}

// UserWalletResponse represents a wallet entry returned by GateHub core API.
type UserWalletResponse struct {
	UUID    string `json:"uuid"`
	Address string `json:"address"`
	Name    string `json:"name"`
	Type    int    `json:"type"`
	Primary bool   `json:"primary"`
	Active  bool   `json:"active"`
	Enabled bool   `json:"enabled"`
}

// WalletBalanceResponse represents a balance entry for /core/v1/wallets/{address}/balances.
type WalletBalanceResponse struct {
	Available string       `json:"available"`
	Pending   string       `json:"pending"`
	Total     string       `json:"total"`
	Vault     VaultSummary `json:"vault"`
}

// VaultSummary represents vault metadata within a balance response.
type VaultSummary struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	AssetCode string `json:"asset_code"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CurrenciesResponse represents the currency list response.
type CurrenciesResponse struct {
	Currencies []string `json:"currencies"`
}

// DepositWebhookData represents the data payload for deposit webhooks.
// Wallet backend expects GateHub parity fields: tx_uuid + deposit_type + address.
type DepositWebhookData struct {
	TransactionID string `json:"transaction_id"` // legacy field (kept for compatibility)
	TxUUID        string `json:"tx_uuid"`        // primary transaction identifier used by wallet backend
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Address       string `json:"address,omitempty"`
	DepositType   string `json:"deposit_type,omitempty"` // hosted | external
}
