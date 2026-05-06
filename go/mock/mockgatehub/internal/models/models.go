package models

import "time"

// User represents a Gatehub user
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Activated bool      `json:"activated"`
	Managed   bool      `json:"managed"`
	Role      string    `json:"role"`
	Features  []string  `json:"features"`
	KYCState  string    `json:"kyc_state"`  // accepted/rejected/action_required
	RiskLevel string    `json:"risk_level"` // low/medium/high
	CreatedAt time.Time `json:"created_at"`
	// Profile fields
	FirstName          string `json:"first_name"`
	MiddleName         string `json:"middle_name"`
	LastName           string `json:"last_name"`
	Gender             string `json:"gender"`
	BirthYear          int    `json:"birth_year"`
	BirthMonth         int    `json:"birth_month"`
	BirthDay           int    `json:"birth_day"`
	BirthCity          string `json:"birth_city"`
	BirthCountryCode   string `json:"birth_country_code"`
	Citizenship        string `json:"citizenship"`
	AddressStreet1     string `json:"address_street1"`
	AddressStreet2     string `json:"address_street2"`
	AddressCity        string `json:"address_city"`
	AddressPostalCode  string `json:"address_postal_code"`
	AddressSubdivision string `json:"address_subdivision"`
	AddressCountryCode string `json:"address_country_code"`
	TaxResidency       string `json:"tax_residency"`
	ExpectedVolume     string `json:"expected_volume"`
}

// Wallet represents an XRPL wallet
type Wallet struct {
	Address   string    `json:"address"` // Mock XRPL address
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Type      int       `json:"type"`
	Network   int       `json:"network"` // 30 for XRP Ledger
	CreatedAt time.Time `json:"created_at"`
}

// Organization represents organization-level configuration
type Organization struct {
	ID         string    `json:"id"`
	APIBaseURL string    `json:"apiBaseUrl"`
	TwoFAType  string    `json:"type2fa"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Transaction represents a deposit or transaction
type Transaction struct {
	ID               string    `json:"uuid"` // GateHub uses "uuid" not "id"
	UserID           string    `json:"user_id"`
	UID              string    `json:"uid"`          // External reference
	Amount           string    `json:"amount"`       // String to match GateHub API
	TotalAmount      string    `json:"total_amount"` // Total including fees
	Fee              string    `json:"fee"`          // Fee amount
	Currency         string    `json:"currency"`
	VaultUUID        string    `json:"vault_uuid"`
	SendingAddress   string    `json:"sending_address"`
	ReceivingAddress string    `json:"receiving_address"`
	Type             int       `json:"type"`         // 1=deposit, 2=hosted
	DepositType      string    `json:"deposit_type"` // external/hosted
	Status           int       `json:"status"`       // 0=pending, 1=completed, 2=failed
	CreatedAt        time.Time `json:"created_at"`
}
