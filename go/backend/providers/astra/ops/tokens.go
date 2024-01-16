package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/astra/external"
)

type dbToken struct {
	ID               string    `db:"id"`
	Token            string    `db:"token"`
	WalletID         string    `db:"wallet_id"`
	ExpiresAt        time.Time `db:"expires_at"`
	RefreshToken     string    `db:"refresh_token"`
	RefreshExpiresAt time.Time `db:"refresh_expires_at"`
}

func CreateOrRefreshToken(ctx context.Context, b Backends, walletID string) (string, error) {
	var token dbToken
	var accessToken *external.AccessToken
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		err := tx.GetContext(ctx, &token, "SELECT * FROM astra_access_tokens WHERE wallet_id=$1 FOR UPDATE", walletID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w %s", astra.ErrInternal, err)
		}

		if !token.ExpiresAt.IsZero() && token.ExpiresAt.Before(time.Now()) {
			// Assume it's already refreshed, return
			accessToken = &external.AccessToken{
				AccessToken:  token.Token,
				RefreshToken: token.RefreshToken,
				TokenType:    "Bearer",
			}
			return nil
		}

		if token.RefreshExpiresAt.Before(time.Now()) {
			// Create new Token
			var intentID string
			err = tx.GetContext(ctx, &intentID, "SELECT intent_id FROM astra_user_intents WHERE wallet_id=$1", walletID)
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w intent not found for wallet ID", astra.ErrNotFound)
			}
			if err != nil {
				return fmt.Errorf("%w %s", astra.ErrInternal, err)
			}

			accessToken, err = b.External().CreateAccessToken(ctx, intentID, walletID)
			if err != nil {
				return fmt.Errorf("%w %s", astra.ErrInternal, err)
			}
			_, err = tx.ExecContext(ctx, "INSERT INTO astra_access_tokens (wallet_id, token, expires_at, refresh_token, refresh_expires_at) VALUES ($1, $2, $3, $4, $5)",
				walletID, accessToken.AccessToken, time.Now().Add(time.Minute*110), accessToken.RefreshToken, time.Now().Add(time.Hour*24*9))
		} else {
			// Use the refresh token
			accessToken, err = b.External().RefreshAccessToken(ctx, token.RefreshToken)
			if err != nil {
				return fmt.Errorf("%w %s", astra.ErrInternal, err)
			}
			_, err = tx.ExecContext(ctx, "UPDATE astra_access_tokens  SET token=$1, expires_at=$2, refresh_token=$3, refresh_expires_at=$4 WHERE id=$5",
				accessToken.AccessToken, time.Now().Add(time.Minute*110), accessToken.RefreshToken, time.Now().Add(time.Hour*24*9), token.ID)
		}
		if err != nil {
			return fmt.Errorf("%w %s", astra.ErrInternal, err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return accessToken.AccessToken, nil

}

func GetToken(ctx context.Context, b Backends, walletID string) (string, error) {
	var token dbToken
	err := b.DB().GetContext(ctx, &token, "SELECT * FROM astra_access_tokens WHERE wallet_id=$1", walletID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	if token.RefreshExpiresAt.Before(time.Now()) {
		return CreateOrRefreshToken(ctx, b, walletID)
	}

	return token.Token, nil
}
