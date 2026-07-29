package metrics

import (
	"cmp"
	"maps"
	"slices"
	"sync"
	"time"
)

// Recorder accumulates every sample of a run in memory.
//
// Keeping raw samples rather than buckets means the reported percentiles are
// exact — which matters when the whole point of the exercise is to argue about
// whether p99 moved. A drain run over a hundred wallets produces samples in the
// tens of thousands, so the memory cost is trivial.
type Recorder struct {
	mu sync.Mutex

	start time.Time
	end   time.Time

	stages   map[Stage]*Samples
	outcomes map[Outcome]int
	errors   map[errorKey]int

	// senders tracks per-wallet progress so the report can show whether load was
	// even or a few wallets did all the work.
	senders map[string]*SenderStat
}

type errorKey struct {
	Stage Stage
	Class string
	Code  string
}

// SenderStat is one sender wallet's contribution to the run.
type SenderStat struct {
	Label string
	// Attempted is payments the sender tried to create.
	Attempted int
	// Confirmed is payments accepted by the backend through confirm.
	Confirmed int
	// Completed is payments that reached a Completed terminal state.
	Completed int
	// StartBalance and EndBalance are the sender's balance in minor units before
	// and after the run.
	StartBalance int64
	EndBalance   int64
	// StopReason explains why this sender stopped.
	StopReason string
}

// NewRecorder returns an empty Recorder.
func NewRecorder() *Recorder {
	return &Recorder{
		stages:   make(map[Stage]*Samples),
		outcomes: make(map[Outcome]int),
		errors:   make(map[errorKey]int),
		senders:  make(map[string]*SenderStat),
	}
}

// Start marks the beginning of the measured window.
func (r *Recorder) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.start = time.Now()
}

// Finish marks the end of the measured window.
func (r *Recorder) Finish() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.end = time.Now()
}

// Elapsed is the length of the measured window.
func (r *Recorder) Elapsed() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.start.IsZero() {
		return 0
	}
	end := r.end
	if end.IsZero() {
		end = time.Now()
	}
	return end.Sub(r.start)
}

// Observe records a latency sample for a stage.
func (r *Recorder) Observe(stage Stage, d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.stages[stage]
	if !ok {
		s = &Samples{Stage: stage}
		r.stages[stage] = s
	}
	s.values = append(s.values, d)
}

// CountOutcome records a terminal payment outcome.
func (r *Recorder) CountOutcome(o Outcome) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.outcomes[o]++
}

// CountError records a failed RPC by stage, class and backend error code.
func (r *Recorder) CountError(stage Stage, class, code string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors[errorKey{Stage: stage, Class: class, Code: code}]++
}

// Sender returns the mutable stats for a sender, creating them on first use.
// The returned pointer must only be mutated through the Recorder's helpers below.
func (r *Recorder) sender(label string) *SenderStat {
	s, ok := r.senders[label]
	if !ok {
		s = &SenderStat{Label: label}
		r.senders[label] = s
	}
	return s
}

// AddAttempt records that a sender tried to create a payment.
func (r *Recorder) AddAttempt(label string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sender(label).Attempted++
}

// AddConfirmed records that a sender got a payment confirmed.
func (r *Recorder) AddConfirmed(label string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sender(label).Confirmed++
}

// AddCompleted records that one of a sender's payments settled.
func (r *Recorder) AddCompleted(label string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sender(label).Completed++
}

// SetStartBalance records a sender's opening balance.
func (r *Recorder) SetStartBalance(label string, minorUnits int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sender(label).StartBalance = minorUnits
}

// SetEndBalance records a sender's closing balance.
func (r *Recorder) SetEndBalance(label string, minorUnits int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sender(label).EndBalance = minorUnits
}

// SetStopReason records why a sender stopped sending.
func (r *Recorder) SetStopReason(label, reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sender(label).StopReason = reason
}

// Samples is the set of latency samples for one stage.
type Samples struct {
	Stage  Stage
	values []time.Duration
}

// Count is the number of samples.
func (s *Samples) Count() int { return len(s.values) }

// Quantile returns the q-th quantile (0 < q < 1) by nearest rank.
//
// Only meaningful on samples obtained from Recorder.Snapshot, which sorts them.
// The same applies to Min and Max.
func (s *Samples) Quantile(q float64) time.Duration {
	if len(s.values) == 0 {
		return 0
	}
	idx := int(q * float64(len(s.values)))
	if idx >= len(s.values) {
		idx = len(s.values) - 1
	}
	return s.values[idx]
}

// Min returns the smallest sample.
func (s *Samples) Min() time.Duration {
	if len(s.values) == 0 {
		return 0
	}
	return s.values[0]
}

// Max returns the largest sample.
func (s *Samples) Max() time.Duration {
	if len(s.values) == 0 {
		return 0
	}
	return s.values[len(s.values)-1]
}

// Mean returns the arithmetic mean of the samples.
func (s *Samples) Mean() time.Duration {
	if len(s.values) == 0 {
		return 0
	}
	var total time.Duration
	for _, v := range s.values {
		total += v
	}
	return total / time.Duration(len(s.values))
}

func (s *Samples) sort() { slices.Sort(s.values) }

// Snapshot is an immutable view of a finished run, safe to read without locking.
type Snapshot struct {
	Elapsed  time.Duration
	Stages   []*Samples
	Outcomes map[Outcome]int
	Errors   []ErrorCount
	Senders  []*SenderStat
}

// ErrorCount is one grouped failure total.
type ErrorCount struct {
	Stage Stage
	Class string
	Code  string
	Count int
}

// Snapshot sorts the samples and returns a stable view of the run.
func (r *Recorder) Snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Stages are reported in lifecycle order, not map order.
	order := []Stage{StageCreate, StageUpdate, StageConfirm, StagePoll, StageSettle}
	stages := make([]*Samples, 0, len(r.stages))
	for _, stage := range order {
		s, ok := r.stages[stage]
		if !ok || len(s.values) == 0 {
			continue
		}
		s.sort()
		stages = append(stages, s)
	}

	outcomes := make(map[Outcome]int, len(r.outcomes))
	maps.Copy(outcomes, r.outcomes)

	errs := make([]ErrorCount, 0, len(r.errors))
	for k, v := range r.errors {
		errs = append(errs, ErrorCount{Stage: k.Stage, Class: k.Class, Code: k.Code, Count: v})
	}
	slices.SortFunc(errs, func(a, b ErrorCount) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return cmp.Compare(a.Code, b.Code)
	})

	senders := make([]*SenderStat, 0, len(r.senders))
	for _, s := range r.senders {
		senders = append(senders, s)
	}
	slices.SortFunc(senders, func(a, b *SenderStat) int {
		return cmp.Compare(a.Label, b.Label)
	})

	elapsed := time.Duration(0)
	if !r.start.IsZero() {
		end := r.end
		if end.IsZero() {
			end = time.Now()
		}
		elapsed = end.Sub(r.start)
	}

	return Snapshot{
		Elapsed:  elapsed,
		Stages:   stages,
		Outcomes: outcomes,
		Errors:   errs,
		Senders:  senders,
	}
}
