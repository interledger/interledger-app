package unit

import (
	context "context"
	"errors"

	"gitlab.com/fynbos/backend/accounts"
	"go.temporal.io/sdk/temporal"
)

// file exists so implementation does not get lost when merged with https://gitlab.com/fynbos/fynbos/-/merge_requests/145

func (a *Activity) UnitCreateDepositAccount(
	ctx context.Context,
	customerID string,
) (*DepositAccount, error) {
	ret, err := a.unitService.CreateDepositAccount(ctx, customerID)
	if err != nil {
		if errors.Is(err, accounts.ErrNotFound) || errors.Is(err, ErrNotFound) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrInternal", err)
		}
		if errors.Is(err, ErrInvalidArgument) {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), "ErrInvalidArgument", err)
		}
		return nil, err
	}

	return ret, nil
}
