package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/transactions"
)

type dbLimits struct {
	ID        string        `db:"id"`
	ForeignID string        `db:"foreign_id"`
	Type      limits.FKType `db:"type"`
	WalletID  string        `db:"wallet_id"`
	Currency  string        `db:"currency"`
	Daily     uint64        `db:"daily"`
	Monthly   uint64        `db:"monthly"`
	Overall   uint64        `db:"overall"`
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
func defaultLimits(ctx context.Context, b Backends, walletID, foreignID string, fkType limits.FKType) (*dbLimits, error) {

	_, err := b.DB().ExecContext(ctx, "INSERT INTO authorisation_limits (foreign_id, type, wallet_id, currency, daily, monthly, overall) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		foreignID, fkType, walletID, "USD", 10_00, 200_00, 1000_00)
	if err != nil && !db.IsErrorCode(err, db.UniqueViolationError) {
		return nil, fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	return getLimits(ctx, b, walletID, foreignID)
}

func yearStart() time.Time {
	year, _, _ := time.Now().Date()
	return time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func monthStart() time.Time {
	year, month, _ := time.Now().Date()
	return time.Date(year, month, 0, 0, 0, 0, 0, time.UTC)
}

func dayStart() time.Time {
	year, month, day := time.Now().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func Exceeds(ctx context.Context, b Backends, walletID, clientID string, amount currency.Amount) (bool, error) {

	// Lookup client limits
	clientLimits, err := getLimits(ctx, b, walletID, clientID)
	if errors.Is(err, limits.ErrNotFound) {
		clientLimits, err = defaultLimits(ctx, b, walletID, clientID, limits.FKTypeClient)
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

func GetPublicKeyLimit(ctx context.Context, b Backends, walletID, keyUuid string) (*limits.Limit, error) {
	lim, err := getLimits(ctx, b, walletID, keyUuid)
	if err != nil {
		return nil, err
	}

	return &limits.Limit{
		Daily:   currency.FromUInt64(lim.Daily, currency.Currency(lim.Currency)),
		Monthly: currency.FromUInt64(lim.Monthly, currency.Currency(lim.Currency)),
		Overall: currency.FromUInt64(lim.Overall, currency.Currency(lim.Currency)),
	}, nil
}

func UpdatePublicKeyLimits(ctx context.Context, b Backends, walletID, keyUuid string, limit limits.Limit) error {

	exists := true
	_, err := getLimits(ctx, b, walletID, keyUuid)
	if errors.Is(err, limits.ErrNotFound) {
		exists = false
	} else if err != nil {
		return err
	}

	if exists {
		_, err = b.DB().ExecContext(ctx, "UPDATE authorisation_limits SET daily=$1, monthly=$2, overall=$3, updated_at=now() WHERE wallet_id=$4 AND foreign_id=$5",
			limit.Daily.Value, limit.Monthly.Value, limit.Overall.Value, walletID, keyUuid)
		if err != nil {
			return fmt.Errorf("%w %s", limits.ErrInternal, err)
		}

		return nil
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO authorisation_limits (foreign_id, type, wallet_id, currency, daily, monthly, overall) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		keyUuid, limits.FKTypeClientPublicKey, walletID, "USD", limit.Daily.Value, limit.Monthly.Value, limit.Overall.Value)
	if err != nil {
		return fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	return nil
}

func UpdateClientLimits(ctx context.Context, b Backends, walletID, clientURL string, limit limits.Limit) error {

	// Lookup the auth client ID for the payment pointer
	var clientID string
	err := b.DB().GetContext(ctx, &clientID, "SELECT id FROM authorisation_clients WHERE url=$1", clientURL)
	if errors.Is(err, sql.ErrNoRows) {
		return limits.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	exists := true
	_, err = getLimits(ctx, b, walletID, clientID)
	if errors.Is(err, limits.ErrNotFound) {
		exists = false
	} else if err != nil {
		return err
	}

	if exists {
		_, err = b.DB().ExecContext(ctx, "UPDATE authorisation_limits SET daily=$1, monthly=$2, overall=$3, updated_at=now() WHERE wallet_id=$4 AND foreign_id=$5",
			limit.Daily.Value, limit.Monthly.Value, limit.Overall.Value, walletID, clientID)
		if err != nil {
			return fmt.Errorf("%w %s", limits.ErrInternal, err)
		}

		return nil
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO authorisation_limits (foreign_id, type, wallet_id, currency, daily, monthly, overall) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		clientID, limits.FKTypeClient, walletID, "USD", limit.Daily.Value, limit.Monthly.Value, limit.Overall.Value)
	if err != nil {
		return fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	return nil
}

func ListLimits(ctx context.Context, b Backends, walletID string) ([]limits.LimitConfigured, error) {
	var l []dbLimits
	err := b.DB().SelectContext(ctx, &l,
		"SELECT id, foreign_id, type, wallet_id, currency, daily, monthly, overall FROM authorisation_limits WHERE wallet_id=$1",
		walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, limits.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	resp := make([]limits.LimitConfigured, len(l))
	for i, dbl := range l {
		var display string
		switch dbl.Type {
		case limits.FKTypeClient:
			err = b.DB().GetContext(ctx, &display, "SELECT url FROM authorisation_clients WHERE id=$1", dbl.ForeignID)
			if err != nil {
				return nil, fmt.Errorf("%w %s", limits.ErrInternal, err)
			}
		}

		resp[i] = limits.LimitConfigured{
			Limit: limits.Limit{
				Daily:   currency.FromUInt64(dbl.Daily, currency.ParseCurrency(dbl.Currency)),
				Monthly: currency.FromUInt64(dbl.Monthly, currency.ParseCurrency(dbl.Currency)),
				Overall: currency.FromUInt64(dbl.Overall, currency.ParseCurrency(dbl.Currency)),
			},
			ForeignID:      dbl.ForeignID,
			ForeignType:    dbl.Type,
			ForeignDisplay: display,
		}
	}

	return resp, nil
}

func ExceedsKYCLimits(ctx context.Context, b Backends, walletID string, amount currency.Amount) (bool, limits.LimitType, error) {
	switch amount.Currency {
	case currency.USD:
		return exceedsKYCLimitsUSD(ctx, b, walletID, amount)
	case currency.ZAR:
		return exceedsKYCLimitsZAR(ctx, b, walletID, amount)
	case currency.CAD:
		return exceedsKYCLimitsCAD(ctx, b, walletID, amount)
	default:
		log.Warn("Unknown currency to apply limits to", zap.String("currency", amount.Currency.String()))
	}
	return false, "", nil
}

func exceedsKYCLimitsCAD(ctx context.Context, b Backends, walletID string, amount currency.Amount) (bool, limits.LimitType, error) {
	if amount.Value > 10_000_00 {
		return true, limits.LimitTypeTransaction, nil
	}

	return false, "", nil
}

func exceedsKYCLimitsZAR(ctx context.Context, b Backends, walletID string, amount currency.Amount) (bool, limits.LimitType, error) {
	const limitYear uint64 = 20_000_00 // R20k limit for the year

	level, err := b.KYC().GetKYCStatus(ctx, walletID)
	if err != nil {
		return false, "", fmt.Errorf("%w %s", limits.ErrInternal, err)
	}
	// No limits on KYC level 2
	if level == kyc.StatusLevel2 {
		return false, "", nil
	}

	if amount.Value > limitYear {
		return true, limits.LimitTypeYearly, nil
	}

	var used sql.NullInt64
	err = b.DB().GetContext(ctx, &used, "SELECT sum(amount) FROM transactions WHERE wallet_id=$1 AND created_at>$2 AND state IN ($3,$4) AND asset_code='ZAR' AND type NOT IN ($5, $6)",
		walletID, yearStart(), transactions.StatePending, transactions.StateCompleted, transactions.TransactionTypeDeposit, transactions.TransactionTypeWithdrawal)
	if err != nil {
		return false, "", fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	// The user has no transactions
	if !used.Valid {
		return false, "", nil
	}

	if uint64(used.Int64)+amount.Value >= limitYear {
		return true, limits.LimitTypeYearly, nil
	}

	return false, "", nil
}

func exceedsKYCLimitsUSD(ctx context.Context, b Backends, walletID string, amount currency.Amount) (bool, limits.LimitType, error) {
	// Only supporting L1 Limits for now:
	// Transaction 	$   250.00
	// 24-Hour    	$ 2,999.00
	// 30-Day 		$ 6,000.00
	// 180-Day 		$ 9,999.00
	var limitTransaction float64
	var limit24Hour, limit30Day, limit180Day uint64

	level, err := b.KYC().GetKYCStatus(ctx, walletID)
	if err != nil {
		return false, "", fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	switch level {
	case kyc.StatusLevel1, kyc.StatusLevel2:
		limitTransaction = 250.0
		limit24Hour = 2999_00
		limit30Day = 6000_00
		limit180Day = 9_999_00
	default:
		limitTransaction = 0.0
		limit24Hour = 0
		limit30Day = 0
		limit180Day = 0
	}

	// Short circuit.
	if amount.Float64() >= limitTransaction {
		return true, limits.LimitTypeTransaction, nil
	}

	stmt, err := b.DB().PreparexContext(ctx, "SELECT sum(amount) FROM transactions WHERE wallet_id=$1 AND created_at>$2 AND state IN ($3,$4) AND asset_code='USD'")
	if err != nil {
		return false, "", fmt.Errorf("%w %s", limits.ErrInternal, err)
	}
	defer stmt.Close()

	var used sql.NullInt64
	err = stmt.GetContext(ctx, &used,
		walletID, time.Now().Add(time.Hour*-24), transactions.StatePending, transactions.StateCompleted)
	if err != nil {
		return false, "", fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	// The user has no transactions
	if !used.Valid {
		return false, "", nil
	}

	if uint64(used.Int64)+amount.Value >= limit24Hour {
		return true, limits.LimitTypeDaily, nil
	}

	err = stmt.GetContext(ctx, &used,
		walletID, time.Now().Add(time.Hour*-24*30), transactions.StatePending, transactions.StateCompleted)
	if err != nil {
		return false, "", fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	if uint64(used.Int64)+amount.Value >= limit30Day {
		return true, limits.LimitTypeMonthly, nil
	}

	err = stmt.GetContext(ctx, &used,
		walletID, time.Now().Add(time.Hour*-24*180), transactions.StatePending, transactions.StateCompleted)
	if err != nil {
		return false, "", fmt.Errorf("%w %s", limits.ErrInternal, err)
	}

	if uint64(used.Int64)+amount.Value >= limit180Day {
		return true, limits.LimitType6Monthly, nil
	}

	return false, "", nil
}
