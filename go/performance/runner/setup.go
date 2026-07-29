// Package runner drives a performance scenario against a running system.
package runner

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"

	"github.com/interledger/interledger-app/go/performance/auth"
	"github.com/interledger/interledger-app/go/performance/client"
	"github.com/interledger/interledger-app/go/performance/config"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
)

// receiverIdentityTypeWalletURL matches payments.IdentityTypeWalletURL. Receivers
// are addressed by wallet URL because that is the only form available without
// credentials for the receiving wallet.
const receiverIdentityTypeWalletURL = 3

// sender is a prepared source wallet, ready to issue payments.
type sender struct {
	cfg    config.Sender
	wallet *client.Wallet

	// linkedAccount is the account payments are drawn from.
	linkedAccount string
	// asset and assetScale describe the currency payments are denominated in.
	asset      string
	assetScale int32
	// startBalance is the opening balance in minor units, used to cap a drain run
	// and to report how much was actually moved.
	startBalance int64

	// receivers is this sender's target list, already resolved by the pairing rule.
	receivers []config.Receiver
	// cursor rotates through receivers for round-robin pairing.
	cursor int
}

// nextReceiver returns the next target for this sender.
func (s *sender) nextReceiver(pairing config.Pairing) config.Receiver {
	if pairing == config.PairingRandom {
		return s.receivers[rand.IntN(len(s.receivers))]
	}
	r := s.receivers[s.cursor%len(s.receivers)]
	s.cursor++
	return r
}

// maxPayments caps a drain run so a stuck balance lookup cannot turn into an
// unbounded run. Returns 0 when the run is not balance-bounded.
func (s *sender) maxPayments(cfg *config.Config) int {
	if cfg.Run.Stop != config.StopDrain || s.startBalance <= 0 {
		return 0
	}
	return int(s.startBalance / cfg.Run.Amount)
}

// setup authenticates every sender, resolves their spending accounts and checks
// every receiver is reachable.
//
// All of it happens before any load is generated: a scenario with one bad wallet
// should fail in a few seconds with a clear message, not halfway through a run as
// a column of Unauthenticated errors.
func setup(ctx context.Context, cfg *config.Config, pool *client.Pool, authClient *auth.Client) ([]*sender, error) {
	senders, err := prepareSenders(ctx, cfg, pool, authClient)
	if err != nil {
		return nil, err
	}

	if err := checkReceivers(ctx, cfg, senders[0]); err != nil {
		return nil, err
	}

	assignReceivers(cfg, senders)

	return senders, nil
}

func prepareSenders(ctx context.Context, cfg *config.Config, pool *client.Pool, authClient *auth.Client) ([]*sender, error) {
	type result struct {
		s   *sender
		err error
	}

	results := make([]result, len(cfg.Senders))
	var wg sync.WaitGroup

	// Logging in a hundred wallets serially is slow enough to be annoying, and
	// Kratos handles the concurrency fine.
	sem := make(chan struct{}, 16)
	for i, sc := range cfg.Senders {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			s, err := prepareSender(ctx, cfg, pool, authClient, sc)
			results[i] = result{s: s, err: err}
		})
	}
	wg.Wait()

	senders := make([]*sender, 0, len(results))
	var errs []error
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
			continue
		}
		senders = append(senders, r.s)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return senders, nil
}

func prepareSender(
	ctx context.Context,
	cfg *config.Config,
	pool *client.Pool,
	authClient *auth.Client,
	sc config.Sender,
) (*sender, error) {
	token := sc.SessionToken
	if token == "" {
		var err error
		token, err = authClient.Login(ctx, sc.Email, sc.Password)
		if err != nil {
			return nil, fmt.Errorf("sender %s: login: %w", sc.Label, err)
		}
	} else if err := authClient.Verify(ctx, token); err != nil {
		return nil, fmt.Errorf("sender %s: session_token rejected: %w", sc.Label, err)
	}

	w := pool.Wallet(sc.Label, token)

	if _, err := w.GetWalletInfo(ctx); err != nil {
		return nil, fmt.Errorf("sender %s: GetWalletInfo: %w", sc.Label, client.Classify("setup", err))
	}

	balances, err := w.GetBalances(ctx)
	if err != nil {
		return nil, fmt.Errorf("sender %s: GetBalances: %w", sc.Label, client.Classify("setup", err))
	}

	bal, err := selectBalance(balances, sc.LinkedAccount, cfg.Run.Asset)
	if err != nil {
		return nil, fmt.Errorf("sender %s: %w", sc.Label, err)
	}

	s := &sender{
		cfg:           sc,
		wallet:        w,
		linkedAccount: bal.GetLinkedAccount(),
		asset:         bal.GetBalance().GetAsset(),
		assetScale:    bal.GetBalance().GetAssetScale(),
		startBalance:  bal.GetBalance().GetAmount(),
	}

	// An explicit asset in the scenario wins, so a run can be pinned to a
	// currency rather than inheriting whatever account happened to be selected.
	if cfg.Run.Asset != "" {
		s.asset = cfg.Run.Asset
		s.assetScale = cfg.Run.AssetScale
	}

	if cfg.Run.Stop == config.StopDrain && s.startBalance < cfg.Run.Amount {
		return nil, fmt.Errorf(
			"sender %s: balance %d %s is below the %d per-payment amount, so a drain run has nothing to send",
			sc.Label, s.startBalance, s.asset, cfg.Run.Amount,
		)
	}

	return s, nil
}

// selectBalance picks the account to spend from: the pinned one if the scenario
// names it, otherwise the largest balance in the requested asset.
func selectBalance(balances []*pb.Balance, linkedAccount, asset string) (*pb.Balance, error) {
	if len(balances) == 0 {
		return nil, errors.New("wallet has no balances; it needs a funded linked account before it can send")
	}

	if linkedAccount != "" {
		for _, b := range balances {
			if b.GetLinkedAccount() == linkedAccount {
				return b, nil
			}
		}
		return nil, fmt.Errorf("linked_account %q not found among the wallet's %d balances", linkedAccount, len(balances))
	}

	var best *pb.Balance
	for _, b := range balances {
		if asset != "" && b.GetBalance().GetAsset() != asset {
			continue
		}
		if best == nil || b.GetBalance().GetAmount() > best.GetBalance().GetAmount() {
			best = b
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no balance in asset %q; set senders[].linked_account or clear run.asset", asset)
	}

	return best, nil
}

// checkReceivers resolves every receiver address once, using the first sender's
// credentials. GetPaymentAddress reports whether the pair can actually transact,
// which catches unapproved KYC and incompatible linked accounts up front — both
// of which CreatePayment would otherwise reject on every single attempt.
func checkReceivers(ctx context.Context, cfg *config.Config, probe *sender) error {
	var errs []error
	for _, r := range cfg.Receivers {
		resp, err := probe.wallet.GetPaymentAddress(ctx, r.WalletAddress)
		if err != nil {
			errs = append(errs, fmt.Errorf("receiver %s: %w", r.Label, client.Classify("setup", err)))
			continue
		}
		if !resp.GetCanSendToAddress() {
			errs = append(errs, fmt.Errorf(
				"receiver %s: %s cannot be paid from %s — check both wallets are KYC approved and have compatible linked accounts",
				r.Label, r.WalletAddress, probe.cfg.Label,
			))
		}
	}
	return errors.Join(errs...)
}

// assignReceivers applies the pairing rule, giving each sender its target list.
func assignReceivers(cfg *config.Config, senders []*sender) {
	for i, s := range senders {
		switch cfg.Run.Pairing {
		case config.PairingIndex:
			s.receivers = []config.Receiver{cfg.Receivers[i]}
		case config.PairingFanIn:
			s.receivers = []config.Receiver{cfg.Receivers[0]}
		default:
			// Round-robin and random both work over the whole list. Staggering the
			// starting offset stops every sender hammering receivers[0] at once.
			s.receivers = cfg.Receivers
			s.cursor = i % len(cfg.Receivers)
		}
	}
}
