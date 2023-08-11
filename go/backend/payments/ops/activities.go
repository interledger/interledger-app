package ops

import (
	"context"
)

type Activity struct{}

func NewActivity(b Backends) *Activity {
	return &Activity{}
}

func (a *Activity) SetPaymentStateComplete(ctx context.Context, id string) error {
	// set status to completed

	// send payment sent email

	// send payment received email

	return nil
}

func (a *Activity) SetPaymentStateProcessing(ctx context.Context, id string) error {
	// set status to processing
	return nil
}

func (a *Activity) SetPaymentStateFailed(ctx context.Context, id string) error {
	// set status to failed

	// send payment failed email
	return nil
}
