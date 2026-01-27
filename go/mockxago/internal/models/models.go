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
	ID                 string    `db:"id"`
	WalletID           string    `db:"wallet_id"`
	AccountID          string    `db:"account_id"`
	DepositAddress     string    `db:"deposit_address"`
	DepositTag         int       `db:"deposit_tag"`
	DepositReference   string    `db:"deposit_reference"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

// Beneficiary represents a bank account beneficiary
type Beneficiary struct {
	ID           string    `db:"id"`
	WalletID     string    `db:"wallet_id"`
	BankName     string    `db:"bank_name"`
	AccountNumber string   `db:"account_number"`
	AccountName  string    `db:"account_name"`
	BranchCode   string    `db:"branch_code"`
	Currency     string    `db:"currency"`
	Country      string    `db:"country"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// Transaction represents a transfer transaction
type Transaction struct {
	ID            string    `db:"id"`
	WalletID      string    `db:"wallet_id"`
	BeneficiaryID string    `db:"beneficiary_id"`
	Amount        float64   `db:"amount"`
	Currency      string    `db:"currency"`
	Reference     string    `db:"reference"`
	Status        string    `db:"status"`
	CreatedAt     time.Time `db:"created_at"`
	SettledAt     *time.Time `db:"settled_at"`
}

// Deposit represents a fiat deposit transaction
type Deposit struct {
	ID               string    `db:"id"`
	AccountID        string    `db:"account_id"`
	Amount           float64   `db:"amount"`
	Currency         string    `db:"currency"`
	DepositReference string    `db:"deposit_reference"`
	Status           string    `db:"status"`
	Code             int       `db:"code"`
	CreatedAt        time.Time `db:"created_at"`
	SettledAt        *time.Time `db:"settled_at"`
}
