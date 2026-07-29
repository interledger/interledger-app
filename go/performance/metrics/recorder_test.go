package metrics

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecorderQuantilesAreExact(t *testing.T) {
	r := NewRecorder()

	// 1ms..100ms, so the expected quantiles are unambiguous.
	for i := 1; i <= 100; i++ {
		r.Observe(StageCreate, time.Duration(i)*time.Millisecond)
	}

	snap := r.Snapshot()
	require.Len(t, snap.Stages, 1)

	s := snap.Stages[0]
	assert.Equal(t, 100, s.Count())
	assert.Equal(t, time.Millisecond, s.Min())
	assert.Equal(t, 100*time.Millisecond, s.Max())
	assert.Equal(t, 51*time.Millisecond, s.Quantile(0.50))
	assert.Equal(t, 96*time.Millisecond, s.Quantile(0.95))
	assert.Equal(t, 100*time.Millisecond, s.Quantile(0.99))

	// Mean of 1..100 ms.
	assert.Equal(t, 50500*time.Microsecond, s.Mean())
}

func TestRecorderQuantilesOnEmptySamples(t *testing.T) {
	s := &Samples{Stage: StageCreate}
	assert.Equal(t, time.Duration(0), s.Quantile(0.99))
	assert.Equal(t, time.Duration(0), s.Min())
	assert.Equal(t, time.Duration(0), s.Max())
	assert.Equal(t, time.Duration(0), s.Mean())
}

func TestRecorderOrdersStagesByLifecycle(t *testing.T) {
	r := NewRecorder()

	// Recorded out of order; the report must still read create → confirm → settle.
	r.Observe(StageSettle, time.Second)
	r.Observe(StageCreate, time.Millisecond)
	r.Observe(StageConfirm, 2*time.Millisecond)

	snap := r.Snapshot()
	require.Len(t, snap.Stages, 3)
	assert.Equal(t, StageCreate, snap.Stages[0].Stage)
	assert.Equal(t, StageConfirm, snap.Stages[1].Stage)
	assert.Equal(t, StageSettle, snap.Stages[2].Stage)
}

func TestRecorderErrorsSortedByCount(t *testing.T) {
	r := NewRecorder()

	r.CountError(StageCreate, "transient", "Unavailable")
	r.CountError(StageCreate, "transient", "Unavailable")
	r.CountError(StageCreate, "transient", "Unavailable")
	r.CountError(StageConfirm, "exhausted", "PAYMENTS_INSUFFICIENT_FUNDS")

	snap := r.Snapshot()
	require.Len(t, snap.Errors, 2)
	assert.Equal(t, "Unavailable", snap.Errors[0].Code, "most frequent first")
	assert.Equal(t, 3, snap.Errors[0].Count)
	assert.Equal(t, 1, snap.Errors[1].Count)
}

func TestRecorderTracksPerSenderTotals(t *testing.T) {
	r := NewRecorder()

	r.SetStartBalance("wallet-a", 500)
	r.AddAttempt("wallet-a")
	r.AddAttempt("wallet-a")
	r.AddConfirmed("wallet-a")
	r.AddCompleted("wallet-a")
	r.SetEndBalance("wallet-a", 499)
	r.SetStopReason("wallet-a", "insufficient funds")

	snap := r.Snapshot()
	require.Len(t, snap.Senders, 1)

	s := snap.Senders[0]
	assert.Equal(t, "wallet-a", s.Label)
	assert.Equal(t, 2, s.Attempted)
	assert.Equal(t, 1, s.Confirmed)
	assert.Equal(t, 1, s.Completed)
	assert.Equal(t, int64(500), s.StartBalance)
	assert.Equal(t, int64(499), s.EndBalance)
	assert.Equal(t, "insufficient funds", s.StopReason)
}

func TestRecorderSendersSortedByLabel(t *testing.T) {
	r := NewRecorder()
	for _, label := range []string{"c", "a", "b"} {
		r.AddAttempt(label)
	}

	snap := r.Snapshot()
	require.Len(t, snap.Senders, 3)
	assert.Equal(t, "a", snap.Senders[0].Label)
	assert.Equal(t, "b", snap.Senders[1].Label)
	assert.Equal(t, "c", snap.Senders[2].Label)
}

func TestRecorderIsSafeUnderConcurrency(t *testing.T) {
	// The real usage: one goroutine per sender wallet plus a pool of settlement
	// pollers, all writing at once.
	r := NewRecorder()
	r.Start()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			label := string(rune('a' + i%26))
			for range 100 {
				r.Observe(StageCreate, time.Millisecond)
				r.AddAttempt(label)
				r.CountOutcome(OutcomeCompleted)
				r.CountError(StageCreate, "transient", "Unavailable")
			}
		})
	}
	wg.Wait()
	r.Finish()

	snap := r.Snapshot()
	require.Len(t, snap.Stages, 1)
	assert.Equal(t, 5000, snap.Stages[0].Count())
	assert.Equal(t, 5000, snap.Outcomes[OutcomeCompleted])
	require.Len(t, snap.Errors, 1)
	assert.Equal(t, 5000, snap.Errors[0].Count)
	assert.Positive(t, snap.Elapsed)
}

func TestRecorderElapsedBeforeStart(t *testing.T) {
	r := NewRecorder()
	assert.Equal(t, time.Duration(0), r.Elapsed())
	assert.Equal(t, time.Duration(0), r.Snapshot().Elapsed)
}

func TestWriteReportIncludesTheHeadlineNumbers(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.Observe(StageCreate, 10*time.Millisecond)
	r.Observe(StageConfirm, 20*time.Millisecond)
	r.Observe(StageSettle, 1500*time.Millisecond)
	r.CountOutcome(OutcomeCompleted)
	r.SetStartBalance("wallet-a", 100)
	r.SetEndBalance("wallet-a", 99)
	r.AddAttempt("wallet-a")
	r.SetStopReason("wallet-a", "balance exhausted")
	r.Finish()

	var buf bytes.Buffer
	WriteReport(&buf, RunHeader{
		Target:      "localhost:8443",
		Stop:        "drain",
		Pairing:     "index",
		Senders:     1,
		Receivers:   1,
		Amount:      1,
		Asset:       "ZAR",
		MaxInFlight: 256,
		Settlement:  true,
	}, r.Snapshot())

	out := buf.String()
	assert.Contains(t, out, "localhost:8443")
	assert.Contains(t, out, "1 senders → 1 receivers (index)")
	assert.Contains(t, out, "throughput")
	assert.Contains(t, out, "settled (completed)")
	assert.Contains(t, out, "settle (confirm → terminal)")
	assert.Contains(t, out, "balance exhausted")
	assert.NotContains(t, out, "settlement tracking is off")
}

func TestWriteReportWarnsWhenSettlementIsOff(t *testing.T) {
	// The whole reason this harness exists: RPC latency alone is misleading, so
	// the report has to say so rather than let the number stand unqualified.
	r := NewRecorder()
	r.Start()
	r.Observe(StageCreate, 10*time.Millisecond)
	r.Finish()

	var buf bytes.Buffer
	WriteReport(&buf, RunHeader{Target: "localhost:8443", Settlement: false}, r.Snapshot())

	out := buf.String()
	assert.Contains(t, out, "settlement tracking is off")
	assert.Contains(t, out, "starts a Temporal workflow")
}

func TestWriteReportFlagsUnsettledPayments(t *testing.T) {
	r := NewRecorder()
	r.Start()
	for range 10 {
		r.Observe(StageConfirm, time.Millisecond)
	}
	for range 7 {
		r.CountOutcome(OutcomeCompleted)
	}
	for range 3 {
		r.CountOutcome(OutcomeTimedOut)
	}
	r.Finish()

	var buf bytes.Buffer
	WriteReport(&buf, RunHeader{Target: "t", Settlement: true}, r.Snapshot())

	out := buf.String()
	assert.Contains(t, out, "3 of 10 confirmed payments did not reach Completed")
	assert.Contains(t, out, "never settled")
}

func TestRoundDurFormats(t *testing.T) {
	assert.Equal(t, "-", roundDur(0))
	assert.Contains(t, roundDur(500*time.Microsecond), "µs")
	assert.Equal(t, "1.5s", roundDur(1500*time.Millisecond))
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "short", truncate("short", 10))
	assert.Equal(t, "abc…", truncate("abcdefgh", 4))
}

func TestSeparator(t *testing.T) {
	assert.Equal(t, 4, len([]rune(Separator(4))))
	assert.True(t, strings.HasPrefix(Separator(3), "─"))
}

func TestWriteBriefProgress(t *testing.T) {
	var buf bytes.Buffer
	WriteBriefProgress(&buf, 12*time.Second, 100, 90, 10, 5)

	out := buf.String()
	assert.Contains(t, out, "confirmed=100")
	assert.Contains(t, out, "settled=90")
	assert.Contains(t, out, "in-flight=10")
	assert.Contains(t, out, "senders=5")
}
