package ops

import (
	"context"

	"gitlab.com/fynbos/backend/payments"
)

type Activity struct{}

func NewActivity(b Backends) *Activity {
	return &Activity{}
}

func (a *Activity) SendPaymentSentEmail(ctx context.Context, paymentID string) error {
	return nil
}

func (a *Activity) SendPaymentReceivedEmail(ctx context.Context, paymentID string) error {
	return nil
}

func (a *Activity) SendPaymentFailedEmail(ctx context.Context, paymentID string) error {
	return nil
}

func (a *Activity) SetPaymentState(ctx context.Context, id string, state payments.State) error {
	return nil
}
