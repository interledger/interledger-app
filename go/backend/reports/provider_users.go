// Package reports builds downloadable reports for the admin portal.
package reports

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/kyc/persona"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/gatehub"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/log"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const walletConcurrency = 10

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Xago() xago.Client
	Gatehub() gatehub.Client
}

// providerReport describes what differs between the per-provider user reports.
// Everything else (columns, KYC, transactions, CSV layout) is shared.
type providerReport struct {
	// name prefixes the provider-specific columns, e.g. "xago" -> "xago_linked_account_id".
	name string
	// linkedAccountProvider and linkedAccountType identify the balance linked account.
	linkedAccountProvider string
	linkedAccountType     string
	// selectUsers picks the report's users and wallets. alwaysInclude lists users that get
	// a row even without a wallet.
	selectUsers func(ctx context.Context, b Backends, identities []user.User) (alwaysInclude []string, wallets []walletRow, err error)
	// balance returns the total and available balance of a balance linked account.
	balance func(ctx context.Context, b Backends, linkedAccountID string) (total, available currency.Amount, err error)
	// externalUserID, if set, adds a "<name>_user_id" column with the provider's user ID.
	externalUserID func(ctx context.Context, b Backends, walletID string) (string, error)
}

type walletRow struct {
	UserID   string `db:"user_id"`
	WalletID string `db:"wallet_id"`
	Country  string `db:"country"`
}

type reportUser struct {
	UserID    string
	Email     string
	FirstName string
	LastName  string
	// KratosCountry is the countryCode trait; empty if the identity is gone from Kratos.
	KratosCountry string
	Wallets       []reportWallet
}

type reportWallet struct {
	WalletID string
	Country  string
	// LocalKYC is the app's own KYC status for the wallet (wallet_kyc_status), e.g. "Approved".
	LocalKYC string
	// PersonaInquiryID is the wallet's latest approved Persona inquiry, or its most recent
	// inquiry if none is approved; empty if the wallet never started KYC.
	PersonaInquiryID    string
	PersonaInquiryState string
	// ExternalUserID is the provider's ID for the user; only set when the report has that column.
	ExternalUserID string
	// LinkedAccountID is empty when the wallet has no balance account with the provider (yet).
	LinkedAccountID  string
	Balance          string
	AvailableBalance string
	BalanceCurrency  string
	Error            string
	Transactions     []reportTransaction
}

type reportTransaction struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Type      string
	State     string
	Provider  string
	Amount    string
	Fee       string
	Currency  string
	Title     string
}

// build returns the report as CSV: one row per transaction, with the user's details
// and the wallet's current balance repeated on each row. A wallet that fails to load
// gets a row with the error instead of failing the whole report.
func (p providerReport) build(ctx context.Context, b Backends) (string, error) {
	users, err := p.listUsers(ctx, b)
	if err != nil {
		return "", err
	}

	sem := make(chan struct{}, walletConcurrency)
	var wg sync.WaitGroup
	for i := range users {
		for j := range users[i].Wallets {
			w := &users[i].Wallets[j]
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				if err := p.fillWallet(ctx, b, w); err != nil {
					log.Warn("failed to build report for wallet", zap.String("report", p.name), zap.String("wallet_id", w.WalletID), zap.Error(err))
					w.Error = err.Error()
				}
			}()
		}
	}
	wg.Wait()

	return p.csv(users)
}

func (p providerReport) header() []string {
	h := []string{
		"email", "first_name", "last_name", "user_id", "kratos_country",
		"wallet_id", "wallet_country", "local_kyc", "persona_inquiry_id", "persona_inquiry_state",
	}
	if p.externalUserID != nil {
		h = append(h, p.name+"_user_id")
	}
	return append(h,
		p.name+"_linked_account_id", "balance", "available_balance", "balance_currency",
		"transaction_id", "created_at", "updated_at", "type", "state", "provider", "amount", "fee", "currency", "title",
		"error",
	)
}

// csv flattens the report to one row per transaction. Wallets without transactions
// and users without wallets still get a row.
func (p providerReport) csv(users []reportUser) (string, error) {
	header := p.header()
	var sb strings.Builder
	cw := csv.NewWriter(&sb)
	if err := cw.Write(header); err != nil {
		return "", err
	}

	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.UTC().Format(time.RFC3339)
	}

	for _, u := range users {
		userCols := []string{u.Email, u.FirstName, u.LastName, u.UserID, u.KratosCountry}
		if len(u.Wallets) == 0 {
			if err := cw.Write(append(userCols, make([]string, len(header)-len(userCols))...)); err != nil {
				return "", err
			}
			continue
		}

		for _, w := range u.Wallets {
			walletCols := []string{w.WalletID, w.Country, w.LocalKYC, w.PersonaInquiryID, w.PersonaInquiryState}
			if p.externalUserID != nil {
				walletCols = append(walletCols, w.ExternalUserID)
			}
			walletCols = append(walletCols, w.LinkedAccountID, w.Balance, w.AvailableBalance, w.BalanceCurrency)

			txs := w.Transactions
			if len(txs) == 0 {
				txs = []reportTransaction{{}}
			}
			for _, t := range txs {
				row := append(append(append([]string{}, userCols...), walletCols...),
					t.ID, formatTime(t.CreatedAt), formatTime(t.UpdatedAt), t.Type, t.State, t.Provider, t.Amount, t.Fee, t.Currency, t.Title,
					w.Error)
				if err := cw.Write(row); err != nil {
					return "", err
				}
			}
		}
	}

	cw.Flush()
	if err := cw.Error(); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// listUsers returns the report's users with their wallets' IDs and countries filled in,
// sorted by email.
func (p providerReport) listUsers(ctx context.Context, b Backends) ([]reportUser, error) {
	ids, err := b.Users().ListIdentities(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing identities: %w", err)
	}

	alwaysInclude, rows, err := p.selectUsers(ctx, b, ids)
	if err != nil {
		return nil, err
	}

	identities := make(map[string]reportUser, len(ids))
	for _, id := range ids {
		identities[id.ID] = reportUser{
			UserID:        id.ID,
			Email:         strings.TrimSpace(id.Email),
			FirstName:     strings.TrimSpace(id.FirstName),
			LastName:      strings.TrimSpace(id.LastName),
			KratosCountry: id.Country.String(),
		}
	}

	users := map[string]*reportUser{}
	addUser := func(userID string) *reportUser {
		if u, ok := users[userID]; ok {
			return u
		}
		u, ok := identities[userID]
		if !ok {
			u = reportUser{UserID: userID}
		}
		users[userID] = &u
		return &u
	}
	for _, id := range alwaysInclude {
		addUser(id)
	}
	for _, r := range rows {
		u := addUser(r.UserID)
		u.Wallets = append(u.Wallets, reportWallet{WalletID: r.WalletID, Country: r.Country})
	}

	result := make([]reportUser, 0, len(users))
	for _, u := range users {
		result = append(result, *u)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Email) < strings.ToLower(result[j].Email)
	})

	log.Info("listed report users", zap.String("report", p.name), zap.Int("users", len(result)), zap.Int("wallets", len(rows)))
	return result, nil
}

// fillWallet adds a wallet's KYC status, Persona inquiry, provider details, balance
// and full transaction history to w.
func (p providerReport) fillWallet(ctx context.Context, b Backends, w *reportWallet) error {
	linkedAccounts, err := b.LinkedAccounts().ListByWalletId(ctx, w.WalletID)
	if err != nil {
		return fmt.Errorf("listing linked accounts: %w", err)
	}
	for _, la := range linkedAccounts {
		if la.Provider == p.linkedAccountProvider && la.Type == p.linkedAccountType && !la.DeletedAt.Valid {
			w.LinkedAccountID = la.ID
			break
		}
	}

	kycStatus, err := b.KYC().GetKYCStatus(ctx, w.WalletID)
	if err != nil {
		return fmt.Errorf("getting local kyc status: %w", err)
	}
	w.LocalKYC = kycStatus.String()

	var inquiry struct {
		ExternalID string `db:"external_id"`
		State      string `db:"state"`
	}
	err = b.DB().GetContext(ctx, &inquiry, `
		SELECT external_id, state
		FROM kyc_persona_inquiries
		WHERE wallet_id = $1
		ORDER BY (state = $2) DESC, updated_at DESC
		LIMIT 1`,
		w.WalletID, persona.InquiryApproved)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("getting persona inquiry: %w", err)
	}
	w.PersonaInquiryID = inquiry.ExternalID
	w.PersonaInquiryState = inquiry.State

	if p.externalUserID != nil {
		w.ExternalUserID, err = p.externalUserID(ctx, b, w.WalletID)
		if err != nil {
			return fmt.Errorf("getting %s user id: %w", p.name, err)
		}
	}

	if w.LinkedAccountID != "" {
		total, available, err := p.balance(ctx, b, w.LinkedAccountID)
		if err != nil {
			return fmt.Errorf("getting %s balance for linked account %s: %w", p.name, w.LinkedAccountID, err)
		}
		w.Balance = total.FormatAmount()
		w.AvailableBalance = available.FormatAmount()
		w.BalanceCurrency = total.Currency.String()
	}

	var txs []struct {
		ID          string         `db:"id"`
		Type        string         `db:"type"`
		State       string         `db:"state"`
		Provider    string         `db:"provider"`
		Title       sql.NullString `db:"title"`
		Amount      int64          `db:"amount"`
		ProviderFee int64          `db:"provider_fee"`
		Scale       int            `db:"asset_scale"`
		Asset       string         `db:"asset_code"`
		CreatedAt   time.Time      `db:"created_at"`
		UpdatedAt   time.Time      `db:"updated_at"`
	}
	// Queried directly because the transactions client doesn't expose created_at.
	err = b.DB().SelectContext(ctx, &txs, `
		SELECT id, type, state, provider, title, amount, provider_fee, asset_scale, asset_code, created_at, updated_at
		FROM transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC, id`,
		w.WalletID)
	if err != nil {
		return fmt.Errorf("listing transactions: %w", err)
	}

	w.Transactions = make([]reportTransaction, 0, len(txs))
	for _, t := range txs {
		amount := currency.Amount{Value: t.Amount, Currency: currency.ParseCurrency(t.Asset), Scale: t.Scale}
		fee := currency.Amount{Value: t.ProviderFee, Currency: amount.Currency, Scale: t.Scale}
		w.Transactions = append(w.Transactions, reportTransaction{
			ID:        t.ID,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
			Type:      t.Type,
			State:     t.State,
			Provider:  t.Provider,
			Amount:    amount.FormatAmount(),
			Fee:       fee.FormatAmount(),
			Currency:  amount.Currency.String(),
			Title:     t.Title.String,
		})
	}
	return nil
}
