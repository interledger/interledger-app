package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/db"

	"gitlab.com/fynbos/backend/user"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/log"
)

func Create(ctx context.Context, b Backends, args wallets.CreateArgs) (*wallets.Wallet, error) {
	err := b.Validator().StructCtx(ctx, args)
	if err != nil {
		return nil, err
	}

	userID := args.UserID
	name := args.Name
	if name == "" {
		name = "default"
	}

	walletID := args.ID
	if walletID == "" {
		walletID = uuid.NewString()
	}

	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO wallets (id, name) VALUES ($1, $2)", walletID, name)
		if err != nil {
			return fmt.Errorf("%w %s", wallets.ErrInternal, err)
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO user_wallets (user_id, wallet_id) VALUES ($1, $2)", userID, walletID)
		if err != nil {
			return fmt.Errorf("%w %s", wallets.ErrInternal, err)
		}

		for _, wa := range args.Addresses {
			_, err = tx.ExecContext(ctx, "INSERT INTO wallet_addresses(wallet_id, url) VALUES ($1, $2)", walletID, wa)
			if err != nil {
				return fmt.Errorf("%w %s", wallets.ErrInternal, err)
			}
		}

		// Check that only 1 wallet exists for the user with that name.
		var exists wallets.Wallet
		err = tx.GetContext(ctx, &exists, "SELECT w.id, w.name FROM wallets w INNER JOIN user_wallets uw ON w.id = uw.wallet_id WHERE user_id=$1 and w.name=$2 AND w.id <> $3", userID, name, walletID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w %s", user.ErrInternal, err)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w duplicate wallet name (%s) for user (%s)", wallets.ErrDuplicateWallet, name, userID)
		}

		return nil
	})

	b.Analytics().TrackWalletCreated(walletID, userID)
	for range args.Addresses {
		b.Analytics().TrackWalletPaymentPointerCreated(walletID)
	}

	if err != nil {
		return nil, err
	}
	err = b.Keys().ProvisionPrivateKey(ctx, walletID)
	if err != nil {
		log.Error("could not provision private key", zap.Error(err))
		return nil, err
	}

	return Get(ctx, b, walletID)
}

func AddAddress(ctx context.Context, b Backends, id, url string) (*wallets.Wallet, error) {
	// Check if it already exists
	w, err := GetFromAddress(ctx, b, url)
	if err != nil && !errors.Is(err, wallets.ErrNoWalletFound) {
		return nil, err
	}

	if w != nil && w.ID == id {
		// The address already belongs to this wallet, nothing to see here.
		return w, nil
	}

	address, err := wallets.ParseAddress(url)
	if err != nil {
		return nil, err
	}
	_, err = b.DB().ExecContext(ctx, "INSERT INTO wallet_addresses(wallet_id, url) VALUES ($1, $2)", id, address)
	if db.IsErrorCode(err, db.UniqueViolationError) {
		return nil, fmt.Errorf("%w %s", wallets.ErrAddressExists, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", wallets.ErrInternal, err)
	}

	b.Analytics().TrackWalletPaymentPointerCreated(id)

	return Get(ctx, b, id)
}

func GetFromAddress(ctx context.Context, b Backends, url string) (*wallets.Wallet, error) {
	address, err := wallets.ParseAddress(url)
	if err != nil {
		return nil, err
	}

	var wid string
	err = b.DB().GetContext(ctx, &wid, "SELECT wallet_id FROM wallet_addresses WHERE lower(url)=lower($1)", address)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w address(%s) not found", wallets.ErrNoWalletFound, address.String())
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", wallets.ErrInternal, err)
	}

	return Get(ctx, b, wid)
}

func Get(ctx context.Context, b Backends, id string) (*wallets.Wallet, error) {
	var wallet wallets.Wallet
	err := b.DB().GetContext(ctx, &wallet,
		"SELECT id, name FROM wallets WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, wallets.ErrNoWalletFound
	}
	if err != nil {
		return nil, err
	}

	var wa []wallets.Address
	err = b.DB().SelectContext(ctx, &wa, "SELECT url FROM wallet_addresses WHERE wallet_id=$1", id)
	if err != nil {
		return nil, err
	}

	wallet.Addresses = wa

	// disable temp
	//users, err := b.Users().ListUsers(ctx, id)
	//if err != nil {
	//	log.Error("error getting users", zap.Error(err))
	//	return &wallet, nil
	//}
	//for _, u := range users {
	//	b.Analytics().GroupUserWallet(id, u.ID)
	//}

	return &wallet, nil
}

func List(ctx context.Context, b Backends, userID string) ([]wallets.Wallet, error) {
	var wl []wallets.Wallet
	err := b.DB().SelectContext(ctx, &wl, "SELECT w.id, w.name FROM wallets w INNER JOIN user_wallets uw ON w.id = uw.wallet_id WHERE user_id=$1", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	for i, w := range wl {
		var wa []wallets.Address
		err = b.DB().SelectContext(ctx, &wa, "SELECT url FROM wallet_addresses WHERE wallet_id=$1", w.ID)
		if err != nil {
			return nil, err
		}

		wl[i].Addresses = wa
	}

	return wl, nil
}

func ListAll(ctx context.Context, b Backends, _ db.Pagination) ([]wallets.Wallet, error) {
	var wl []wallets.Wallet
	err := b.DB().SelectContext(ctx, &wl, "SELECT id, name FROM wallets ")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	for i, w := range wl {
		var wa []wallets.Address
		err = b.DB().SelectContext(ctx, &wa, "SELECT url FROM wallet_addresses WHERE wallet_id=$1", w.ID)
		if err != nil {
			return nil, err
		}

		wl[i].Addresses = wa
	}

	return wl, nil
}

func SetWalletName(ctx context.Context, b Backends, id, name string) (*wallets.Wallet, error) {
	_, err := b.DB().ExecContext(ctx, "UPDATE wallets set name = $1 where id = $2", name, id)
	if err != nil {
		return nil, err
	}

	return Get(ctx, b, id)
}

func WalletForContext(ctx context.Context) (*wallets.Wallet, error) {
	w, ok := ctx.Value(wallets.CtxKey).(*wallets.Wallet)
	if !ok || w == nil {
		return nil, wallets.ErrNoWalletFound
	}
	return w, nil
}
