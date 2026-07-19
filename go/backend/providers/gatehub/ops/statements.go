package ops

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub/external"
)

func GetAccountStatement(ctx context.Context, b Backends, ec external.Client, walletID string, year, month int) (io.ReadCloser, error) {
	linkedAccounts, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	var linkedAccount linkedaccounts.LinkedAccount
	for _, la := range linkedAccounts {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			linkedAccount = la
			break
		}
	}

	if linkedAccount.ProviderID == "" {
		return nil, fmt.Errorf("%w no gatehub linked account found for wallet", gatehub.ErrNotFound)
	}

	now := time.Now().UTC()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	accountCreatedAt := time.Date(linkedAccount.CreatedAt.UTC().Year(), linkedAccount.CreatedAt.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	requestedDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	if requestedDate.Before(accountCreatedAt) {
		return nil, fmt.Errorf("%w: requested date is before account creation", gatehub.ErrBadRequest)
	}
	if requestedDate.After(currentMonth) {
		return nil, fmt.Errorf("%w: requested date must not be in the future", gatehub.ErrBadRequest)
	}

	externalIDs, err := GetExternalIDs(ctx, b, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	result, err := ec.GetAccountStatement(ctx, externalIDs.UserID, linkedAccount.ProviderID, year, month)
	if errors.Is(err, external.ErrNotFound) {
		return nil, fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	return result, nil
}

func GetAccountConfirmation(ctx context.Context, b Backends, ec external.Client, walletID string) (io.ReadCloser, error) {
	linkedAccounts, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	var linkedAccount linkedaccounts.LinkedAccount
	for _, la := range linkedAccounts {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			linkedAccount = la
			break
		}
	}

	if linkedAccount.ProviderID == "" {
		return nil, fmt.Errorf("%w no gatehub linked account found for wallet", gatehub.ErrNotFound)
	}

	externalIDs, err := GetExternalIDs(ctx, b, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	result, err := ec.GetAccountConfirmation(ctx, externalIDs.UserID, linkedAccount.ProviderID)
	if errors.Is(err, external.ErrNotFound) {
		return nil, fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}
	return result, nil
}

func GetTransactionStatement(ctx context.Context, b Backends, ec external.Client, walletID string, txID string) (io.ReadCloser, error) {
	linkedAccounts, err := b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	var linkedAccount linkedaccounts.LinkedAccount
	for _, la := range linkedAccounts {
		if la.Provider == gatehub.ProviderName && la.Type == gatehub.AccTypeBalance {
			linkedAccount = la
			break
		}
	}

	if linkedAccount.ProviderID == "" {
		return nil, fmt.Errorf("%w no gatehub linked account found for wallet", gatehub.ErrNotFound)
	}

	externalIDs, err := GetExternalIDs(ctx, b, linkedAccount.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	transaction, err := b.Transactions().GetTransaction(ctx, walletID, txID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	if transaction == nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrNotFound, "transaction not found")
	}

	result, err := ec.GetTransferConfirmation(ctx, externalIDs.UserID, transaction.ForeignID)
	if errors.Is(err, external.ErrNotFound) {
		return nil, fmt.Errorf("%w %s", gatehub.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", gatehub.ErrInternal, err)
	}

	return result, nil
}
