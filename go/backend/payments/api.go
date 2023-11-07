package payments

import "context"

type Client interface {
	Lookup(ctx context.Context, id string) (*Payment, error)
	Create(ctx context.Context, args CreateArgs) (*Payment, error)
	Update(ctx context.Context, args UpdateArgs) (*Payment, error)
	Confirm(ctx context.Context, id string) (*Payment, []RequiredActionType, error)

	UpdateReceiver(ctx context.Context, id string, identity Identity) error
	SignalIdentityCreated(ctx context.Context, identifier string) error
	SignalAccountLinked(ctx context.Context, walletID string) error
	SignalExternalPayoutComplete(ctx context.Context, id string, success bool) error
	CreatePaymentLink(ctx context.Context, id string) (*PaymentLink, error)
	ConsumePaymentLink(ctx context.Context, args ConsumePaymentLinkArgs) (*PaymentLink, error)
	GetPaymentLink(ctx context.Context, id string) (*PaymentLink, error)
	GetPaymentLinkByToken(ctx context.Context, token string) (*PaymentLink, error)

	AdminListAwaitingSignal(ctx context.Context) ([]Payment, error)
}
