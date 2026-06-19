package client

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/interledger/interledger-app/go/backend/payments"
	"github.com/interledger/interledger-app/go/backend/payments/ops"
)

var _ payments.Client = client{}

type client struct {
	b ops.Backends
}

func New(b ops.Backends) payments.Client {
	return &client{
		b: b,
	}
}

func (c client) Lookup(ctx context.Context, id string) (*payments.Payment, error) {
	return ops.Lookup(ctx, c.b, id)
}

// This will retry up to 3 times with exponential back-off as there may be PublicID collisions.
func (c client) Create(ctx context.Context, args payments.CreateArgs) (*payments.Payment, error) {
	operation := func() (*payments.Payment, error) {
		return ops.Create(ctx, c.b, args)
	}
	wrappedOperation := retryWithBackoff(operation)

	return wrappedOperation()
}

func (c client) Update(ctx context.Context, args payments.UpdateArgs) (*payments.Payment, error) {
	return ops.Update(ctx, c.b, args)
}

func (c client) Confirm(ctx context.Context, id string) (*payments.Payment, []payments.RequiredActionType, error) {
	return ops.Confirm(ctx, c.b, id)
}

func (c client) UpdateReceiver(ctx context.Context, id string, identity payments.Identity) error {
	return ops.UpdateReceiver(ctx, c.b, id, identity)
}

func (c client) SignalIdentityCreated(ctx context.Context, identifier string) error {
	return ops.SignalIdentityCreated(ctx, c.b, identifier)
}

func (c client) SignalAccountLinked(ctx context.Context, walletID string) error {
	return ops.SignalAccountLinked(ctx, c.b, walletID)
}

func (c client) AdminListAwaitingSignal(ctx context.Context) ([]payments.Payment, error) {
	return ops.ListAwaitingSignal(ctx, c.b)
}

func (c client) SignalExternalPayoutComplete(ctx context.Context, id string, success bool) error {
	return ops.SignalExternalPayoutComplete(ctx, c.b, id, success)
}

func (c client) SignalGatehubTransferComplete(ctx context.Context, externalTransactionID string) error {
	return ops.SignalGatehubTransferComplete(ctx, c.b, externalTransactionID)
}

var maxRetries = 3
var baseDelay = 1 * time.Millisecond

// RetryFunc is a function that can be retried
type retryFunc func() (*payments.Payment, error)

// RetryWithBackoff retries the given operation with exponential backoff
func retryWithBackoff(operation retryFunc) retryFunc {
	return func() (*payments.Payment, error) {
		var lastError error

		for i := 0; i < maxRetries; i++ {
			val, err := operation()
			if err == nil {
				return val, nil
			}

			secRetry := math.Pow(2, float64(i))
			fmt.Printf("Retrying operation in %f seconds\n", secRetry)
			delay := time.Duration(secRetry) * baseDelay
			time.Sleep(delay)
			lastError = err
		}

		return nil, lastError
	}
}
