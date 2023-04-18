package ops

import "time"

type (
	CreateAuthURLArgs struct {
		WalletID     string
		ClientID     string
		Scopes       []string
		RedirectURL  string
		AuthEndpoint string
	}

	CreateAccessTokenArgs struct {
		AuthCode string
		State    string
	}

	TwitterAuth struct {
		ID           string    `db:"id"`
		WalletID     string    `db:"wallet_id"`
		State        string    `db:"state"`
		CodeVerifier string    `db:"code_verifier"`
		ClientID     string    `db:"client_id"`
		RedirectURL  string    `db:"redirect_url"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	TwitterAccessToken struct {
		ID           string    `db:"id"`
		WalletID     string    `db:"wallet_id"`
		AccessToken  string    `db:"access_token"`
		RefreshToken string    `db:"refresh_token"`
		TokenType    string    `db:"token_type"`
		Expiry       time.Time `db:"expiry"`
		ClientID     string    `db:"client_id"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	Authorization struct {
		URL   string
		State string
	}
)
