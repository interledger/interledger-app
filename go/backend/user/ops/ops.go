package ops

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"gitlab.com/fynbos/backend/user"
)

const (
	userCtxKey   = user.UserCtxKey("user")
	walletCtxKey = user.WalletCtxKey("wallet")

	kratosTimeout    = 500 * time.Millisecond
	kratosCookieName = "ory_kratos_session"
)

func UserForCookie(ctx context.Context, b Backends, cookie string) (*user.User, error) {
	if cookie == "" {
		return nil, user.ErrNoUserFound
	}

	ctx, cancel := context.WithTimeout(ctx, kratosTimeout)
	defer cancel()

	session, resp, err := b.Kratos().V0alpha2Api.ToSession(ctx).Cookie(kratosCookieName + "=" + cookie).Execute()
	if err != nil {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, nil
		}
		return nil, err
	}
	traits := session.Identity.Traits.(map[string]interface{})
	u := user.User{
		ID:    session.Identity.Id,
		Email: traits["email"].(string),
	}
	return &u, nil
}

func UserForContext(ctx context.Context) (*user.User, error) {
	u, ok := ctx.Value(userCtxKey).(*user.User)
	if !ok || u == nil {
		return nil, user.ErrNoUserFound
	}
	return u, nil
}

func WalletForContext(ctx context.Context) (*user.Wallet, error) {
	w, ok := ctx.Value(walletCtxKey).(*user.Wallet)
	if !ok || w == nil {
		return nil, user.ErrNoWalletFound
	}
	return w, nil
}

func CreateWallet(ctx context.Context, b Backends, userID, name string) (*user.Wallet, error) {
	if name == "" {
		name = "default"
	}

	walletID := uuid.NewString()
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {

		_, err := tx.ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", walletID, name)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO user_wallets (user_id, wallet_id) VALUES ($1, $2)", userID, walletID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &user.Wallet{
		ID:   walletID,
		Name: name,
	}, nil
}

func ListWallets(ctx context.Context, b Backends, userID string) ([]user.Wallet, error) {
	var wallets []user.Wallet
	err := b.DB().SelectContext(ctx, &wallets, "SELECT w.id, w.name FROM wallets w INNER JOIN user_wallets uw ON w.id = uw.wallet_id WHERE user_id=$1", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return wallets, nil
}

func GetWallet(ctx context.Context, b Backends, userID, walletID string) (*user.Wallet, error) {

	var wallet user.Wallet
	err := b.DB().GetContext(ctx, &wallet,
		"SELECT w.id, w.name FROM wallets w INNER JOIN user_wallets uw ON w.id = uw.wallet_id WHERE user_id=$1 and w.id=$2", userID, wallet)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}
