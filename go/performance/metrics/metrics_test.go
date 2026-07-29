package metrics

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gather reads a metric family straight off the run's registry. Deliberately not
// prometheus/testutil: that package pulls an extra module into go.mod for the
// sake of a handful of assertions.
func gather(t *testing.T, c *Collectors, name string) *dto.MetricFamily {
	t.Helper()

	families, err := c.Registry().Gather()
	require.NoError(t, err)

	for _, f := range families {
		if f.GetName() == name {
			return f
		}
	}
	return nil
}

// counterFor returns the value of the counter whose label values match, in order.
func counterFor(t *testing.T, c *Collectors, name string, labelValues ...string) float64 {
	t.Helper()

	family := gather(t, c, name)
	require.NotNil(t, family, "metric family %s not registered", name)

	for _, m := range family.GetMetric() {
		if matchesLabels(m, labelValues) {
			return m.GetCounter().GetValue()
		}
	}
	t.Fatalf("no %s series with labels %v", name, labelValues)
	return 0
}

func matchesLabels(m *dto.Metric, want []string) bool {
	got := make([]string, 0, len(m.GetLabel()))
	for _, l := range m.GetLabel() {
		// job_name is a constant label on every series and is not part of what the
		// caller is selecting on.
		if l.GetName() == "job_name" {
			continue
		}
		got = append(got, l.GetValue())
	}
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func gaugeValue(t *testing.T, c *Collectors, name string) float64 {
	t.Helper()

	family := gather(t, c, name)
	require.NotNil(t, family, "metric family %s not registered", name)
	require.Len(t, family.GetMetric(), 1)
	return family.GetMetric()[0].GetGauge().GetValue()
}

func TestCollectorsCountPayments(t *testing.T) {
	c := NewCollectors("test-job")

	c.CountPayment(OutcomeCompleted)
	c.CountPayment(OutcomeCompleted)
	c.CountPayment(OutcomeFailed)

	assert.Equal(t, 2.0, counterFor(t, c, "perf_payments_total", "completed"))
	assert.Equal(t, 1.0, counterFor(t, c, "perf_payments_total", "failed"))
}

func TestCollectorsCountErrorsByCode(t *testing.T) {
	c := NewCollectors("test-job")

	c.CountError(StageCreate, "exhausted", "PAYMENTS_INSUFFICIENT_FUNDS")
	c.CountError(StageCreate, "exhausted", "PAYMENTS_INSUFFICIENT_FUNDS")

	// Labels are sorted by name in the exposition: class, code, stage.
	got := counterFor(t, c, "perf_errors_total", "exhausted", "PAYMENTS_INSUFFICIENT_FUNDS", "create")
	assert.Equal(t, 2.0, got)
}

func TestCollectorsGauges(t *testing.T) {
	c := NewCollectors("test-job")

	c.SetInFlight(42)
	c.SetSendersActive(7)

	assert.Equal(t, 42.0, gaugeValue(t, c, "perf_payments_in_flight"))
	assert.Equal(t, 7.0, gaugeValue(t, c, "perf_senders_active"))
}

func TestCollectorsRunInfo(t *testing.T) {
	c := NewCollectors("test-job")
	c.SetRunInfo("drain", "index", "100", "100", "unthrottled")

	family := gather(t, c, "perf_run_info")
	require.NotNil(t, family)
	require.Len(t, family.GetMetric(), 1)
	assert.Equal(t, 1.0, family.GetMetric()[0].GetGauge().GetValue())
}

func TestCollectorsRecordLatencies(t *testing.T) {
	c := NewCollectors("test-job")

	c.ObserveRPC(StageCreate, 10*time.Millisecond)
	c.ObserveRPC(StageCreate, 20*time.Millisecond)
	c.ObserveSettlement(1500 * time.Millisecond)

	rpc := gather(t, c, "perf_rpc_duration_seconds")
	require.NotNil(t, rpc)
	require.Len(t, rpc.GetMetric(), 1, "one series per stage")
	assert.Equal(t, uint64(2), rpc.GetMetric()[0].GetHistogram().GetSampleCount())
	assert.InDelta(t, 0.03, rpc.GetMetric()[0].GetHistogram().GetSampleSum(), 0.0001)

	settle := gather(t, c, "perf_settlement_duration_seconds")
	require.NotNil(t, settle)
	assert.Equal(t, uint64(1), settle.GetMetric()[0].GetHistogram().GetSampleCount())
}

func TestCollectorsUseAPrivateRegistry(t *testing.T) {
	// Two runs in one process must not fight over metric registration.
	a := NewCollectors("job-a")
	b := NewCollectors("job-b")

	assert.NotSame(t, a.Registry(), b.Registry())

	a.CountPayment(OutcomeCompleted)
	assert.Equal(t, 1.0, counterFor(t, a, "perf_payments_total", "completed"))
	assert.Nil(t, gather(t, b, "perf_payments_total"), "b saw none of a's samples")
}

func TestServeExposesMetrics(t *testing.T) {
	c := NewCollectors("serve-test")
	c.CountPayment(OutcomeCompleted)
	c.ObserveSettlement(1500 * time.Millisecond)

	const addr = "127.0.0.1:19464"
	srv := Serve(addr, c)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})

	var body string
	require.Eventually(t, func() bool {
		resp, err := http.Get("http://" + addr + "/metrics")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return false
		}
		body = string(b)
		return resp.StatusCode == http.StatusOK
	}, 3*time.Second, 50*time.Millisecond, "metrics endpoint did not come up")

	assert.Contains(t, body, "perf_payments_total")
	assert.Contains(t, body, `job_name="serve-test"`)
	assert.Contains(t, body, "perf_settlement_duration_seconds")
}
