package payments

import "context"

type Client interface {
	Get(ctx context.Context, paymentID, walletID string) (*Payment, error)

	Lookup(ctx context.Context, id string) (*Payment, error)
	Create(ctx context.Context, args CreateArgs) (*Payment, error)
	Update(ctx context.Context, args UpdateArgs) (*Payment, error)
	Confirm(ctx context.Context, id string) (*Payment, []RequiredActionType, error)

	UpdateReceiver(ctx context.Context, id string, identity Identity) error
	SignalIdentityCreated(ctx context.Context, identifier string) error
	SignalAccountLinked(ctx context.Context, walletID string) error
	SignalExternalPayoutComplete(ctx context.Context, id string, success bool) error
	SignalGatehubTransferComplete(ctx context.Context, externalTransactionID string) error

	AdminListAwaitingSignal(ctx context.Context) ([]Payment, error)
}
