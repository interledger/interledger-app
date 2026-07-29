// Package provision bootstraps performance-test wallets against a local
// development environment.
//
// This is a convenience for local work only, and it is deliberately separate
// from the runner: a load test should never create accounts in an environment it
// does not own. Against anything other than local Compose, list wallets in the
// scenario file instead.
//
// The flow mirrors what protea does during signup — SetSignupUserData, a native
// Kratos registration, CompleteSignup, CreateUserDefaultWallet,
// CreateWalletAddress — and then does the two things the product deliberately
// does not expose over gRPC:
//
//   - KYC approval, written straight into wallet_kyc_status. There is no RPC for
//     this because approval belongs to the KYC provider; the same direct-write
//     shortcut already exists for email verification in local/scripts.
//   - Funding, via the Xago test deposit RPC, which the backend itself refuses to
//     serve outside a non-prod environment.
package provision

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/interledger/interledger-app/go/backend/errcodes"
	"github.com/interledger/interledger-app/go/performance/client"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"

	_ "github.com/lib/pq" // registers the "postgres" driver, matching the backend
	kratos "github.com/ory/kratos-client-go"
)

// kycStatusApproved matches kyc.StatusApproved.
const kycStatusApproved = 3

// Options configures a provisioning run.
type Options struct {
	// Countries is the list of country codes to provision, e.g. []string{"za", "de", "us"}.
	Countries []string
	// PerCountry is how many wallets to create per country.
	PerCountry int
	// TargetMajor is the target balance in major units for each wallet.
	TargetMajor int64
	// Prefix seeds the generated emails, phone numbers and wallet addresses, e.g.
	// "perf" yields perf-001@… and https://local.ilp.link/perf-001.
	Prefix string
	// Password is set on every created identity.
	Password string
	// PhonePrefix is the E.164 prefix that generated numbers are built on.
	PhonePrefix string
	// AddressHost is the wallet address host, e.g. "https://local.ilp.link".
	AddressHost string

	// GRPCAddress and KratosURL point at the local backend and Kratos.
	GRPCAddress string
	KratosURL   string
	// BackendDSN is the backend Postgres DSN, needed for the KYC approval write.
	// Empty skips KYC approval, which leaves the wallets unable to transact.
	BackendDSN string
}

// Wallet is a provisioned wallet, ready to be written into a scenario file.
type Wallet struct {
	Label         string
	Email         string
	Password      string
	UserID        string
	WalletAddress string
	SessionToken  string
	// Balance is the wallet's funded balance in minor units, zero when unfunded.
	Balance int64
	// Notes records steps that were skipped or failed for this wallet.
	Notes []string
}

// Run provisions Options.Count wallets, writing progress to out.
//
// Wallets are created one at a time. Signup touches Kratos, Postgres, Rafiki and
// a provider mock, so running it concurrently mostly produces confusing failures
// for no real time saving.
func Run(ctx context.Context, opts Options, out io.Writer) ([]Wallet, error) {
	if len(opts.Countries) == 0 {
		return nil, errors.New("at least one country is required")
	}
	if opts.PerCountry < 1 {
		return nil, errors.New("per-country must be at least 1")
	}
	applyDefaults(&opts)

	pool, err := client.NewPool(ctx, client.Options{
		Address:        opts.GRPCAddress,
		Connections:    1,
		DialTimeout:    15 * time.Second,
		RequestTimeout: 60 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	defer pool.Close()

	kratosClient := newKratosClient(opts.KratosURL)

	var db *sql.DB
	if opts.BackendDSN != "" {
		db, err = sql.Open("postgres", opts.BackendDSN)
		if err != nil {
			return nil, fmt.Errorf("open backend database: %w", err)
		}
		defer db.Close()

		if err := db.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("reach backend database at the configured DSN: %w", err)
		}
	} else {
		fmt.Fprintln(out, "warning: no backend DSN provided, so KYC approval is skipped — the wallets will not be able to transact")
	}

	countries, err := parseCountries(opts.Countries)
	if err != nil {
		return nil, err
	}

	wallets := make([]Wallet, 0, len(countries)*opts.PerCountry)
	globalIndex := 0
	for _, countrySpec := range countries {
		for i := 1; i <= opts.PerCountry; i++ {
			globalIndex++
			w, err := provisionOne(ctx, opts, pool, kratosClient, db, countrySpec, i, globalIndex)
			if err != nil {
				fmt.Fprintf(out, "  %s: FAILED: %v\n", walletLabel(opts.Prefix, countrySpec.code, i), err)
				return wallets, fmt.Errorf("provision %s: %w", walletLabel(opts.Prefix, countrySpec.code, i), err)
			}
			fmt.Fprintf(out, "  %s: %s (balance %d)\n", w.Label, w.WalletAddress, w.Balance)
			for _, note := range w.Notes {
				fmt.Fprintf(out, "      note: %s\n", note)
			}
			wallets = append(wallets, w)
		}
	}

	return wallets, nil
}

func applyDefaults(opts *Options) {
	if opts.Prefix == "" {
		opts.Prefix = "perf"
	}
	if opts.Password == "" {
		// Long enough for Kratos's default password policy.
		opts.Password = "PerfTestPassword123!"
	}
	if opts.PhonePrefix == "" {
		opts.PhonePrefix = "+2782"
	}
	if opts.AddressHost == "" {
		opts.AddressHost = "https://local.ilp.link"
	}
	if opts.TargetMajor <= 0 {
		opts.TargetMajor = 5000
	}
	if opts.PerCountry < 1 {
		opts.PerCountry = 100
	}
}

func newKratosClient(url string) *kratos.APIClient {
	cfg := kratos.NewConfiguration()
	cfg.Servers = kratos.ServerConfigurations{{URL: url, Description: "Public Kratos"}}
	return kratos.NewAPIClient(cfg)
}

func provisionOne(
	ctx context.Context,
	opts Options,
	pool *client.Pool,
	kratosClient *kratos.APIClient,
	db *sql.DB,
	spec countrySpec,
	i int,
	globalIndex int,
) (Wallet, error) {
	lbl := walletLabel(opts.Prefix, spec.code, i)
	w := Wallet{
		Label:         lbl,
		Email:         fmt.Sprintf("%s@perf.interledger.test", lbl),
		Password:      opts.Password,
		WalletAddress: walletAddress(opts.AddressHost, lbl),
	}
	phone := phoneNumber(opts.PhonePrefix, globalIndex)

	// 1. Register the signup with the backend, which validates the details and
	// reserves the email and phone number.
	anon := pool.Wallet(lbl, "")
	signupID, err := setSignupUserData(ctx, anon, spec, w.Email, phone, lbl)
	if err != nil {
		return w, err
	}

	// 2. Create the Kratos identity. The registration flow's session hook returns
	// a session token, so no separate login is needed.
	userID, token, err := register(ctx, kratosClient, w.Email, opts.Password, phone, lbl, spec.country)
	if err != nil {
		return w, err
	}
	w.UserID = userID
	w.SessionToken = token

	authed := pool.Wallet(lbl, token)

	// 3. Link the Kratos identity to the signup and sign the signup agreements.
	if err := completeSignup(ctx, authed, signupID, userID); err != nil {
		return w, err
	}

	// 4. Create the platform wallet, then its wallet address (which also registers
	// a payment pointer with Rafiki).
	if err := createWallet(ctx, authed, userID); err != nil {
		return w, err
	}
	if err := createAddress(ctx, authed, spec, w.WalletAddress, lbl); err != nil {
		return w, err
	}

	// 5. Approve KYC. Without this, CreatePayment rejects the wallet as both
	// sender and receiver.
	if db == nil {
		w.Notes = append(w.Notes, "KYC not approved (no backend DSN) — this wallet cannot send or receive")
	} else if err := approveKYC(ctx, db, userID); err != nil {
		return w, fmt.Errorf("approve KYC: %w", err)
	}

	// 6. Fund it.
	balance, notes, err := fund(ctx, authed, spec, opts.TargetMajor)
	if err != nil {
		w.Notes = append(w.Notes, err.Error())
	} else {
		w.Balance = balance
		w.Notes = append(w.Notes, notes...)
	}

	return w, nil
}

func setSignupUserData(ctx context.Context, anon *client.Wallet, spec countrySpec, email, phone, lbl string) (string, error) {
	resp, err := anon.SetSignupUserData(ctx, &pb.SetSignupUserDataRequest{
		FirstName:   "Perf",
		LastName:    strings.ToUpper(lbl),
		Email:       email,
		CountryCode: spec.country,
		Mobile:      phone,
	})
	if err != nil {
		return "", fmt.Errorf("SetSignupUserData: %w", client.Classify("signup", err))
	}
	return resp.GetId(), nil
}

func register(
	ctx context.Context,
	kratosClient *kratos.APIClient,
	email, password, phone, lbl, countryCode string,
) (userID, token string, err error) {
	flow, _, err := kratosClient.FrontendAPI.CreateNativeRegistrationFlow(ctx).Execute()
	if err != nil {
		return "", "", fmt.Errorf("create registration flow: %w", err)
	}

	traits := map[string]any{
		"email":       email,
		"phone":       phone,
		"firstName":   "Perf",
		"lastName":    strings.ToUpper(lbl),
		"countryCode": countryCode,
		// Skips the OTP round trip. Locally Twilio is disabled and any OTP is
		// accepted anyway, so this only saves calls.
		"phoneVerified": true,
	}

	body := kratos.UpdateRegistrationFlowWithPasswordMethodAsUpdateRegistrationFlowBody(
		kratos.NewUpdateRegistrationFlowWithPasswordMethod("password", password, traits),
	)

	reg, _, err := kratosClient.FrontendAPI.UpdateRegistrationFlow(ctx).
		Flow(flow.Id).
		UpdateRegistrationFlowBody(body).
		Execute()
	if err != nil {
		return "", "", fmt.Errorf("submit registration: %w", err)
	}

	token = reg.GetSessionToken()
	if token == "" {
		return "", "", errors.New("kratos returned no session token; the registration session hook may be disabled")
	}

	return reg.Identity.Id, token, nil
}

func completeSignup(ctx context.Context, w *client.Wallet, signupID, userID string) error {
	if err := w.CompleteSignup(ctx, signupID, userID); err != nil {
		return fmt.Errorf("CompleteSignup: %w", client.Classify("signup", err))
	}
	return nil
}

func createWallet(ctx context.Context, w *client.Wallet, userID string) error {
	err := w.CreateUserDefaultWallet(ctx, userID)
	if err == nil {
		return nil
	}

	// The wallets gRPC middleware auto-creates the user's default wallet — using
	// the country from their profile (ZA here) — on the first authenticated call,
	// which already happened during CompleteSignup above. This RPC carries no
	// country, so the backend defaults it to US and reports a conflict against
	// that existing, correctly-countried wallet. That specific conflict means
	// "the default wallet already exists", which is exactly the state we want, so
	// treat it as success. Any other failure is real.
	f := client.Classify("signup", err)
	if f.AppCode == errcodes.ErrCodeWalletsWalletConflict {
		return nil
	}
	return fmt.Errorf("CreateUserDefaultWallet: %w", f)
}

func createAddress(ctx context.Context, w *client.Wallet, spec countrySpec, address, alias string) error {
	err := w.CreateWalletAddress(ctx, &pb.CreateWalletAddressRequest{
		Url:        address,
		Asset:      spec.asset,
		AssetScale: spec.scale,
		Alias:      alias,
	})
	if err != nil {
		return fmt.Errorf("CreateWalletAddress %s: %w", address, client.Classify("signup", err))
	}
	return nil
}

// approveKYC marks the user's wallet KYC approved.
//
// There is no RPC for this — approval is the KYC provider's decision, and locally
// that means driving Persona or a provider mock through the UI. For a load test
// the approval itself is not what is being measured, so the row is written
// directly. The wallet is reached through user_wallets so the caller does not
// need the wallet ID; wallets itself has no user column.
func approveKYC(ctx context.Context, db *sql.DB, userID string) error {
	const q = `
INSERT INTO wallet_kyc_status (wallet_id, status)
SELECT wallet_id, $2 FROM user_wallets WHERE user_id = $1
ON CONFLICT (wallet_id) DO UPDATE SET status = $2, updated_at = now()::TIMESTAMP`

	res, err := db.ExecContext(ctx, q, userID, kycStatusApproved)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no wallet found for user %s", userID)
	}

	return nil
}

func waitForBalance(ctx context.Context, w *client.Wallet, linkedAccount string) (int64, error) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		balances, err := w.GetBalances(ctx)
		if err == nil {
			for _, b := range balances {
				if b.GetLinkedAccount() == linkedAccount && b.GetBalance().GetAmount() > 0 {
					return b.GetBalance().GetAmount(), nil
				}
			}
		} else {
			return 0, err
		}

		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
	return 0, nil
}
