package linkedin

import (
	"github.com/lib/pq"
	"time"
)

type (
	CreateAuthURLArgs struct {
		State    string
		Scopes   []string
		WalletID string
	}

	CreateConnectionArgs struct {
		AuthCode string
		State    string
	}

	Connection struct {
		ID           string         `db:"id"`
		UserID       string         `db:"user_id"`
		WalletID     string         `db:"wallet_id"`
		AccessToken  string         `db:"access_token"`
		RefreshToken string         `db:"refresh_token"`
		TokenType    string         `db:"token_type"`
		Scopes       pq.StringArray `db:"scopes"`
		Username     string         `db:"username"`
		Expiry       time.Time      `db:"expiry"`
		ClientID     string         `db:"client_id"`
		CreatedAt    time.Time      `db:"created_at"`
		UpdatedAt    time.Time      `db:"updated_at"`
	}

	User struct {
		ID       string `json:"id"`
		Username string `json:"vanityName"`
	}
)
