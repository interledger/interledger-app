package ops

import (
	"time"

	"github.com/lib/pq"
)

type (
	CreateAuthURLArgs struct {
		WalletID     string
		ClientID     string
		Scopes       []string
		RedirectURL  string
		AuthEndpoint string
		State        string
	}

	discordAuthorization struct {
		ID          string         `db:"id"`
		WalletID    string         `db:"wallet_id"`
		State       string         `db:"state"`
		ClientID    string         `db:"client_id"`
		RedirectURL string         `db:"redirect_url"`
		Scopes      pq.StringArray `db:"scopes"`
		CreatedAt   time.Time      `db:"created_at"`
		UpdatedAt   time.Time      `db:"updated_at"`
	}
)
