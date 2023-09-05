package slack

import (
	"context"
	"time"

	"github.com/lib/pq"
)

type Client interface {
	CreateAuthURL(ctx context.Context, walletID string) (string, error)
	CreateConnection(ctx context.Context, args CreateConnectionArgs) (*Connection, error)
}

type Connection struct {
	ID           string         `db:"id"`
	UserID       string         `db:"user_id"`
	WalletID     string         `db:"wallet_id"`
	AccessToken  string         `db:"access_token"`
	RefreshToken string         `db:"refresh_token"`
	TokenType    string         `db:"token_type"`
	Scopes       pq.StringArray `db:"scopes"`
	Username     string         `db:"username"`
	TeamName     string         `db:"team_name"`
	TeamDomain   string         `db:"team_domain"`
	Expiry       time.Time      `db:"expiry"`
	ClientID     string         `db:"client_id"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

type CreateConnectionArgs struct {
	AuthCode string
	State    string
}
