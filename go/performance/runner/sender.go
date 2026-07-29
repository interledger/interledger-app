package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/performance/client"
	"github.com/interledger/interledger-app/go/performance/config"
	"github.com/interledger/interledger-app/go/performance/metrics"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"

	"golang.org/x/time/rate"
)

// senderLoop issues payments from one wallet until it runs out of reasons to
// continue: an empty balance, a reached count, the run's deadline, or a fatal
// failure.
type senderLoop struct {
	cfg     *config.Config
	sender  *sender
	rec     *metrics.Recorder
	coll    *metrics.Collectors
	limiter *rate.Limiter
	watch   *watcher
	// measuring reports whether the warmup window has passed, so warmup samples
	// can be dropped from the percentiles without dropping the payments.
	measuring func() bool
}

// run drives the sender until it stops, returning why it stopped and, when the
// failure was fatal, an error that aborts the run.
func (l *senderLoop) run(ctx context.Context) (reason string, err error) {
	maxPayments := l.sender.maxPayments(l.cfg)
	sent := 0

	for {
		if ctx.Err() != nil {
			return "run ended", nil
		}

		switch l.cfg.Run.Stop {
		case config.StopCount:
			if sent >= l.cfg.Run.CountPerSender {
				return fmt.Sprintf("sent %d payments", sent), nil
			}
		case config.StopDrain:
			// Balance/amount is the theoretical ceiling. The authoritative stop is
			// the backend's insufficient-funds error below — fees mean a wallet can
			// empty before this cap is reached.
			if maxPayments > 0 && sent >= maxPayments {
				return fmt.Sprintf("balance exhausted after %d payments", sent), nil
			}
		case config.StopDuration:
			// Bounded by the run context alone.
		}

		// The rate limiter is shared by every sender, so it caps the run's total
		// arrival rate rather than each wallet's.
		if l.limiter != nil {
			if err := l.limiter.Wait(ctx); err != nil {
				return "run ended", nil
			}
		}

		// Backpressure: wait for an in-flight slot before adding more work.
		if !l.watch.acquire(ctx) {
			return "run ended", nil
		}

		outcome, failure := l.sendOne(ctx)
		sent++

		switch outcome {
		case sendTracked:
			// Slot handed to the watcher; it will release it.
		case sendRejected:
			l.watch.release()
			if failure != nil && failure.Class == client.ClassExhausted {
				return failure.Message, nil
			}
			if failure != nil && failure.Class == client.ClassFatal {
				return failure.Error(), fmt.Errorf("sender %s: %w", l.sender.cfg.Label, failure)
			}
		}
	}
}

type sendOutcome int

const (
	// sendTracked means the payment was confirmed and handed to the watcher.
	sendTracked sendOutcome = iota
	// sendRejected means the backend refused the payment at some stage.
	sendRejected
)

// sendOne performs one payment: create, optionally update, then confirm.
func (l *senderLoop) sendOne(ctx context.Context) (sendOutcome, *client.Failure) {
	l.rec.AddAttempt(l.sender.cfg.Label)

	receiver := l.sender.nextReceiver(l.cfg.Run.Pairing)
	amount := &pb.Amount{
		Amount:     l.cfg.Run.Amount,
		Asset:      l.sender.asset,
		AssetScale: l.sender.assetScale,
	}

	created, failure := l.create(ctx, receiver, amount)
	if failure != nil {
		return sendRejected, failure
	}

	if l.cfg.Run.IncludeUpdateStep {
		if failure := l.update(ctx, created.GetId(), amount); failure != nil {
			return sendRejected, failure
		}
	}

	confirmedAt, failure := l.confirm(ctx, created.GetId())
	if failure != nil {
		return sendRejected, failure
	}

	l.rec.AddConfirmed(l.sender.cfg.Label)

	l.watch.track(ctx, inFlight{
		paymentID:   created.GetId(),
		sender:      l.sender,
		confirmedAt: confirmedAt,
	})

	return sendTracked, nil
}

func (l *senderLoop) create(ctx context.Context, receiver config.Receiver, amount *pb.Amount) (*pb.Payment, *client.Failure) {
	req := &pb.CreatePaymentRequest{
		SenderAmount:         amount,
		ReceiverIdentity:     receiver.WalletAddress,
		ReceiverIdentityType: receiverIdentityTypeWalletURL,
	}
	if l.sender.linkedAccount != "" {
		req.SenderAccount = &l.sender.linkedAccount
	}
	if receiver.LinkedAccount != "" {
		req.ReceiverAccount = &receiver.LinkedAccount
	}
	if l.cfg.Run.Note != "" {
		req.Note = &l.cfg.Run.Note
	}

	start := time.Now()
	payment, err := l.sender.wallet.CreatePayment(ctx, req)
	l.observe(metrics.StageCreate, time.Since(start))

	if err != nil {
		return nil, l.recordFailure(metrics.StageCreate, "create", err)
	}
	return payment, nil
}

// update mirrors protea's send flow, which creates a payment and then amends it
// before confirming. Off by default: it doubles the write path for no extra
// transaction, so it only belongs in a run that is deliberately measuring what
// the UI does.
func (l *senderLoop) update(ctx context.Context, paymentID string, amount *pb.Amount) *client.Failure {
	req := &pb.UpdatePaymentRequest{
		Id:           paymentID,
		SenderAmount: amount,
	}
	if l.sender.linkedAccount != "" {
		req.SenderAccount = &l.sender.linkedAccount
	}

	start := time.Now()
	_, err := l.sender.wallet.UpdatePayment(ctx, req)
	l.observe(metrics.StageUpdate, time.Since(start))

	if err != nil {
		return l.recordFailure(metrics.StageUpdate, "update", err)
	}
	return nil
}

// confirm authorises the payment and returns the instant the call completed,
// which is where settlement latency starts counting.
func (l *senderLoop) confirm(ctx context.Context, paymentID string) (time.Time, *client.Failure) {
	start := time.Now()
	_, err := l.sender.wallet.ConfirmPayment(ctx, paymentID)
	done := time.Now()
	l.observe(metrics.StageConfirm, done.Sub(start))

	if err != nil {
		return time.Time{}, l.recordFailure(metrics.StageConfirm, "confirm", err)
	}
	return done, nil
}

func (l *senderLoop) recordFailure(stage metrics.Stage, stageName string, err error) *client.Failure {
	f := client.Classify(stageName, err)
	l.rec.CountError(stage, f.Class.String(), f.Key())
	l.coll.CountError(stage, f.Class.String(), f.Key())
	l.rec.CountOutcome(metrics.OutcomeRejected)
	l.coll.CountPayment(metrics.OutcomeRejected)
	return f
}

// observe records a latency sample, skipping the recorder during warmup so
// connection setup does not pollute the percentiles. The Prometheus series keep
// every sample, since the warmup is visible there as a time range.
func (l *senderLoop) observe(stage metrics.Stage, d time.Duration) {
	l.coll.ObserveRPC(stage, d)
	if l.measuring() {
		l.rec.Observe(stage, d)
	}
}
