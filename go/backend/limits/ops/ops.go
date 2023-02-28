package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/transactions"
)

type dbLimits struct {
	ID        string `db:"id"`
	ForeignID string `db:"foreign_id"`
	Type      string `db:"type"`
	WalletID  string `db:"wallet_id"`
	Currency  string `db:"currency"`
	Daily     uint64 `db:"daily"`
	Monthly   uint64 `db:"monthly"`
	Overall   uint64 `db:"overall"`
}

func getLimits(ctx context.Context, b Backends, walletID, foreignID string) (*dbLimits, error) {
	var l dbLimits
	err := b.DB().GetContext(ctx, &l,
		"SELECT id, foreign_id, type, wallet_id, currency, daily, monthly, overall FROM authorisation_limits WHERE wallet_id=$1 AND foreign_id=$2",
		walletID, foreignID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, limits.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	return &l, nil
}

// defaultLimits inserts default limits for foreign_id as a fallback in case this step was skipped by the UI.
func defaultLimits(ctx context.Context, b Backends, walletID, foreignID string, fkType limits.LimitFKType) (*dbLimits, error) {

	// Lookup the payment pointer for it's currency.
	pp, err := b.OpenPayments().GetWalletPaymentPointer(ctx, walletID)
	if err != nil {
		return nil, err
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO authorisation_limits (foreign_id, type, wallet_id, currency, daily, monthly, overall) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		foreignID, fkType, walletID, pp.Asset, 10_00, 200_00, 1000_00)
	if err != nil && !db.IsErrorCode(err, db.UniqueViolationError) {
		return nil, fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	return getLimits(ctx, b, walletID, foreignID)
}

func Exceeds(ctx context.Context, b Backends, walletID, clientID string, amount currency.Amount) (bool, error) {

	// Lookup client limits
	clientLimits, err := getLimits(ctx, b, walletID, clientID)
	if errors.Is(err, limits.ErrNotFound) {
		clientLimits, err = defaultLimits(ctx, b, walletID, clientID, limits.LimitFKTypeClient)
	}
	if err != nil {
		return false, err
	}

	if clientLimits.Daily > 0 {
		var clientDaily uint64
		err = b.DB().GetContext(ctx, &clientDaily,
			"SELECT sum(amount) FROM transactions WHERE grant_id in (SELECT id FROM authorisation_grants WHERE client_id=$1) AND wallet_id=$2 AND created_at>$3 AND state IN ($4,$5)",
			clientID, walletID, dayStart(), transactions.StatePending, transactions.StateCompleted)
		if err != nil {
			return false, fmt.Errorf("%w %s", limits.ErrInternal, err)
		}

		if clientLimits.Daily < amount.Value+clientDaily {
			return true, nil
		}
	}

	if clientLimits.Monthly > 0 {
		var clientMonthly uint64
		err = b.DB().GetContext(ctx, &clientMonthly,
			"SELECT sum(amount) FROM transactions WHERE grant_id in (SELECT id FROM authorisation_grants WHERE client_id=$1) AND wallet_id=$2 AND created_at>$3 AND state IN ($4,$5)",
			clientID, walletID, monthStart(), transactions.StatePending, transactions.StateCompleted)
		if err != nil {
			return false, fmt.Errorf("%w %s", limits.ErrInternal, err)
		}

		if clientLimits.Monthly < amount.Value+clientMonthly {
			return true, nil
		}
	}

	if clientLimits.Overall > 0 {
		var clientOverall uint64
		err = b.DB().GetContext(ctx, &clientOverall,
			"SELECT sum(amount) FROM transactions WHERE grant_id in (SELECT id FROM authorisation_grants WHERE client_id=$1) AND wallet_id=$2 AND state IN ($3,$4)",
			clientID, walletID, transactions.StatePending, transactions.StateCompleted)
		if err != nil {
			return false, fmt.Errorf("%w %s", limits.ErrInternal, err)
		}

		if clientLimits.Overall < amount.Value+clientOverall {
			return true, nil
		}
	}

	return false, nil
}

func monthStart() time.Time {
	year, month, _ := time.Now().Date()
	return time.Date(year, month, 0, 0, 0, 0, 0, time.UTC)
}

func dayStart() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
