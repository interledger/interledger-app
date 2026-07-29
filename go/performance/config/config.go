// Package config loads and validates performance-test scenario files.
//
// A scenario is one or more YAML files, deep-merged in order (later files win),
// parsed through configa so the same layering and {{ secret }} resolution used by
// the backend also works here. That lets a committed scenario file describe the
// shape of a run while a local, git-ignored overlay supplies wallet credentials.
package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/configa"
)

// StopMode determines when a sender stops issuing payments.
type StopMode string

const (
	// StopDrain sends until the source wallet is empty: the runner caps the run
	// at balance/amount payments and also stops early if the backend reports
	// insufficient funds. This is the default shape described in the README.
	StopDrain StopMode = "drain"
	// StopCount sends a fixed number of payments per sender.
	StopCount StopMode = "count"
	// StopDuration sends for a fixed wall-clock duration.
	StopDuration StopMode = "duration"
)

// Pairing determines how senders are matched to receivers.
type Pairing string

const (
	// PairingIndex maps senders[i] to receivers[i] — the 100-wallets-to-100-wallets
	// shape. Requires the two lists to be the same length.
	PairingIndex Pairing = "index"
	// PairingFanIn points every sender at receivers[0] — the 100-wallets-to-1-wallet
	// shape, which concentrates contention on a single receiving wallet.
	PairingFanIn Pairing = "fan_in"
	// PairingRoundRobin rotates each sender through the whole receiver list, so
	// consecutive payments from one sender land on different receivers.
	PairingRoundRobin Pairing = "round_robin"
	// PairingRandom picks a receiver uniformly at random for every payment.
	PairingRandom Pairing = "random"
)

// Config is a complete performance scenario.
type Config struct {
	Target     Target        `yaml:"target"`
	Run        Run           `yaml:"run"`
	Settlement Settlement    `yaml:"settlement"`
	Metrics    Metrics       `yaml:"metrics"`
	Senders    []Sender      `yaml:"senders"`
	Receivers  []Receiver    `yaml:"receivers"`
	Wallets    []WalletEntry `yaml:"wallets"`
}

// Target describes the system under test. The same struct covers a local Compose
// stack and a deployed environment reached through a port-forward or tunnel —
// only the addresses differ.
type Target struct {
	// GRPCAddress is the backend gRPC endpoint, e.g. "localhost:8443" locally or
	// the local end of a `kubectl port-forward` against a deployed backend.
	GRPCAddress string `yaml:"grpc_address" validate:"required"`
	// KratosURL is the public Kratos endpoint used to mint session tokens.
	KratosURL string `yaml:"kratos_url" validate:"required,url"`
	// TLS enables transport credentials. The backend serves plaintext gRPC (see
	// serveGrpc in backend/main.go), so this stays false for a direct dial and is
	// only needed when a proxy terminates TLS in front of it.
	TLS bool `yaml:"tls"`
	// TLSSkipVerify disables certificate verification when TLS is enabled.
	TLSSkipVerify bool `yaml:"tls_skip_verify"`
	// Connections is the size of the gRPC connection pool. A single HTTP/2
	// connection caps concurrent streams, so high arrival rates want several.
	Connections int `yaml:"connections" validate:"min=0"`
	// DialTimeout bounds initial connection establishment.
	DialTimeout time.Duration `yaml:"dial_timeout"`
	// RequestTimeout bounds each individual unary RPC.
	RequestTimeout time.Duration `yaml:"request_timeout"`
}

// Run describes the load profile.
type Run struct {
	Stop StopMode `yaml:"stop" validate:"omitempty,oneof=drain count duration"`
	// Duration is the wall-clock cap. Required for stop=duration, and honoured as
	// a safety ceiling for the other modes when non-zero.
	Duration time.Duration `yaml:"duration"`
	// CountPerSender is the number of payments each sender attempts when
	// stop=count. Attempts, not successes: a sender that is being rejected every
	// time still terminates rather than looping forever.
	CountPerSender int `yaml:"count_per_sender" validate:"min=0"`
	// ArrivalRate is the target payments per second across all senders combined.
	// Zero means unthrottled — send as fast as the system accepts.
	ArrivalRate float64 `yaml:"arrival_rate" validate:"min=0"`
	// Burst is the token-bucket burst for ArrivalRate. Defaults to one second of
	// arrivals so a brief stall does not permanently lose throughput.
	Burst int `yaml:"burst" validate:"min=0"`
	// MaxInFlight caps payments that have been confirmed but not yet reached a
	// terminal state. This is the backpressure that stops the harness from
	// queueing unbounded work into Temporal and reporting a meaningless
	// throughput number.
	MaxInFlight int `yaml:"max_in_flight" validate:"min=0"`
	// Amount is the value of each payment in minor units of the sender's asset.
	// Checked in validate rather than by a min tag: configa validates struct tags
	// during Resolve, which runs before defaults are applied, so a min=1 tag here
	// would reject every scenario that omits the field.
	Amount int64 `yaml:"amount"`
	// Asset and AssetScale override the sender's selected balance. Leave empty to
	// inherit them from the linked account the payments are drawn from.
	Asset      string `yaml:"asset"`
	AssetScale int32  `yaml:"asset_scale" validate:"min=0"`
	// Pairing selects the sender-to-receiver topology.
	Pairing Pairing `yaml:"pairing" validate:"omitempty,oneof=index fan_in round_robin random"`
	// IncludeUpdateStep inserts an UpdatePayment call between create and confirm,
	// mirroring what protea's send flow actually does. Off by default so the
	// measured path is the minimum the API allows.
	IncludeUpdateStep bool `yaml:"include_update_step"`
	// Note is attached to every payment, which makes the generated rows easy to
	// find and clean up afterwards.
	Note string `yaml:"note"`
	// Warmup discards results for this long after start, so connection setup and
	// JIT-ish first-call costs do not skew the percentiles.
	Warmup time.Duration `yaml:"warmup"`
}

// Settlement controls tracking of payments through to a terminal state.
//
// CreatePayment only starts a Temporal workflow, so the RPC latency says nothing
// about how long a payment takes to settle. When Track is on, the runner polls
// GetPayment until the payment reaches Completed or Failed and reports that as
// the end-to-end latency.
type Settlement struct {
	Track        bool          `yaml:"track"`
	PollInterval time.Duration `yaml:"poll_interval"`
	// Timeout is how long a single payment may stay non-terminal before the
	// runner gives up on it and records it as timed out.
	Timeout time.Duration `yaml:"timeout"`
	// Workers is the size of the polling pool.
	Workers int `yaml:"workers" validate:"min=0"`
}

// Metrics controls how results are reported.
type Metrics struct {
	// Console prints a summary table when the run finishes.
	Console bool `yaml:"console"`
	// Listen exposes /metrics for Prometheus, e.g. ":9464". Empty disables it.
	Listen string `yaml:"listen"`
	// JobLabel is attached to every exported series so several concurrent runs
	// can be told apart in Grafana.
	JobLabel string `yaml:"job_label"`
	// LingerAfterRun keeps the metrics endpoint up after the run finishes, giving
	// Prometheus time for a final scrape. Ignored when Listen is empty.
	LingerAfterRun time.Duration `yaml:"linger_after_run"`
}

// Sender is a wallet that payments are drawn from.
type Sender struct {
	// Label names the sender in logs and in the console report. Defaults to Email
	// or, failing that, the sender's index.
	Label string `yaml:"label"`
	// Email and Password authenticate against Kratos to mint a session token.
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
	// SessionToken skips the login round trip when a token is already available.
	// Takes precedence over Email/Password.
	SessionToken string `yaml:"session_token"`
	// LinkedAccount is the account to spend from. Empty selects the funded
	// account with the largest balance.
	LinkedAccount string `yaml:"linked_account"`
}

// WalletEntry models a wallet produced by the provisioner. When present, the
// loader expands each entry into a sender and receiver pair unless the scenario
// already declared explicit sender/receiver lists.
type WalletEntry struct {
	Label         string `yaml:"label"`
	Email         string `yaml:"email"`
	Password      string `yaml:"password"`
	WalletAddress string `yaml:"wallet_address" validate:"required"`
}

// Receiver is a wallet that payments are sent to. Receivers are addressed
// publicly, so no credentials are needed.
type Receiver struct {
	Label string `yaml:"label"`
	// WalletAddress is the receiver's wallet URL, e.g. https://ilp.link/alice.
	// Sent as receiverIdentity with receiverIdentityType 3 (WalletURL).
	WalletAddress string `yaml:"wallet_address" validate:"required"`
	// LinkedAccount optionally pins the receiving account.
	LinkedAccount string `yaml:"linked_account"`
}

// Default values applied to any field the scenario leaves unset.
const (
	DefaultConnections    = 4
	DefaultDialTimeout    = 15 * time.Second
	DefaultRequestTimeout = 30 * time.Second
	DefaultMaxInFlight    = 256
	DefaultAmount         = 1
	DefaultPollInterval   = 500 * time.Millisecond
	DefaultTimeout        = 2 * time.Minute
	DefaultWorkers        = 16
	DefaultJobLabel       = "interledger-perf"
	DefaultLinger         = 15 * time.Second
)

// Load reads the given YAML files, deep-merges them in order, applies defaults
// and validates the result.
func Load(ctx context.Context, files []string) (*Config, error) {
	if len(files) == 0 {
		return nil, errors.New("at least one scenario file is required")
	}

	parsed, err := configa.Parse[Config](files, configa.WithSecretClient(configa.NewInClusterSecretClient()))
	if err != nil {
		return nil, fmt.Errorf("parse scenario: %w", err)
	}

	cfg, err := parsed.Resolve(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve scenario: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Target.Connections == 0 {
		c.Target.Connections = DefaultConnections
	}
	if c.Target.DialTimeout == 0 {
		c.Target.DialTimeout = DefaultDialTimeout
	}
	if c.Target.RequestTimeout == 0 {
		c.Target.RequestTimeout = DefaultRequestTimeout
	}

	if c.Run.Stop == "" {
		c.Run.Stop = StopDrain
	}
	if c.Run.Pairing == "" {
		c.Run.Pairing = PairingRoundRobin
	}
	if c.Run.Amount == 0 {
		c.Run.Amount = DefaultAmount
	}
	if c.Run.MaxInFlight == 0 {
		c.Run.MaxInFlight = DefaultMaxInFlight
	}
	if c.Run.Burst == 0 && c.Run.ArrivalRate > 0 {
		// One second of arrivals, at least one token.
		c.Run.Burst = max(int(c.Run.ArrivalRate), 1)
	}

	if c.Settlement.PollInterval == 0 {
		c.Settlement.PollInterval = DefaultPollInterval
	}
	if c.Settlement.Timeout == 0 {
		c.Settlement.Timeout = DefaultTimeout
	}
	if c.Settlement.Workers == 0 {
		c.Settlement.Workers = DefaultWorkers
	}

	if c.Metrics.JobLabel == "" {
		c.Metrics.JobLabel = DefaultJobLabel
	}
	if c.Metrics.Listen != "" && c.Metrics.LingerAfterRun == 0 {
		c.Metrics.LingerAfterRun = DefaultLinger
	}

	if len(c.Wallets) > 0 && len(c.Senders) == 0 && len(c.Receivers) == 0 {
		for _, wallet := range c.Wallets {
			c.Senders = append(c.Senders, Sender{
				Label:    wallet.Label,
				Email:    wallet.Email,
				Password: wallet.Password,
			})
			c.Receivers = append(c.Receivers, Receiver{
				Label:         wallet.Label,
				WalletAddress: wallet.WalletAddress,
			})
		}
	}

	for i := range c.Senders {
		if c.Senders[i].Label == "" {
			if c.Senders[i].Email != "" {
				c.Senders[i].Label = c.Senders[i].Email
			} else {
				c.Senders[i].Label = fmt.Sprintf("sender-%03d", i+1)
			}
		}
	}
	for i := range c.Receivers {
		if c.Receivers[i].Label == "" {
			c.Receivers[i].Label = c.Receivers[i].WalletAddress
		}
	}
}

// validate enforces the rules that cannot be expressed as struct tags because
// they depend on other fields.
func (c *Config) validate() error {
	if len(c.Senders) == 0 {
		return errors.New("senders: at least one sender is required")
	}
	if len(c.Receivers) == 0 {
		return errors.New("receivers: at least one receiver is required")
	}

	for i, s := range c.Senders {
		if s.SessionToken == "" && (s.Email == "" || s.Password == "") {
			return fmt.Errorf("senders[%d] (%s): either session_token or both email and password are required", i, s.Label)
		}
	}

	if c.Run.Amount < 1 {
		return errors.New("run.amount must be at least 1 minor unit")
	}

	switch c.Run.Stop {
	case StopCount:
		if c.Run.CountPerSender < 1 {
			return errors.New("run.count_per_sender must be at least 1 when run.stop is count")
		}
	case StopDuration:
		if c.Run.Duration <= 0 {
			return errors.New("run.duration must be positive when run.stop is duration")
		}
	case StopDrain:
		// Bounded by wallet balance; run.duration is an optional safety ceiling.
	}

	if c.Run.Pairing == PairingIndex && len(c.Senders) != len(c.Receivers) {
		return fmt.Errorf(
			"run.pairing index requires equal list lengths, got %d senders and %d receivers",
			len(c.Senders), len(c.Receivers),
		)
	}

	if c.Run.Asset == "" && c.Run.AssetScale != 0 {
		return errors.New("run.asset is required when run.asset_scale is set")
	}

	return nil
}
