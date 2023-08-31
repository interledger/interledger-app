package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gitlab.com/fynbos/backend/discord"
	"gitlab.com/fynbos/backend/discord/external"
	"golang.org/x/oauth2"
)

func CreateAuthURL(ctx context.Context, b Backends, args CreateAuthURLArgs) (string, error) {
	oauthConfig := &oauth2.Config{
		ClientID:    args.ClientID,
		RedirectURL: args.RedirectURL,
		Scopes:      args.Scopes,
		Endpoint:    oauth2.Endpoint{AuthURL: args.AuthEndpoint},
	}

	var authId string
	err := b.DB().GetContext(ctx, &authId, "INSERT INTO discord_authorizations (client_id, state, scopes, wallet_id, redirect_url) VALUES ($1, $2, $3, $4, $5) RETURNING id", args.ClientID, args.State, pq.Array(args.Scopes), args.WalletID, args.RedirectURL)
	if err != nil {
		return "", fmt.Errorf("%w %s", discord.ErrInternal, err)
	}

	url := oauthConfig.AuthCodeURL(args.State)

	return url, nil
}

func CreateConnection(ctx context.Context, b Backends, args discord.CreateConnectionArgs) (*discord.Connection, error) {
	var authorization discordAuthorization
	err := b.DB().GetContext(ctx, &authorization, "SELECT * FROM discord_authorizations WHERE state = $1", args.State)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w no authorization found for state %q", discord.ErrNotFound, args.State)
		}
		return nil, fmt.Errorf("%w %s", discord.ErrInternal, err)
	}
	if args.AuthCode == "" {
		return nil, fmt.Errorf("%w couldn't find an auth code for this grant", discord.ErrInternal)
	}

	token, err := b.External().CreateToken(ctx, &external.CreateTokenArgs{
		AuthCode: args.AuthCode,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", discord.ErrInternal, err)
	}

	user, err := b.External().GetAuthorizedUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("%w %s", discord.ErrInternal, err)
	}

	var connection discord.Connection
	query := `
		INSERT INTO discord_connections (user_id, wallet_id, access_token, refresh_token, token_type, scopes, username, expiry, client_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, wallet_id) 
		DO UPDATE SET 
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_type = EXCLUDED.token_type,
			scopes = EXCLUDED.scopes,
			username = EXCLUDED.username,
			expiry = EXCLUDED.expiry,
			updated_at = NOW()
		RETURNING *;
	`
	err = b.DB().GetContext(ctx, &connection, query,
		user.ID, authorization.WalletID, token.AccessToken, token.RefreshToken, token.TokenType, pq.Array(authorization.Scopes), user.Username, token.Expiry, authorization.ClientID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", discord.ErrInternal, err)
	}

	return &connection, nil
}

func GetWalletConnections(ctx context.Context, b Backends, id string) ([]discord.Connection, error) {
	var connections []discord.Connection

	err := b.DB().SelectContext(ctx, &connections, "SELECT * FROM discord_connections WHERE wallet_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", discord.ErrInternal, err)
	}

	return connections, nil
}
