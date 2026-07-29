// Package client provides authenticated backend gRPC clients for the
// performance harness.
//
// Connections are pooled rather than created per wallet: a single HTTP/2
// connection caps concurrent streams, so a few connections shared by many
// wallets keeps stream contention off the measurement while avoiding the cost of
// a connection per virtual user. Authentication is per-call metadata, so wallets
// can share a connection safely.
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync/atomic"
	"time"

	pb "github.com/interledger/interledger-app/go/proto/backend/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Options configures a Pool.
type Options struct {
	Address        string
	Connections    int
	TLS            bool
	TLSSkipVerify  bool
	DialTimeout    time.Duration
	RequestTimeout time.Duration
}

// Pool holds a fixed set of gRPC connections handed out round-robin.
type Pool struct {
	conns          []*grpc.ClientConn
	clients        []pb.BackendServiceClient
	next           atomic.Uint64
	requestTimeout time.Duration
}

// NewPool dials the target and returns a ready pool.
func NewPool(ctx context.Context, opts Options) (*Pool, error) {
	if opts.Connections < 1 {
		opts.Connections = 1
	}

	creds := insecure.NewCredentials()
	if opts.TLS {
		creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: opts.TLSSkipVerify}) //nolint:gosec // opt-in, for self-signed certs in front of a test environment
	}

	p := &Pool{
		conns:          make([]*grpc.ClientConn, 0, opts.Connections),
		clients:        make([]pb.BackendServiceClient, 0, opts.Connections),
		requestTimeout: opts.RequestTimeout,
	}

	for range opts.Connections {
		conn, err := grpc.NewClient(opts.Address, grpc.WithTransportCredentials(creds))
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("dial %s: %w", opts.Address, err)
		}
		p.conns = append(p.conns, conn)
		p.clients = append(p.clients, pb.NewBackendServiceClient(conn))
	}

	// grpc.NewClient connects lazily. Force the first connection now so a wrong
	// address or a dead port-forward fails during setup instead of showing up as
	// a run full of Unavailable errors.
	dialCtx, cancel := context.WithTimeout(ctx, opts.DialTimeout)
	defer cancel()
	if err := waitForReady(dialCtx, p.conns[0]); err != nil {
		p.Close()
		return nil, fmt.Errorf("connect to %s: %w", opts.Address, err)
	}

	return p, nil
}

func waitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	conn.Connect()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}
	}
}

// Close releases every connection in the pool.
func (p *Pool) Close() {
	for _, conn := range p.conns {
		_ = conn.Close()
	}
}

// RequestTimeout is the per-RPC deadline configured for this pool.
func (p *Pool) RequestTimeout() time.Duration {
	return p.requestTimeout
}

// Wallet is an authenticated view of the backend for one test wallet.
type Wallet struct {
	// Label identifies the wallet in logs and reports.
	Label string

	client pb.BackendServiceClient
	token  string
	pool   *Pool
}

// Wallet binds a session token to one connection from the pool. Assignment is
// round-robin, so wallets spread evenly across connections.
func (p *Pool) Wallet(label, token string) *Wallet {
	idx := p.next.Add(1) % uint64(len(p.clients))
	return &Wallet{
		Label:  label,
		client: p.clients[idx],
		token:  token,
		pool:   p,
	}
}

// ctx attaches the bearer token and the per-request deadline. The caller keeps
// ownership of cancellation.
func (w *Wallet) ctx(parent context.Context) (context.Context, context.CancelFunc) {
	authed := metadata.AppendToOutgoingContext(parent, "authorization", "Bearer "+w.token)
	if w.pool.requestTimeout <= 0 {
		return authed, func() {}
	}
	return context.WithTimeout(authed, w.pool.requestTimeout)
}

// GetWalletInfo returns the wallet's own details, used as a setup smoke test.
func (w *Wallet) GetWalletInfo(ctx context.Context) (*pb.WalletInfo, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.GetWalletInfo(callCtx, &pb.Empty{})
}

// GetBalances lists the wallet's per-linked-account balances.
func (w *Wallet) GetBalances(ctx context.Context) ([]*pb.Balance, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	resp, err := w.client.GetBalances(callCtx, &pb.Empty{})
	if err != nil {
		return nil, err
	}
	return resp.GetBalances(), nil
}

// CreatePayment starts a payment. This returns as soon as the backend has
// persisted the payment and started its Temporal workflow — it does not mean the
// payment has moved any money.
func (w *Wallet) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.Payment, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.CreatePayment(callCtx, req)
}

// UpdatePayment amends a created payment. Only used when the scenario asks to
// mirror protea's create-then-update-then-confirm sequence.
func (w *Wallet) UpdatePayment(ctx context.Context, req *pb.UpdatePaymentRequest) (*pb.Payment, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.UpdatePayment(callCtx, req)
}

// ConfirmPayment authorises the payment, which is the point the workflow is
// allowed to move funds.
func (w *Wallet) ConfirmPayment(ctx context.Context, id string) (*pb.Payment, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.ConfirmPayment(callCtx, &pb.ConfirmPaymentRequest{Id: id})
}

// GetPayment reads a payment's current state, used to poll for settlement.
func (w *Wallet) GetPayment(ctx context.Context, id string) (*pb.Payment, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.GetPayment(callCtx, &pb.GetPaymentRequest{Id: id})
}

// GetPaymentAddress resolves a wallet address, used to check a receiver is
// reachable and can be paid before the run starts.
func (w *Wallet) GetPaymentAddress(ctx context.Context, address string) (*pb.GetPaymentAddressResponse, error) {
	callCtx, cancel := w.ctx(ctx)
	defer cancel()
	return w.client.GetPaymentAddress(callCtx, &pb.GetPaymentAddressRequest{Address: address})
}
