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

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/kyc"
	"github.com/interledger/interledger-app/go/backend/kyc/persona"
	"github.com/interledger/interledger-app/go/backend/linkedaccounts"
	"github.com/interledger/interledger-app/go/backend/providers/xago"
	"github.com/interledger/interledger-app/go/backend/user"
	"github.com/interledger/interledger-app/go/log"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const xagoUsersConcurrency = 10

type Backends interface {
	DB() *sqlx.DB
	Users() user.Client
	KYC() kyc.Client
	LinkedAccounts() linkedaccounts.Client
	Xago() xago.Client
}

var xagoUsersCSVHeader = []string{
	"email", "first_name", "last_name", "user_id", "kratos_country",
	"wallet_id", "wallet_country", "local_kyc", "persona_inquiry_id", "persona_inquiry_state", "xago_linked_account_id", "balance", "available_balance", "balance_currency",
	"transaction_id", "created_at", "updated_at", "type", "state", "provider", "amount", "fee", "currency", "title",
	"error",
}

type xagoReportUser struct {
	UserID    string
	Email     string
	FirstName string
	LastName  string
	// KratosCountry is the countryCode trait; empty if the identity is gone from Kratos.
	KratosCountry string
	Wallets       []xagoReportWallet
}

type xagoReportWallet struct {
	WalletID string
	Country  string
	// LocalKYC is the app's own KYC status for the wallet (wallet_kyc_status), e.g. "Approved".
	LocalKYC string
	// PersonaInquiryID is the wallet's latest approved Persona inquiry, or its most recent
	// inquiry if none is approved; empty if the wallet never started KYC.
	PersonaInquiryID    string
	PersonaInquiryState string
	// XagoLinkedAccountID is empty when the wallet has no Xago balance account (yet).
	XagoLinkedAccountID string
	Balance             string
	AvailableBalance    string
	BalanceCurrency     string
	Error               string
	Transactions        []xagoReportTransaction
}

type xagoReportTransaction struct {
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

// XagoUsersCSV returns a CSV of every user that is ZA in Kratos (the only country that
// gets a Xago balance account) or owns a ZA wallet: one row per transaction, with the
// user's details and the wallet's current Xago balance repeated on each row. A wallet
// that fails to load gets a row with the error instead of failing the whole report.
func XagoUsersCSV(ctx context.Context, b Backends) (string, error) {
	users, err := listXagoReportUsers(ctx, b)
	if err != nil {
		return "", err
	}

	sem := make(chan struct{}, xagoUsersConcurrency)
	var wg sync.WaitGroup
	for i := range users {
		for j := range users[i].Wallets {
			w := &users[i].Wallets[j]
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				if err := fillXagoReportWallet(ctx, b, w); err != nil {
					log.Warn("failed to build xago report for wallet", zap.String("wallet_id", w.WalletID), zap.Error(err))
					w.Error = err.Error()
				}
			}()
		}
	}
	wg.Wait()

	return xagoUsersCSV(users)
}

// xagoUsersCSV flattens the report to one row per transaction. Wallets without
// transactions and users without wallets still get a row.
func xagoUsersCSV(users []xagoReportUser) (string, error) {
	var sb strings.Builder
	cw := csv.NewWriter(&sb)
	if err := cw.Write(xagoUsersCSVHeader); err != nil {
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
			if err := cw.Write(append(userCols, make([]string, len(xagoUsersCSVHeader)-len(userCols))...)); err != nil {
				return "", err
			}
			continue
		}

		for _, w := range u.Wallets {
			walletCols := []string{w.WalletID, w.Country, w.LocalKYC, w.PersonaInquiryID, w.PersonaInquiryState, w.XagoLinkedAccountID, w.Balance, w.AvailableBalance, w.BalanceCurrency}
			txs := w.Transactions
			if len(txs) == 0 {
				txs = []xagoReportTransaction{{}}
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

// listXagoReportUsers returns the users that are ZA in Kratos or own a ZA wallet,
// with their wallets' IDs and countries filled in, sorted by email.
func listXagoReportUsers(ctx context.Context, b Backends) ([]xagoReportUser, error) {
	ids, err := b.Users().ListIdentities(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing identities: %w", err)
	}

	// Every identity is kept: a ZA wallet owner may have another country in Kratos.
	identities := make(map[string]xagoReportUser, len(ids))
	var kratosZAUserIDs []string
	for _, id := range ids {
		identities[id.ID] = xagoReportUser{
			UserID:        id.ID,
			Email:         strings.TrimSpace(id.Email),
			FirstName:     strings.TrimSpace(id.FirstName),
			LastName:      strings.TrimSpace(id.LastName),
			KratosCountry: id.Country.String(),
		}
		if id.Country == country.ZA {
			kratosZAUserIDs = append(kratosZAUserIDs, id.ID)
		}
	}

	var rows []struct {
		UserID   string `db:"user_id"`
		WalletID string `db:"wallet_id"`
		Country  string `db:"country"`
	}
	err = b.DB().SelectContext(ctx, &rows, `
		SELECT uw.user_id, w.id AS wallet_id, w.country
		FROM user_wallets uw
		JOIN wallets w ON w.id = uw.wallet_id
		WHERE uw.user_id IN (SELECT user_id FROM user_wallets uw2 JOIN wallets w2 ON w2.id = uw2.wallet_id WHERE w2.country = $1)
		   OR uw.user_id = ANY($2)
		ORDER BY uw.user_id, w.created_at`,
		country.ZA, pq.Array(kratosZAUserIDs))
	if err != nil {
		return nil, fmt.Errorf("listing ZA wallets: %w", err)
	}

	users := map[string]*xagoReportUser{}
	addUser := func(userID string) *xagoReportUser {
		if u, ok := users[userID]; ok {
			return u
		}
		u, ok := identities[userID]
		if !ok {
			u = xagoReportUser{UserID: userID}
		}
		users[userID] = &u
		return &u
	}
	// ZA users in Kratos are reported even without a wallet.
	for _, id := range kratosZAUserIDs {
		addUser(id)
	}
	for _, r := range rows {
		u := addUser(r.UserID)
		u.Wallets = append(u.Wallets, xagoReportWallet{WalletID: r.WalletID, Country: r.Country})
	}

	result := make([]xagoReportUser, 0, len(users))
	for _, u := range users {
		result = append(result, *u)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Email) < strings.ToLower(result[j].Email)
	})

	log.Info("listed xago report users", zap.Int("users", len(result)), zap.Int("wallets", len(rows)))
	return result, nil
}

// fillXagoReportWallet adds a wallet's Persona inquiry, Xago balance and full
// transaction history to w.
func fillXagoReportWallet(ctx context.Context, b Backends, w *xagoReportWallet) error {
	linkedAccounts, err := b.LinkedAccounts().ListByWalletId(ctx, w.WalletID)
	if err != nil {
		return fmt.Errorf("listing linked accounts: %w", err)
	}
	for _, la := range linkedAccounts {
		if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBalance && !la.DeletedAt.Valid {
			w.XagoLinkedAccountID = la.ID
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

	if w.XagoLinkedAccountID != "" {
		bal, err := b.Xago().GetBalance(ctx, w.XagoLinkedAccountID)
		if err != nil {
			return fmt.Errorf("getting xago balance for linked account %s: %w", w.XagoLinkedAccountID, err)
		}
		w.Balance = bal.Total.FormatAmount()
		w.AvailableBalance = bal.Available.FormatAmount()
		w.BalanceCurrency = bal.Total.Currency.String()
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

	w.Transactions = make([]xagoReportTransaction, 0, len(txs))
	for _, t := range txs {
		amount := currency.Amount{Value: t.Amount, Currency: currency.ParseCurrency(t.Asset), Scale: t.Scale}
		fee := currency.Amount{Value: t.ProviderFee, Currency: amount.Currency, Scale: t.Scale}
		w.Transactions = append(w.Transactions, xagoReportTransaction{
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
