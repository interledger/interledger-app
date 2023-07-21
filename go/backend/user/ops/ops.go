package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/db"

	client "github.com/ory/kratos-client-go"

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
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return nil, nil
		}
		return nil, err
	}

	u := convertTraits(session.Identity.Id, session.Identity.Traits)
	return &u, nil
}

func GetUser(ctx context.Context, b Backends, userID string) (*user.User, error) {
	id, _, err := b.Kratos().V0alpha2Api.AdminGetIdentity(ctx, userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("%w %s", user.ErrInternal, err)
	}

	u := convertTraits(id.Id, id.Traits)
	return &u, nil
}

func convertTraits(userID string, traits interface{}) user.User {
	traitsMap := traits.(map[string]interface{})
	u := user.User{
		ID:          userID,
		Email:       traitsMap["email"].(string),
		PhoneNumber: traitsMap["phone"].(string),
	}
	// All trait values:  "email", "phone", "firstName", "lastName", "countryCode"
	return u
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

func CreateWallet(ctx context.Context, b Backends, args user.CreateWalletArgs) (*user.Wallet, error) {
	userID := args.UserID
	name := args.Name
	if name == "" {
		name = "default"
	}

	walletID := args.ID
	if walletID == "" {
		walletID = uuid.NewString()
	}
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", walletID, name)
		if err != nil {
			return fmt.Errorf("%w %s", user.ErrInternal, err)
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO user_wallets (user_id, wallet_id) VALUES ($1, $2)", userID, walletID)
		if err != nil {
			return fmt.Errorf("%w %s", user.ErrInternal, err)
		}

		// Check that only 1 wallet exists for the user with that name.
		var exists user.Wallet
		err = tx.GetContext(ctx, &exists, "SELECT w.id, w.name FROM wallets w INNER JOIN user_wallets uw ON w.id = uw.wallet_id WHERE user_id=$1 and w.name=$2 AND w.id <> $3", userID, name, walletID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w %s", user.ErrInternal, err)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w duplicate wallet name (%s) for user (%s)", user.ErrDuplicateWallet, name, userID)
		}

		return nil
	})

	b.Analytics().TrackWalletCreated(walletID, userID)

	if err != nil {
		return nil, err
	}
	err = b.Keys().ProvisionPrivateKey(ctx, walletID)
	if err != nil {
		log.Error("could not provision private key", zap.Error(err))
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

func GetWallet(ctx context.Context, b Backends, walletID string) (*user.Wallet, error) {

	var wallet user.Wallet
	err := b.DB().GetContext(ctx, &wallet,
		"SELECT id, name FROM wallets WHERE id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrNoWalletFound
	}
	if err != nil {
		return nil, err
	}

	users, err := ListUsers(ctx, b, walletID)
	if err != nil {
		log.Error("error getting users", zap.Error(err))
		return &wallet, nil
	}
	for _, u := range users {
		b.Analytics().GroupUserWallet(walletID, u.ID)
	}

	return &wallet, nil
}

func ListUsers(ctx context.Context, b Backends, walletID string) ([]user.User, error) {
	ctx, cancel := context.WithTimeout(ctx, kratosTimeout)
	defer cancel()

	var userIDs []string
	err := b.DB().SelectContext(ctx, &userIDs, "SELECT user_id FROM user_wallets WHERE wallet_id=$1", walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, user.ErrNoWalletFound
	}
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mx sync.Mutex

	// Required for kratos to use admin server
	ctx = context.WithValue(ctx, client.ContextServerIndex, 1)

	var resp []user.User
	var anyErr error

	// Required to check so test can pass due to kratos not being mocked.
	if b.Kratos() != nil {
		for _, userID := range userIDs {
			wg.Add(1)
			go func(uID string) {
				defer wg.Done()
				id, _, err := b.Kratos().V0alpha2Api.AdminGetIdentity(ctx, uID).Execute()
				if err != nil {
					anyErr = err
					return
				}

				// lock
				mx.Lock()
				defer mx.Unlock()
				resp = append(resp, convertTraits(id.Id, id.Traits))
			}(userID)
		}
	}

	wg.Wait()
	if anyErr != nil {
		return nil, anyErr
	}

	return resp, nil
}

func ListAllWallets(ctx context.Context, b Backends, _ db.Pagination) ([]user.Wallet, error) {
	var wallets []user.Wallet

	// TODO pagination
	err := b.DB().SelectContext(ctx, &wallets, "SELECT id, name FROM wallets ORDER BY created_at DESC ")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return wallets, nil
}

func SetWalletName(ctx context.Context, b Backends, id, name string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE wallets set name = $1 where id = $2", name, id)
	if err != nil {
		return err
	}

	return nil
}
