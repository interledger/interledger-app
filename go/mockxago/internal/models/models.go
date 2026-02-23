package models

import "time"

// AccessToken represents an authentication token
type AccessToken struct {
	ID        string    `db:"id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// IsExpired checks if the token has expired
func (at *AccessToken) IsExpired() bool {
	return time.Now().After(at.ExpiresAt)
}

// IsValid checks if the token is valid and not expired
func (at *AccessToken) IsValid() bool {
	return at.Token != "" && !at.IsExpired()
}

// SubAccount represents a Xago sub-account
type SubAccount struct {
	ID                        string    `db:"id"`
	WalletID                  string    `db:"wallet_id"`
	AccountID                 string    `db:"account_id"`
	FirstName                 string    `db:"first_name"`
	LastName                  string    `db:"last_name"`
	Email                     string    `db:"email"`
	MobileNumber              string    `db:"mobile_number"`
	IdentityType              string    `db:"identity_type"`
	IDNumber                  string    `db:"id_number"`
	PhysicalAddress           string    `db:"physical_address"`
	ThirdPartyVerificationURL string    `db:"third_party_verification_url"`
	DepositAddress            string    `db:"deposit_address"`
	DepositTag                int       `db:"deposit_tag"`
	DepositReferenceZAR       string    `db:"deposit_reference_zar"`
	DepositReferenceUSD       string    `db:"deposit_reference_usd"`
	CreatedAt                 time.Time `db:"created_at"`
	UpdatedAt                 time.Time `db:"updated_at"`
}

// Beneficiary represents a bank account beneficiary
type Beneficiary struct {
	ID            string    `db:"id"`
	AccountID     string    `db:"account_id"`
	WalletID      string    `db:"wallet_id"`
	Name          string    `db:"name"`
	Scope         string    `db:"scope"`
	IsOwn         bool      `db:"is_own"`
	BankName      string    `db:"bank_name"`
	AccountNumber string    `db:"account_number"`
	AccountName   string    `db:"account_name"`
	BranchCode    string    `db:"branch_code"`
	Reference     string    `db:"reference"`
	Currency      string    `db:"currency"`
	Country       string    `db:"country"`
	Status        string    `db:"status"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// Transaction represents a transfer transaction
type Transaction struct {
	ID            string     `db:"id"`
	AccountID     string     `db:"account_id"`
	WalletID      string     `db:"wallet_id"`
	BeneficiaryID string     `db:"beneficiary_id"`
	Amount        float64    `db:"amount"`
	Currency      string     `db:"currency"`
	Reference     string     `db:"reference"`
	Status        string     `db:"status"`
	CreatedAt     time.Time  `db:"created_at"`
	SettledAt     *time.Time `db:"settled_at"`
}

// Deposit represents a fiat deposit transaction
type Deposit struct {
	ID               string     `db:"id"`
	AccountID        string     `db:"account_id"`
	Amount           float64    `db:"amount"`
	Currency         string     `db:"currency"`
	DepositReference string     `db:"deposit_reference"`
	Status           string     `db:"status"`
	Code             int        `db:"code"`
	CreatedAt        time.Time  `db:"created_at"`
	SettledAt        *time.Time `db:"settled_at"`
}
