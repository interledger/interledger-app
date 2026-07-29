package runner

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/interledger/interledger-app/go/performance/client"
	"github.com/interledger/interledger-app/go/performance/config"
	"github.com/interledger/interledger-app/go/performance/metrics"
)

// Payment states, mirroring payments.State. The proto carries these as a bare
// int32, so the harness needs its own copy.
const (
	stateCreated    int32 = 1
	stateConfirmed  int32 = 2
	stateProcessing int32 = 3
	stateCompleted  int32 = 4
	stateFailed     int32 = 5
)

func isTerminal(state int32) bool {
	return state == stateCompleted || state == stateFailed
}

// inFlight is one payment being watched through to a terminal state.
type inFlight struct {
	paymentID string
	sender    *sender
	// confirmedAt is when ConfirmPayment returned. Settlement latency is measured
	// from here, because that is the moment the backend is free to move money.
	confirmedAt time.Time
}

// watcher polls confirmed payments until they settle.
//
// This is the piece a pure RPC load generator cannot give you. CreatePayment
// starts a Temporal workflow and returns immediately, so without this the harness
// would report how fast it can enqueue work and call it throughput. The watcher
// also owns the in-flight semaphore, which is what stops the harness from
// queueing unbounded work into a backend that has stopped keeping up.
type watcher struct {
	cfg      *config.Config
	rec      *metrics.Recorder
	coll     *metrics.Collectors
	queue    chan inFlight
	slots    chan struct{}
	inFlight atomic.Int64
	settled  atomic.Int64
	wg       sync.WaitGroup
	// measuring reports whether the warmup window has passed.
	measuring func() bool
}

func newWatcher(cfg *config.Config, rec *metrics.Recorder, coll *metrics.Collectors, measuring func() bool) *watcher {
	return &watcher{
		cfg:       cfg,
		rec:       rec,
		coll:      coll,
		queue:     make(chan inFlight, cfg.Run.MaxInFlight),
		slots:     make(chan struct{}, cfg.Run.MaxInFlight),
		measuring: measuring,
	}
}

// start launches the polling pool.
func (w *watcher) start(ctx context.Context) {
	if !w.cfg.Settlement.Track {
		return
	}
	for range w.cfg.Settlement.Workers {
		w.wg.Go(func() {
			w.poll(ctx)
		})
	}
}

// acquire blocks until an in-flight slot is free, applying backpressure to the
// senders. Returns false if the context is cancelled first.
//
// When settlement tracking is off there is nothing to release the slots, so the
// cap does not apply and senders run unthrottled by design.
func (w *watcher) acquire(ctx context.Context) bool {
	if !w.cfg.Settlement.Track {
		return true
	}
	select {
	case w.slots <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

// release returns a slot taken by acquire without a payment having been queued,
// e.g. when the create or confirm call failed.
func (w *watcher) release() {
	if !w.cfg.Settlement.Track {
		return
	}
	select {
	case <-w.slots:
	default:
	}
}

// track hands a confirmed payment to the polling pool. The caller must already
// hold a slot from acquire.
func (w *watcher) track(ctx context.Context, p inFlight) {
	if !w.cfg.Settlement.Track {
		return
	}
	w.inFlight.Add(1)
	w.coll.SetInFlight(int(w.inFlight.Load()))

	select {
	case w.queue <- p:
	case <-ctx.Done():
		w.finish(p, metrics.OutcomeTimedOut, 0)
	}
}

// close stops the pool once all queued payments have been drained.
func (w *watcher) close() {
	if !w.cfg.Settlement.Track {
		return
	}
	close(w.queue)
	w.wg.Wait()
}

// stats reports current progress for the periodic progress line.
func (w *watcher) stats() (inFlightCount, settledCount int) {
	return int(w.inFlight.Load()), int(w.settled.Load())
}

func (w *watcher) poll(ctx context.Context) {
	for p := range w.queue {
		w.pollOne(ctx, p)
	}
}

func (w *watcher) pollOne(ctx context.Context, p inFlight) {
	deadline := p.confirmedAt.Add(w.cfg.Settlement.Timeout)
	ticker := time.NewTicker(w.cfg.Settlement.PollInterval)
	defer ticker.Stop()

	for {
		// A cancelled run still needs its in-flight payments accounted for, or the
		// report would silently lose them.
		if ctx.Err() != nil {
			w.finish(p, metrics.OutcomeTimedOut, time.Since(p.confirmedAt))
			return
		}
		if time.Now().After(deadline) {
			w.finish(p, metrics.OutcomeTimedOut, time.Since(p.confirmedAt))
			return
		}

		start := time.Now()
		payment, err := p.sender.wallet.GetPayment(ctx, p.paymentID)
		w.observe(metrics.StagePoll, time.Since(start))

		if err != nil {
			f := client.Classify("poll", err)
			w.rec.CountError(metrics.StagePoll, f.Class.String(), f.Key())
			w.coll.CountError(metrics.StagePoll, f.Class.String(), f.Key())
			if f.Class == client.ClassFatal {
				w.finish(p, metrics.OutcomeFailed, time.Since(p.confirmedAt))
				return
			}
			// Transient: fall through and try again on the next tick.
		} else if isTerminal(payment.GetState()) {
			outcome := metrics.OutcomeCompleted
			if payment.GetState() == stateFailed {
				outcome = metrics.OutcomeFailed
			}
			w.finish(p, outcome, time.Since(p.confirmedAt))
			return
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			w.finish(p, metrics.OutcomeTimedOut, time.Since(p.confirmedAt))
			return
		}
	}
}

// finish records a terminal payment and frees its in-flight slot.
func (w *watcher) finish(p inFlight, outcome metrics.Outcome, elapsed time.Duration) {
	if elapsed > 0 {
		w.observe(metrics.StageSettle, elapsed)
	}

	w.rec.CountOutcome(outcome)
	w.coll.CountPayment(outcome)
	if outcome == metrics.OutcomeCompleted {
		w.rec.AddCompleted(p.sender.cfg.Label)
	}

	w.settled.Add(1)
	w.inFlight.Add(-1)
	w.coll.SetInFlight(int(w.inFlight.Load()))
	w.release()
}

// observe records a latency sample to both sinks, skipping the recorder during
// warmup so early samples stay out of the reported percentiles. Prometheus keeps
// every sample, since the warmup is visible there as a time range.
func (w *watcher) observe(stage metrics.Stage, d time.Duration) {
	if stage == metrics.StageSettle {
		w.coll.ObserveSettlement(d)
	} else {
		w.coll.ObserveRPC(stage, d)
	}

	if w.measuring() {
		w.rec.Observe(stage, d)
	}
}
