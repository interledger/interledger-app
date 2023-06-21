package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/tabapay"

	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"github.com/google/uuid"
)

const (
	allFields = "id, wallet_id, name, nickname, mask, provider, provider_id, type, can_send, can_receive, state, created_at, updated_at"

	// If you update this, then remember to update the create and createBatch functions.
	insertFields  = "id, wallet_id, name, nickname, mask, provider, provider_id, type, can_send, can_receive, state"
	insertColumns = 11
)

func Create(ctx context.Context, b Backends, args *linkedaccounts.CreateArgs) (*linkedaccounts.LinkedAccount, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInvalidArgument, err.Error())
	}

	linkedAccountID := args.ID
	if linkedAccountID == "" {
		linkedAccountID = uuid.NewString()
	}

	state := args.State
	if state == "" {
		state = linkedaccounts.Verified
	}

	var linkedAccount linkedaccounts.LinkedAccount
	err = b.DB().GetContext(
		ctx,
		&linkedAccount,
		fmt.Sprintf("INSERT INTO linked_accounts (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING %s;", insertFields, allFields),
		linkedAccountID,
		args.WalletID,
		args.Name,
		args.Nickname,
		args.Mask,
		args.Provider,
		args.ProviderID,
		args.Type,
		args.CanSend,
		args.CanReceive,
		state,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	err = b.Notify().NotifyWallet(ctx, args.WalletID, notify.NotificationTypeLinkedAccount)
	if err != nil {
		log.Error("notify failed for linked account", zap.String("walletId", args.WalletID), zap.Error(err))
	}

	return &linkedAccount, nil
}

func CreateBatch(ctx context.Context, b Backends, args []linkedaccounts.CreateArgs) ([]linkedaccounts.LinkedAccount, error) {
	if len(args) == 0 {
		return nil, nil
	}

	for _, arg := range args {
		err := b.Validator().Struct(arg)
		if err != nil {
			return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInvalidArgument, err.Error())
		}
	}

	var placeholders []string
	var values []interface{}
	for i, arg := range args {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*insertColumns+1, i*insertColumns+2, i*insertColumns+3, i*insertColumns+4, i*insertColumns+5, i*insertColumns+6, i*insertColumns+7, i*insertColumns+8, i*insertColumns+9, i*insertColumns+10, i*insertColumns+11))

		linkedAccountID := arg.ID
		if linkedAccountID == "" {
			linkedAccountID = uuid.NewString()
		}
		state := arg.State
		if state == "" {
			state = linkedaccounts.Verified
		}

		values = append(values, linkedAccountID, arg.WalletID, arg.Name, arg.Nickname, arg.Mask, arg.Provider, arg.ProviderID, arg.Type, arg.CanSend, arg.CanReceive, state)
	}

	var linkedAccounts []linkedaccounts.LinkedAccount
	err := b.DB().SelectContext(
		ctx,
		&linkedAccounts,
		fmt.Sprintf("INSERT INTO linked_accounts (%s) VALUES %s RETURNING %s;", insertFields, strings.Join(placeholders, ","), allFields),
		values...,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	notifiedWallets := map[string]bool{}
	for _, la := range linkedAccounts {
		notified := notifiedWallets[la.WalletID]
		if notified {
			continue
		}

		err = b.Notify().NotifyWallet(ctx, la.WalletID, notify.NotificationTypeLinkedAccount)
		if err != nil {
			log.Error("notify failed for linked account", zap.String("walletId", la.WalletID), zap.Error(err))
		}

		notifiedWallets[la.WalletID] = true
	}

	return linkedAccounts, nil
}

func Get(ctx context.Context, b Backends, id string) (*linkedaccounts.LinkedAccount, error) {
	if id == "" {
		return nil, fmt.Errorf("%w ID is required.", linkedaccounts.ErrInvalidArgument)
	}

	var linkedAccount linkedaccounts.LinkedAccount
	err := b.DB().GetContext(
		ctx,
		&linkedAccount,
		fmt.Sprintf("SELECT %s FROM linked_accounts where id=$1 and deleted_at IS NULL LIMIT 1;", allFields),
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

func Delete(ctx context.Context, b Backends, id string) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE linked_accounts SET deleted_at=now() WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	return nil
}

func MarkNotDeleted(ctx context.Context, b Backends, id string) (*linkedaccounts.LinkedAccount, error) {
	result, err := b.DB().ExecContext(ctx, "UPDATE linked_accounts SET deleted_at=NULL WHERE id=$1", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}

	if rows < 1 {
		return nil, linkedaccounts.ErrNotFound
	}

	return Get(ctx, b, id)
}

func GetByProviderID(ctx context.Context, b Backends, args linkedaccounts.GetByProviderIDArgs) (*linkedaccounts.LinkedAccount, error) {
	var linkedAccount linkedaccounts.LinkedAccount
	err := b.DB().GetContext(
		ctx,
		&linkedAccount,
		fmt.Sprintf("SELECT %s FROM linked_accounts WHERE provider=$1 AND provider_id=$2;", allFields),
		args.Provider,
		args.ProviderID,
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
		fmt.Sprintf("SELECT %s FROM linked_accounts WHERE deleted_at IS NULL AND wallet_id=$1;", allFields),
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

func ListMXBankAccounts(ctx context.Context, b Backends) ([]linkedaccounts.LinkedAccount, error) {
	var linkedAccounts []linkedaccounts.LinkedAccount
	err := b.DB().SelectContext(
		ctx,
		&linkedAccounts,
		fmt.Sprintf("SELECT %s FROM linked_accounts WHERE deleted_at IS NULL AND provider=$1 AND type=$2;", allFields),
		mx.ProviderName, mx.TypeBankAccount,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err)
	}

	return linkedAccounts, nil
}

func SetNickname(ctx context.Context, b Backends, id, nickname string) error {
	result, err := b.DB().ExecContext(ctx, `UPDATE linked_accounts set nickname = $1 where id = $2`, nickname, id)
	if err != nil {
		return fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}
	ra, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w %s", linkedaccounts.ErrInternal, err.Error())
	}
	if ra == 0 {
		return linkedaccounts.ErrNotFound
	}

	return nil
}

func Requires3DS(ctx context.Context, b Backends, id string) (bool, error) {
	la, err := Get(ctx, b, id)
	if err != nil {
		return false, err
	}

	return la.Provider == tabapay.ProviderName && la.Type == tabapay.TypeCard, nil
}
