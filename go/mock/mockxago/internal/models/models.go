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
