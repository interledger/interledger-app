package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"github.com/google/uuid"
)

func Create(ctx context.Context, b Backends, args *linkedaccounts.CreateArgs) (*linkedaccounts.LinkedAccount, error) {
	// TODO: refactor errors
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInvalidArgument, err.Error())
	}

	// TODO: add back ACL

	linkedAccountID := args.ID
	if linkedAccountID == "" {
		linkedAccountID = uuid.NewString()
	}
	var linkedAccount linkedaccounts.LinkedAccount
	err = b.DB().GetContext(
		ctx,
		&linkedAccount,
		`
			INSERT INTO linked_accounts (
				id, wallet_id, name, mask, provider, provider_id, type
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, wallet_id, name, mask, provider, provider_id, type, created_at, updated_at;
		`,
		linkedAccountID,
		args.WalletID,
		args.Name,
		args.Mask,
		args.Provider,
		args.ProviderID,
		args.Type,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	return &linkedAccount, nil
}

func Get(ctx context.Context, b Backends, id string) (*linkedaccounts.LinkedAccount, error) {
	if id == "" {
		return nil, fmt.Errorf("%w ID is required.", linkedaccounts.ErrInvalidArgument)
	}

	var linkedAccount linkedaccounts.LinkedAccount
	err := b.DB().GetContext(
		ctx,
		&linkedAccount,
		"SELECT id, wallet_id, name, mask, provider, provider_id, type, created_at, updated_at FROM linked_accounts where id=$1 LIMIT 1;",
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, linkedaccounts.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	return &linkedAccount, nil
}

func GetByProviderID(ctx context.Context, b Backends, args linkedaccounts.GetByProviderIDArgs) (*linkedaccounts.LinkedAccount, error) {
	var linkedAccount linkedaccounts.LinkedAccount
	err := b.DB().GetContext(
		ctx,
		&linkedAccount,
		`
			SELECT id, wallet_id, name, mask, provider, provider_id, type, created_at, updated_at FROM linked_accounts 
			WHERE provider=$1 AND provider_id=$2 AND type=$3 AND wallet_id=$4;
		`,
		args.Provider,
		args.ProviderID,
		args.Type,
		args.WalletID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, linkedaccounts.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	return &linkedAccount, nil
}

func ListByWalletId(ctx context.Context, b Backends, walletId string) ([]linkedaccounts.LinkedAccount, error) {

	var linkedAccounts []linkedaccounts.LinkedAccount
	err := b.DB().SelectContext(
		ctx,
		&linkedAccounts,
		"SELECT id, wallet_id, name, mask, provider, provider_id, type, created_at, updated_at FROM linked_accounts WHERE wallet_id=$1;",
		walletId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, linkedaccounts.ErrNotFound
		}

		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	return linkedAccounts, nil
}
