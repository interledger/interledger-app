package ops

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/limits"
	"gitlab.com/fynbos/backend/transactions"
)

func yearStart() time.Time {
	year, _, _ := time.Now().Date()
	return time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
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
