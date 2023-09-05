package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gitlab.com/fynbos/backend/slack"
)

func CreateAuthURL(ctx context.Context, b Backends, walletID string) (string, error) {
	state := uuid.NewString()
	nonce := uuid.NewString()

	_, err := b.DB().ExecContext(ctx, "INSERT INTO slack_authorizations (wallet_id, client_id, nonce, state, scopes, redirect_url) VALUES ($1, $2, $3, $4, $5, $6)",
		walletID, b.External().GetConfig().ClientID, nonce, state, pq.Array(b.External().GetConfig().Scopes), b.External().GetConfig().RedirectURL)
	if err != nil {
		return "", err
	}

	return b.External().GetConfig().AuthCodeURL(state, oidc.Nonce(nonce)), nil
}

type authorization struct {
	ID          string         `db:"id"`
	WalletID    string         `db:"wallet_id"`
	State       string         `db:"state"`
	Nonce       string         `db:"nonce"`
	ClientID    string         `db:"client_id"`
	RedirectURL string         `db:"redirect_url"`
	Scopes      pq.StringArray `db:"scopes"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func CreateConnection(ctx context.Context, b Backends, args slack.CreateConnectionArgs) (*slack.Connection, error) {
	var auth authorization
	err := b.DB().GetContext(ctx, &auth, "SELECT * FROM slack_authorizations WHERE state = $1", args.State)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w no authorization found for state %q", slack.ErrNotFound, args.State)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", slack.ErrInternal, err)
	}

	token, user, err := b.External().CreateUserToken(ctx, args.AuthCode)
	if err != nil {
		return nil, fmt.Errorf("%w %s", slack.ErrInternal, err)
	}

	var connection slack.Connection
	query := `
		INSERT INTO slack_connections (user_id, wallet_id, access_token, refresh_token, token_type, scopes, username, team_name, team_domain, expiry, client_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (user_id, wallet_id, team_domain) 
		DO UPDATE SET 
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_type = EXCLUDED.token_type,
			scopes = EXCLUDED.scopes,
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
		  	team_domain = EXCLUDED.team_domain,
			expiry = EXCLUDED.expiry,
			updated_at = NOW()
		RETURNING *;
	`
	err = b.DB().GetContext(ctx, &connection, query,
		user.ID, auth.WalletID, token.AccessToken, token.RefreshToken, token.TokenType, pq.Array(auth.Scopes), user.Username, user.TeamName, user.TeamDomain, token.Expiry, auth.ClientID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", slack.ErrInternal, err)
	}

	return &connection, nil
}
