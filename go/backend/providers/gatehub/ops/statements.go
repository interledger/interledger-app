package ops

import (
	"context"
	"errors"
	"fmt"
	"io"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
)

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
