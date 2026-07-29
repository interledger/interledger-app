package metrics

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

// RunHeader describes the run being reported.
type RunHeader struct {
	Target    string
	Stop      string
	Pairing   string
	Senders   int
	Receivers int
	// Amount is the per-payment value in minor units, with its asset.
	Amount int64
	Asset  string
	// ArrivalRate is the configured target rate, zero when unthrottled.
	ArrivalRate float64
	MaxInFlight int
	Settlement  bool
}

// WriteReport prints a human-readable summary of a finished run.
//
// The structure is deliberate: throughput first (the headline), then latency
// split by stage so an RPC-level regression is not mistaken for a settlement
// regression, then errors, then per-wallet totals to expose uneven load.
func WriteReport(w io.Writer, h RunHeader, s Snapshot) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "════ interledger performance run ════")
	fmt.Fprintln(w)

	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, "target\t%s\n", h.Target)
	fmt.Fprintf(tw, "shape\t%d senders → %d receivers (%s)\n", h.Senders, h.Receivers, h.Pairing)
	fmt.Fprintf(tw, "stop\t%s\n", h.Stop)
	fmt.Fprintf(tw, "amount\t%d %s per payment\n", h.Amount, h.Asset)
	if h.ArrivalRate > 0 {
		fmt.Fprintf(tw, "arrival rate\t%.1f/s target\n", h.ArrivalRate)
	} else {
		fmt.Fprintf(tw, "arrival rate\tunthrottled\n")
	}
	fmt.Fprintf(tw, "max in flight\t%d\n", h.MaxInFlight)
	fmt.Fprintf(tw, "settlement tracking\t%t\n", h.Settlement)
	fmt.Fprintf(tw, "duration\t%s\n", roundDur(s.Elapsed))
	tw.Flush()

	writeThroughput(w, h, s)
	writeLatency(w, s)
	writeErrors(w, s)
	writeSenders(w, s)

	if !h.Settlement {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "note: settlement tracking is off, so the latencies above are RPC round trips only.")
		fmt.Fprintln(w, "      CreatePayment starts a Temporal workflow and returns before any money moves,")
		fmt.Fprintln(w, "      so these numbers do not describe how fast transactions actually settle.")
	}
}

func writeThroughput(w io.Writer, h RunHeader, s Snapshot) {
	completed := s.Outcomes[OutcomeCompleted]
	failed := s.Outcomes[OutcomeFailed]
	timedOut := s.Outcomes[OutcomeTimedOut]
	rejected := s.Outcomes[OutcomeRejected]

	var confirmed int
	for _, st := range s.Stages {
		if st.Stage == StageConfirm {
			confirmed = st.Count()
		}
	}

	secs := s.Elapsed.Seconds()

	fmt.Fprintln(w)
	fmt.Fprintln(w, "── throughput ──")
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, "\tcount\tper second")
	fmt.Fprintf(tw, "confirmed\t%d\t%s\n", confirmed, rate(confirmed, secs))
	if h.Settlement {
		fmt.Fprintf(tw, "settled (completed)\t%d\t%s\n", completed, rate(completed, secs))
		if failed > 0 {
			fmt.Fprintf(tw, "settled (failed)\t%d\t%s\n", failed, rate(failed, secs))
		}
		if timedOut > 0 {
			fmt.Fprintf(tw, "never settled\t%d\t%s\n", timedOut, rate(timedOut, secs))
		}
	}
	if rejected > 0 {
		fmt.Fprintf(tw, "rejected\t%d\t%s\n", rejected, rate(rejected, secs))
	}
	tw.Flush()

	if h.Settlement && confirmed > 0 && completed < confirmed {
		fmt.Fprintf(w, "\n%d of %d confirmed payments did not reach Completed — settled throughput,\nnot confirm throughput, is the real capacity figure.\n", confirmed-completed, confirmed)
	}
}

func writeLatency(w io.Writer, s Snapshot) {
	if len(s.Stages) == 0 {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "── latency ──")
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, "stage\tn\tmin\tmean\tp50\tp95\tp99\tmax")
	for _, st := range s.Stages {
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			stageLabel(st.Stage),
			st.Count(),
			roundDur(st.Min()),
			roundDur(st.Mean()),
			roundDur(st.Quantile(0.50)),
			roundDur(st.Quantile(0.95)),
			roundDur(st.Quantile(0.99)),
			roundDur(st.Max()),
		)
	}
	tw.Flush()
}

func writeErrors(w io.Writer, s Snapshot) {
	if len(s.Errors) == 0 {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "── errors ──")
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, "stage\tclass\tcode\tcount")
	for _, e := range s.Errors {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n", e.Stage, e.Class, e.Code, e.Count)
	}
	tw.Flush()
}

func writeSenders(w io.Writer, s Snapshot) {
	if len(s.Senders) == 0 {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "── senders ──")
	tw := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, "wallet\tattempted\tconfirmed\tcompleted\tbalance drawn\tstopped because")
	for _, snd := range s.Senders {
		drawn := snd.StartBalance - snd.EndBalance
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%s\n",
			truncate(snd.Label, 36),
			snd.Attempted,
			snd.Confirmed,
			snd.Completed,
			drawn,
			snd.StopReason,
		)
	}
	tw.Flush()
}

func stageLabel(s Stage) string {
	if s == StageSettle {
		return "settle (confirm → terminal)"
	}
	return string(s)
}

func rate(n int, secs float64) string {
	if secs <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f", float64(n)/secs)
}

// roundDur trims durations to a readable precision: sub-second values keep
// microseconds, longer ones keep milliseconds.
func roundDur(d time.Duration) string {
	switch {
	case d == 0:
		return "-"
	case d < time.Millisecond:
		return d.Round(time.Microsecond).String()
	case d < time.Second:
		return d.Round(100 * time.Microsecond).String()
	default:
		return d.Round(time.Millisecond).String()
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

// WriteBriefProgress renders a single progress line, used while a run is in
// flight so a long drain does not look like a hang.
func WriteBriefProgress(w io.Writer, elapsed time.Duration, confirmed, settled, inFlight, activeSenders int) {
	fmt.Fprintf(w, "[%s] confirmed=%d settled=%d in-flight=%d senders=%d\n",
		roundDur(elapsed), confirmed, settled, inFlight, activeSenders)
}

// Separator returns a horizontal rule of n characters, for callers assembling
// their own output around a report.
func Separator(n int) string { return strings.Repeat("─", n) }
