package ops

import (
	"time"

	"github.com/playwright-community/playwright-go"
)

type CredentialsCheckRequest struct {
	WealthUserID int64  `json:"user_id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

type CredentialsCheckResponse struct {
	HasMFA           bool `json:"has_mfa"`
	CredentialsValid bool `json:"credentials_valid"`
}

type EasyEquitiesSession struct {
	page             playwright.Page
	hasMFA           bool
	credentialsValid bool
}

type EasyEquitiesDeposit struct {
	Hash        string    `json:"hash"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}
