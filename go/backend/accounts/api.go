package accounts

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Client interface {
	Init(ctx context.Context) error
	Create(ctx context.Context, args CreateAccountArgs) (*Account, error)
	GetByIdentityIDWithTrx(ctx context.Context, tx *sqlx.Tx, id string) (*Account, error)
	GetByIdentityID(ctx context.Context, id string) (*Account, error)
	Get(ctx context.Context, id string) (*Account, error)
	VerifyWithTx(ctx context.Context, tx *sqlx.Tx, args VerifyArgs) (*Account, error)
	CanMakeOutgoingPayment(acc Account, identityID string) bool
	CanMakeDeposit(acc Account, identityID string) bool
	CanCreateFundingSource(acc Account, identityID string) bool
	CanVerifyFundingSource(acc Account, identityID string) bool
}
