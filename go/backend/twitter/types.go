package twitter

import (
	"time"

	"github.com/lib/pq"
)

type (
	CreateAuthURLArgs struct {
		State    string
		Scopes   []string
		WalletID string
	}

	TwitterUser struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}

	Connection struct {
		ID           string         `db:"id"`
		UserID       string         `db:"user_id"`
		WalletID     string         `db:"wallet_id"`
		AccessToken  string         `db:"access_token"`
		RefreshToken string         `db:"refresh_token"`
		TokenType    string         `db:"token_type"`
		Scopes       pq.StringArray `db:"scopes"`
		Expiry       time.Time      `db:"expiry"`
		ClientID     string         `db:"client_id"`
		CreatedAt    time.Time      `db:"created_at"`
		UpdatedAt    time.Time      `db:"updated_at"`
	}

	CreateConnectionArgs struct {
		AuthCode string
		State    string
	}

	GetTokensByUserIdArgs struct {
		Scopes   []string
		UserID   string
		WalletID string
	}

	Tweet struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}
)
