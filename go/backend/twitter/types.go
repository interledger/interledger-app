package twitter

import (
	"time"

	"github.com/lib/pq"
)

type (
	CreateAuthURLArgs struct {
		Scopes   []string
		WalletID string
	}

	Authorization struct {
		URL   string
		State string
	}

	TwitterAccessToken struct {
		ID           string         `db:"id"`
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

	CreateAccessTokenArgs struct {
		AuthCode string
		State    string
	}
)
