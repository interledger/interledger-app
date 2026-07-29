package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/interledger/interledger-app/go/performance/auth"
	"github.com/interledger/interledger-app/go/performance/client"
	"github.com/interledger/interledger-app/go/performance/config"
	"github.com/interledger/interledger-app/go/performance/metrics"

	"golang.org/x/time/rate"
)

// progressInterval is how often the run prints a one-line status while working.
const progressInterval = 5 * time.Second

// Runner executes one scenario.
type Runner struct {
	cfg  *config.Config
	out  io.Writer
	rec  *metrics.Recorder
	coll *metrics.Collectors
}

// New builds a Runner. Progress and the final report are written to out.
func New(cfg *config.Config, out io.Writer) *Runner {
	return &Runner{
		cfg:  cfg,
		out:  out,
		rec:  metrics.NewRecorder(),
		coll: metrics.NewCollectors(cfg.Metrics.JobLabel),
	}
}

// Collectors exposes the run's Prometheus metrics so the caller can serve them.
func (r *Runner) Collectors() *metrics.Collectors { return r.coll }

// Run executes the scenario and writes the report. It returns an error only when
// the run could not be carried out — a run in which the system under test
// performed badly is a successful run with bad numbers.
func (r *Runner) Run(ctx context.Context) error {
	pool, err := client.NewPool(ctx, client.Options{
		Address:        r.cfg.Target.GRPCAddress,
		Connections:    r.cfg.Target.Connections,
		TLS:            r.cfg.Target.TLS,
		TLSSkipVerify:  r.cfg.Target.TLSSkipVerify,
		DialTimeout:    r.cfg.Target.DialTimeout,
		RequestTimeout: r.cfg.Target.RequestTimeout,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	authClient := auth.New(r.cfg.Target.KratosURL, r.cfg.Target.DialTimeout)

	fmt.Fprintf(r.out, "preparing %d senders and %d receivers against %s\n",
		len(r.cfg.Senders), len(r.cfg.Receivers), r.cfg.Target.GRPCAddress)

	senders, err := setup(ctx, r.cfg, pool, authClient)
	if err != nil {
		return fmt.Errorf("setup failed:\n%w", err)
	}

	r.publishRunInfo()
	for _, s := range senders {
		r.rec.SetStartBalance(s.cfg.Label, s.startBalance)
	}

	runCtx, cancel := r.runContext(ctx)
	defer cancel()

	measuring := r.warmupGate()
	watch := newWatcher(r.cfg, r.rec, r.coll, measuring)
	watch.start(runCtx)

	limiter := r.limiter()

	fmt.Fprintf(r.out, "starting run (%s, %s pairing)\n", r.cfg.Run.Stop, r.cfg.Run.Pairing)
	r.rec.Start()

	stopProgress := r.startProgress(runCtx, watch, senders)
	fatal := r.drive(runCtx, senders, limiter, watch, measuring, cancel)

	// Senders are done issuing payments; wait for the in-flight ones to settle
	// before closing the books, otherwise the settled count is meaningless.
	if r.cfg.Settlement.Track {
		inFlightCount, _ := watch.stats()
		if inFlightCount > 0 {
			fmt.Fprintf(r.out, "senders finished; waiting for %d in-flight payments to settle\n", inFlightCount)
		}
	}
	watch.close()

	r.rec.Finish()
	stopProgress()

	r.recordEndBalances(ctx, senders)
	r.report()

	if fatal != nil {
		return fatal
	}
	return nil
}

// drive runs one goroutine per sender wallet and waits for them all.
//
// A goroutine per virtual wallet is the whole reason this is a Go harness: each
// sender keeps its own session, its own receiver cursor and its own stop
// condition, while a shared limiter controls the aggregate arrival rate.
func (r *Runner) drive(
	ctx context.Context,
	senders []*sender,
	limiter *rate.Limiter,
	watch *watcher,
	measuring func() bool,
	cancel context.CancelFunc,
) error {
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		fatals []error
		active atomic.Int64
	)

	active.Store(int64(len(senders)))
	r.coll.SetSendersActive(len(senders))

	for _, s := range senders {
		wg.Go(func() {
			defer func() {
				r.coll.SetSendersActive(int(active.Add(-1)))
			}()

			loop := &senderLoop{
				cfg:     r.cfg,
				sender:  s,
				rec:     r.rec,
				coll:    r.coll,
				limiter: limiter,
				watch:   watch,
				// One gate shared by every sender and the watcher, so the warmup
				// window is a single instant rather than per-goroutine.
				measuring: measuring,
			}

			reason, err := loop.run(ctx)
			r.rec.SetStopReason(s.cfg.Label, reason)

			if err != nil {
				mu.Lock()
				fatals = append(fatals, err)
				mu.Unlock()
				// A fatal failure means the harness is misconfigured, so every other
				// sender is producing noise too. Stop the run rather than fill the
				// report with the same error a hundred times.
				cancel()
			}
		})
	}

	wg.Wait()

	if len(fatals) > 0 {
		return errors.Join(fatals...)
	}
	return nil
}

// runContext applies the run's wall-clock ceiling, when one is configured.
func (r *Runner) runContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if r.cfg.Run.Duration > 0 {
		return context.WithTimeout(ctx, r.cfg.Run.Duration)
	}
	return context.WithCancel(ctx)
}

// limiter returns the shared arrival-rate limiter, or nil when unthrottled.
func (r *Runner) limiter() *rate.Limiter {
	if r.cfg.Run.ArrivalRate <= 0 {
		return nil
	}
	return rate.NewLimiter(rate.Limit(r.cfg.Run.ArrivalRate), r.cfg.Run.Burst)
}

// warmupGate returns a predicate that reports whether the warmup window has
// elapsed. Call it once per run and share the result: a gate built per goroutine
// would give each sender its own slightly different warmup deadline.
func (r *Runner) warmupGate() func() bool {
	if r.cfg.Run.Warmup <= 0 {
		return func() bool { return true }
	}
	until := time.Now().Add(r.cfg.Run.Warmup)
	return func() bool { return time.Now().After(until) }
}

// startProgress prints a periodic status line and returns a stop function.
func (r *Runner) startProgress(ctx context.Context, watch *watcher, senders []*sender) func() {
	done := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		ticker := time.NewTicker(progressInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				snap := r.rec.Snapshot()
				var confirmed int
				for _, st := range snap.Stages {
					if st.Stage == metrics.StageConfirm {
						confirmed = st.Count()
					}
				}
				inFlightCount, settled := watch.stats()
				metrics.WriteBriefProgress(r.out, r.rec.Elapsed(), confirmed, settled, inFlightCount, len(senders))
			case <-ctx.Done():
				return
			case <-done:
				return
			}
		}
	}()

	return func() {
		close(done)
		<-stopped
	}
}

// recordEndBalances reads each sender's closing balance so the report can show
// how much value actually moved. It uses the parent context because the run
// context may already have expired.
func (r *Runner) recordEndBalances(ctx context.Context, senders []*sender) {
	var wg sync.WaitGroup
	for _, s := range senders {
		wg.Go(func() {
			// Default to the opening balance so a failed lookup reports a zero draw
			// rather than appearing to have drained the whole wallet.
			r.rec.SetEndBalance(s.cfg.Label, s.startBalance)

			balances, err := s.wallet.GetBalances(ctx)
			if err != nil {
				// Not worth failing the run over — the numbers that matter are already
				// recorded.
				return
			}
			for _, b := range balances {
				if b.GetLinkedAccount() == s.linkedAccount {
					r.rec.SetEndBalance(s.cfg.Label, b.GetBalance().GetAmount())
					return
				}
			}
		})
	}
	wg.Wait()
}

func (r *Runner) publishRunInfo() {
	rateLabel := "unthrottled"
	if r.cfg.Run.ArrivalRate > 0 {
		rateLabel = strconv.FormatFloat(r.cfg.Run.ArrivalRate, 'f', -1, 64)
	}
	r.coll.SetRunInfo(
		string(r.cfg.Run.Stop),
		string(r.cfg.Run.Pairing),
		strconv.Itoa(len(r.cfg.Senders)),
		strconv.Itoa(len(r.cfg.Receivers)),
		rateLabel,
	)
}

func (r *Runner) report() {
	if !r.cfg.Metrics.Console {
		return
	}

	asset := r.cfg.Run.Asset
	if asset == "" {
		asset = "minor units"
	}

	metrics.WriteReport(r.out, metrics.RunHeader{
		Target:      r.cfg.Target.GRPCAddress,
		Stop:        string(r.cfg.Run.Stop),
		Pairing:     string(r.cfg.Run.Pairing),
		Senders:     len(r.cfg.Senders),
		Receivers:   len(r.cfg.Receivers),
		Amount:      r.cfg.Run.Amount,
		Asset:       asset,
		ArrivalRate: r.cfg.Run.ArrivalRate,
		MaxInFlight: r.cfg.Run.MaxInFlight,
		Settlement:  r.cfg.Settlement.Track,
	}, r.rec.Snapshot())
}
