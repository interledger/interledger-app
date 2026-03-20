package models

import "time"

// SubAccount represents a Chimoney multicurrency sub-account.
type SubAccount struct {
	ID          string
	ParentID    string
	UID         string
	Name        string
	Email       string
	FirstName   string
	LastName    string
	PhoneNumber string
	SubAccount  bool
	KYCStatus   string
	CreatedAt   time.Time
}

// Payment represents a Chimoney deposit payment record.
type Payment struct {
	IssueID     string
	SubAccount  string
	Amount      float64
	Currency    string
	Status      string
	PayerEmail  string
	RedirectURL string
	ChiRef      string
	CreatedAt   time.Time
}

// Payout represents a Chimoney withdrawal payout record.
type Payout struct {
	ID           string
	IssueID      string
	SubAccount   string
	Amount       float64
	Fee          float64
	Currency     string
	Status       string
	ChiRef       string
	InteracEmail string
	CreatedAt    time.Time
}
