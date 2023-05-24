package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

	walletID := uuid.NewString()

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
			_, err = tx.ExecContext(ctx, "INSERT INTO wallet_addresses(wallet_id, address) VALUES ($1, $2)", walletID, wa)
			if err != nil {
				return fmt.Errorf("%w %s", wallets.ErrInternal, err)
			}
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

	return Get(ctx, b, walletID)
}

func AddAddresses(ctx context.Context, b Backends, id string, args []wallets.Address) (*wallets.Wallet, error) {
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		for _, wa := range args {
			_, err := tx.ExecContext(ctx, "INSERT INTO wallet_addresses(wallet_id, address) VALUES ($1, $2)", id, wa)
			if err != nil {
				return fmt.Errorf("%w %s", wallets.ErrInternal, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return Get(ctx, b, id)
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

	users, err := b.Users().ListUsers(ctx, id)
	if err != nil {
		log.Error("error getting users", zap.Error(err))
		return &wallet, nil
	}
	for _, u := range users {
		b.Analytics().GroupUserWallet(id, u.ID)
	}

	return &wallet, nil
}

func SetWalletName(ctx context.Context, b Backends, id, name string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE wallets set name = $1 where id = $2", name, id)
	if err != nil {
		return err
	}

	return nil
}
