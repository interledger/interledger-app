package engine

import (
	"fmt"
	"math/rand"
	"sync"
)

// Rate is an integer rational representation of a forex quote, read as
// 1 unit of Base = Num/Den units of Quote.
type Rate struct {
	Num int64
	Den int64
}

// String renders the rate as "num/den" for logging / metadata.
func (r Rate) String() string { return fmt.Sprintf("%d/%d", r.Num, r.Den) }

// DirectionSource chooses the post-conversion mutation direction. NextUp()
// returns true to apply +5% and false to apply -5%. Tests inject a
// deterministic implementation; production uses randomness.
type DirectionSource interface {
	NextUp() bool
}

// RandDirection is a DirectionSource backed by math/rand. Not safe for
// concurrent use without synchronization (FXService holds a mutex, so calls
// made via FXService are safe).
type RandDirection struct {
	r *rand.Rand
}

// NewRandDirection seeds a RandDirection with the given value.
func NewRandDirection(seed int64) *RandDirection {
	return &RandDirection{r: rand.New(rand.NewSource(seed))}
}

// NextUp returns true roughly half of the time.
func (d *RandDirection) NextUp() bool { return d.r.Intn(2) == 0 }

// ScriptedDirection replays a fixed sequence of directions, looping forever.
// Useful for deterministic tests.
type ScriptedDirection struct {
	steps []bool
	i     int
}

// NewScriptedDirection constructs a ScriptedDirection from the given sequence.
func NewScriptedDirection(steps ...bool) *ScriptedDirection {
	return &ScriptedDirection{steps: steps}
}

// NextUp returns the next scripted direction, wrapping at the end of the list.
func (d *ScriptedDirection) NextUp() bool {
	if len(d.steps) == 0 {
		return true
	}
	v := d.steps[d.i%len(d.steps)]
	d.i++
	return v
}

// FXService holds the simulator's current forex rates and mutates them after
// each successful conversion. It is safe for concurrent use.
//
// Rates are stored keyed by (base, quote). Rate(a, b) also consults the
// inverse (b, a) pair, returning the reciprocal when needed so the caller
// does not have to register both directions.
type FXService struct {
	mu    sync.Mutex
	rates map[[2]string]Rate
	dir   DirectionSource
}

// NewFXService constructs an empty FXService. Pass nil to use the default
// random source seeded from the current time.
func NewFXService(dir DirectionSource) *FXService {
	if dir == nil {
		dir = NewRandDirection(defaultSeed())
	}
	return &FXService{
		rates: make(map[[2]string]Rate),
		dir:   dir,
	}
}

// Set records an initial rate for the given base→quote pair.
func (f *FXService) Set(base, quote string, num, den int64) error {
	if base == "" || quote == "" {
		return fmt.Errorf("fx: base and quote must be non-empty")
	}
	if num <= 0 || den <= 0 {
		return fmt.Errorf("fx: rate must be positive, got %d/%d", num, den)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rates[[2]string{base, quote}] = reduce(Rate{Num: num, Den: den})
	return nil
}

// Rate returns the current rate for base→quote. If only the inverse pair was
// registered, the reciprocal is returned.
func (f *FXService) Rate(base, quote string) (Rate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.rates[[2]string{base, quote}]; ok {
		return r, nil
	}
	if r, ok := f.rates[[2]string{quote, base}]; ok {
		// Invert.
		return Rate{Num: r.Den, Den: r.Num}, nil
	}
	return Rate{}, fmt.Errorf("fx: no rate for %s/%s", base, quote)
}

// Mutate applies one +5% or -5% step to the base→quote rate using the
// DirectionSource and returns the new rate. The mutation is recorded on the
// canonical pair (whichever direction was originally Set); if the caller
// asked via the inverse, we still mutate the stored canonical entry.
func (f *FXService) Mutate(base, quote string) (Rate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := [2]string{base, quote}
	if _, ok := f.rates[key]; !ok {
		if _, inv := f.rates[[2]string{quote, base}]; inv {
			key = [2]string{quote, base}
		} else {
			return Rate{}, fmt.Errorf("fx: no rate for %s/%s", base, quote)
		}
	}
	var mult Rate
	if f.dir.NextUp() {
		mult = Rate{Num: 21, Den: 20} // +5 %
	} else {
		mult = Rate{Num: 19, Den: 20} // -5 %
	}
	current := f.rates[key]
	next := reduce(Rate{
		Num: current.Num * mult.Num,
		Den: current.Den * mult.Den,
	})
	f.rates[key] = next
	return next, nil
}

// Snapshot returns a copy of all stored rates for observability.
func (f *FXService) Snapshot() map[[2]string]Rate {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[[2]string]Rate, len(f.rates))
	for k, v := range f.rates {
		out[k] = v
	}
	return out
}

// reduce divides num and den by their greatest common divisor to keep
// long mutation chains from overflowing int64.
func reduce(r Rate) Rate {
	if r.Num == 0 || r.Den == 0 {
		return r
	}
	g := gcd(absInt64(r.Num), absInt64(r.Den))
	return Rate{Num: r.Num / g, Den: r.Den / g}
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}

func absInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// defaultSeed is overridable in tests if needed.
var defaultSeed = func() int64 { return now().UnixNano() }
