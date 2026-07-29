// Package metrics records and reports performance-run results.
//
// Two sinks, deliberately: an in-process recorder that produces the console
// summary (exact percentiles over every sample, no bucket approximation), and
// Prometheus collectors scraped by the local monitoring stack so a run's numbers
// sit on the same Grafana dashboard as the backend's own metrics and Tempo
// traces. Neither depends on the other.
package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Stage names the step of the payment lifecycle a measurement belongs to.
type Stage string

const (
	StageCreate  Stage = "create"
	StageUpdate  Stage = "update"
	StageConfirm Stage = "confirm"
	StagePoll    Stage = "poll"
	// StageSettle is not an RPC: it is the wall-clock time from a confirmed
	// payment to its terminal state, which is the number that actually describes
	// how fast the system moves money.
	StageSettle Stage = "settle"
)

// Outcome is the terminal disposition of a payment.
type Outcome string

const (
	OutcomeCompleted Outcome = "completed"
	OutcomeFailed    Outcome = "failed"
	OutcomeTimedOut  Outcome = "timed_out"
	OutcomeRejected  Outcome = "rejected"
)

// Collectors holds the Prometheus metrics for a run.
type Collectors struct {
	registry *prometheus.Registry

	rpcDuration    *prometheus.HistogramVec
	settleDuration prometheus.Histogram
	payments       *prometheus.CounterVec
	errorsTotal    *prometheus.CounterVec
	inFlight       prometheus.Gauge
	sendersActive  prometheus.Gauge
	runInfo        *prometheus.GaugeVec
}

// NewCollectors registers the run's metrics on a private registry, so a perf run
// exports only its own series.
func NewCollectors(job string) *Collectors {
	labels := prometheus.Labels{"job_name": job}
	reg := prometheus.NewRegistry()

	c := &Collectors{
		registry: reg,
		rpcDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "perf_rpc_duration_seconds",
			Help:        "Latency of individual backend RPCs issued by the performance harness.",
			ConstLabels: labels,
			// Native histograms give useful percentiles across the three orders of
			// magnitude a loaded backend can span without picking buckets by hand.
			NativeHistogramBucketFactor:    1.1,
			NativeHistogramMaxBucketNumber: 160,
			Buckets:                        prometheus.ExponentialBuckets(0.005, 2, 12),
		}, []string{"stage"}),
		settleDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:                           "perf_settlement_duration_seconds",
			Help:                           "Time from a confirmed payment to a terminal state.",
			ConstLabels:                    labels,
			NativeHistogramBucketFactor:    1.1,
			NativeHistogramMaxBucketNumber: 160,
			Buckets:                        prometheus.ExponentialBuckets(0.05, 2, 14),
		}),
		payments: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "perf_payments_total",
			Help:        "Payments by terminal outcome.",
			ConstLabels: labels,
		}, []string{"outcome"}),
		errorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "perf_errors_total",
			Help:        "Failed RPCs by stage, classification and backend error code.",
			ConstLabels: labels,
		}, []string{"stage", "class", "code"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "perf_payments_in_flight",
			Help:        "Payments confirmed but not yet terminal.",
			ConstLabels: labels,
		}),
		sendersActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "perf_senders_active",
			Help:        "Sender wallets still issuing payments.",
			ConstLabels: labels,
		}),
		runInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "perf_run_info",
			Help:        "Static description of the run; value is always 1.",
			ConstLabels: labels,
		}, []string{"stop", "pairing", "senders", "receivers", "arrival_rate"}),
	}

	reg.MustRegister(
		c.rpcDuration,
		c.settleDuration,
		c.payments,
		c.errorsTotal,
		c.inFlight,
		c.sendersActive,
		c.runInfo,
	)

	return c
}

// Registry exposes the run's registry for serving or for tests.
func (c *Collectors) Registry() *prometheus.Registry { return c.registry }

// SetRunInfo publishes the run's shape as a labelled constant series.
func (c *Collectors) SetRunInfo(stop, pairing, senders, receivers, arrivalRate string) {
	c.runInfo.WithLabelValues(stop, pairing, senders, receivers, arrivalRate).Set(1)
}

// ObserveRPC records the latency of one RPC.
func (c *Collectors) ObserveRPC(stage Stage, d time.Duration) {
	c.rpcDuration.WithLabelValues(string(stage)).Observe(d.Seconds())
}

// ObserveSettlement records how long a payment took to reach a terminal state.
func (c *Collectors) ObserveSettlement(d time.Duration) {
	c.settleDuration.Observe(d.Seconds())
}

// CountPayment records a terminal payment outcome.
func (c *Collectors) CountPayment(o Outcome) {
	c.payments.WithLabelValues(string(o)).Inc()
}

// CountError records a failed RPC.
func (c *Collectors) CountError(stage Stage, class, code string) {
	c.errorsTotal.WithLabelValues(string(stage), class, code).Inc()
}

// SetInFlight publishes the current in-flight payment count.
func (c *Collectors) SetInFlight(n int) { c.inFlight.Set(float64(n)) }

// SetSendersActive publishes the number of senders still working.
func (c *Collectors) SetSendersActive(n int) { c.sendersActive.Set(float64(n)) }

// Server serves the run's /metrics endpoint.
type Server struct {
	http *http.Server
	errs chan error
}

// Serve starts a metrics endpoint on addr. The caller must call Shutdown.
func Serve(addr string, c *Collectors) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{
		// Required for native histogram exposition.
		EnableOpenMetrics: true,
	}))

	srv := &Server{
		http: &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second},
		errs: make(chan error, 1),
	}

	go func() {
		err := srv.http.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.errs <- err
			return
		}
		close(srv.errs)
	}()

	return srv
}

// Err returns a channel that yields a listener error, if the endpoint failed to
// start. A closed channel means a clean shutdown.
func (s *Server) Err() <-chan error { return s.errs }

// Shutdown stops the metrics endpoint.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
