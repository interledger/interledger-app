package discord

import (
	"time"

	"github.com/lib/pq"
)

type (
	CreateAuthURLArgs struct {
		Scopes   []string
		WalletID string
	}

	User struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
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
		GlobalName   string         `db:"global_name"`
		Expiry       time.Time      `db:"expiry"`
		ClientID     string         `db:"client_id"`
		CreatedAt    time.Time      `db:"created_at"`
		UpdatedAt    time.Time      `db:"updated_at"`
	}

	CreateConnectionArgs struct {
		AuthCode string
		State    string
	}
)
