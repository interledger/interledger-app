// Package limits
// Limits is a hybrid service and has logical dependencies on transactions and authorisation and accesses their tables directly.
// For future scaling purposes, should the limits service become it's own gRPC service it will need access to the authorisation and transactions DB tables.
// This design also allows for decoupling limits from authorisation and transactions DB tables as the limits Client interface will not need to change,
// the authorisation and transactions DB access will just need to be moved to gRPC endpoints on the authorisation and transactions services respectively to keep separation.

package limits

import (
	"context"

	"gitlab.com/fynbos/backend/currency"
)

type Client interface {
	ExceedsKYCLimits(ctx context.Context, walletID string, amount currency.Amount) (bool, LimitType, error)
}
