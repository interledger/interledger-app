package contacts

import (
	"context"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/paymentpointers"
)

type Client interface {
	Create(ctx context.Context, args CreateContactArgs) (*Contact, error)
	List(ctx context.Context, walletID string, page db.Pagination, orderBy string) ([]Contact, error)
	Get(ctx context.Context, walletID string, pp paymentpointers.PaymentPointer) (*Contact, error)
	SetLastPaidAtNow(ctx context.Context, walletID string, pp paymentpointers.PaymentPointer) error
}
