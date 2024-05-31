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
	HasMFA           bool
	CredentialsValid bool
}

type EasyEquitiesSession struct {
	page             playwright.Page
	hasMFA           bool
	credentialsValid bool
}

type EasyEquitiesDeposit struct {
	Hash        string
	Amount      float64
	Date        time.Time
	Description string
}
