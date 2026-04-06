package tigerroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/bits"
	"sort"
	"time"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/pacioli"
)

/*
 * Pending(init)───────────┬────────────►Voided(final)
 *                         │
 *                         └────────────►Timedout(final)
 * Posted(init/final)
 */

func validStateTransition(init, target pacioli.TransferState) bool {
	if !init.IsValid() || !target.IsValid() {
		return false
	}

	// Check for no-ops
	if init == target {
		return false
	}

	// Only state the machine can transition from
	if init != pacioli.TransferStatePending {
		return false
	}

	return target == pacioli.TransferStateVoided ||
		target == pacioli.TransferStateTimeout ||
		target == pacioli.TransferStatePosted
}

func CreateTransfers(ctx context.Context, b Backends, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error) {

	resMap := make(map[int]pacioli.TransferResult)
	// Validations outside of DB lookups and covered by `validate` tags and tb logic copy
	for i, ta := range args {
		err := b.Validator().Struct(ta)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err, pacioli.ErrInvalidArg)
		}

		if ta.Pending && ta.Timeout == 0 {
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  pacioli.TransferPendingTransferMustTimeout,
			}
			continue
		}
	}

	for i, ta := range args {
		if _, ok := resMap[i]; ok {
			// Already failed validation
			continue
		}

		// TODO: better handle optimistic updates
		code, err := createTransfer(ctx, b, ta)
		if err != nil {
			return nil, fmt.Errorf("%s %d %s %w", "transfer index: ", i, err, pacioli.ErrInternal)
		}

		if code == 0 {
			continue
		}

		resMap[i] = pacioli.TransferResult{
			Index: uint32(i),
			Code:  code,
		}
	}

	var res []pacioli.TransferResult
	for _, v := range resMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Index < res[j].Index
	})
	return res, nil
}

func createTransfer(ctx context.Context, b Backends, args pacioli.CreateTransferArgs) (pacioli.TransferResultCode, error) {
	debitAcc, err := GetAccount(ctx, b, args.DebitAccountID)
	if errors.Is(err, pacioli.ErrNotFound) {
		return pacioli.TransferDebitAccountNotFound, nil
	}
	if err != nil {
		return 0, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	creditAcc, err := GetAccount(ctx, b, args.CreditAccountID)
	if errors.Is(err, pacioli.ErrNotFound) {
		return pacioli.TransferCreditAccountNotFound, nil
	} else if err != nil {
		return 0, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	if debitAcc.ID == creditAcc.ID {
		return pacioli.TransferAccountsMustBeDifferent, nil
	}

	if debitAcc.LedgerID != creditAcc.LedgerID {
		return pacioli.TransferAccountsMustHaveTheSameLedger, nil
	}

	if debitAcc.LedgerID != args.Ledger {
		return pacioli.TransferTransferMustHaveTheSameLedgerAsAccounts, nil
	}

	code, err := transferExists(ctx, b, args)
	if err != nil && !errors.Is(err, pacioli.ErrNotFound) {
		return 0, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}
	if code == pacioli.TransferExists {
		return 0, nil
	}
	if code != 0 {
		return code, nil
	}

	// Transfer with that ID doesn't exist.

	// Check for overflows

	amount := uint64(args.Amount)
	if args.Pending {
		_, carry := bits.Add64(amount, debitAcc.DebitsPending, 0)
		if carry > 0 {
			return pacioli.TransferOverflowsDebitsPending, nil
		}
		_, carry = bits.Add64(amount, creditAcc.CreditsPending, 0)
		if carry > 0 {
			return pacioli.TransferOverflowsCreditsPending, nil
		}
	}
	_, carry := bits.Add64(amount, debitAcc.DebitsPosted, 0)
	if carry > 0 {
		return pacioli.TransferOverflowsDebitsPosted, nil
	}
	_, carry = bits.Add64(amount, creditAcc.CreditsPosted, 0)
	if carry > 0 {
		return pacioli.TransferOverflowsCreditsPosted, nil
	}
	_, carry = bits.Add64(amount+debitAcc.DebitsPosted, debitAcc.DebitsPending, 0)
	if carry > 0 {
		return pacioli.TransferOverflowsDebits, nil
	}
	_, carry = bits.Add64(amount+creditAcc.CreditsPosted, creditAcc.CreditsPending, 0)
	if carry > 0 {
		return pacioli.TransferOverflowsCredits, nil
	}

	if debitAcc.DebitsMustNotExceedCredits &&
		debitAcc.DebitsPosted+debitAcc.DebitsPending+amount > debitAcc.CreditsPosted {
		return pacioli.TransferExceedsCredits, nil
	}

	if creditAcc.CreditsMustNotExceedDebits &&
		debitAcc.CreditsPending+debitAcc.CreditsPosted+amount > debitAcc.DebitsPosted {
		return pacioli.TransferExceedsDebits, nil
	}

	// All validation passed, create entry and update account values
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		state := pacioli.TransferStatePosted
		var timeoutAt sql.NullTime
		if args.Pending {
			state = pacioli.TransferStatePending

			timeout := time.Now().Add(time.Nanosecond * time.Duration(args.Timeout)) // This is not a time that we will round to the nearest minute
			timeoutAt = sql.NullTime{
				Time:  time.Date(timeout.Year(), timeout.Month(), timeout.Day(), timeout.Hour(), timeout.Minute(), 0, 0, time.UTC),
				Valid: true,
			}
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO ledger_transfers (id, ledger_id, code, debit_account_id, credit_account_id, amount, state, timeout_at) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", args.ID, args.Ledger, args.Code, args.DebitAccountID, args.CreditAccountID, args.Amount, state, timeoutAt)
		if err != nil {
			return err
		}

		// Update Credit account
		sql := "UPDATE ledger_accounts SET credits_posted=credits_posted+$1, updated_at=now() WHERE id=$2 and credits_pending=$3 and credits_posted=$4"
		if state == pacioli.TransferStatePending {
			sql = "UPDATE ledger_accounts SET credits_pending=credits_pending+$1, updated_at=now() WHERE id=$2 and credits_pending=$3 and credits_posted=$4"
		}

		rows, err := tx.ExecContext(ctx, sql, args.Amount, args.CreditAccountID, creditAcc.CreditsPending, creditAcc.CreditsPosted)
		if err != nil {
			return fmt.Errorf("%s %w", err, pacioli.ErrInternal)
		}
		updateCnt, err := rows.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s %w", err, pacioli.ErrInternal)
		}
		if updateCnt != 1 {
			return fmt.Errorf("%s %s %w", "unable to update credit account", args.CreditAccountID, pacioli.ErrInternal)
		}

		// Update Debit account
		sql = "UPDATE ledger_accounts SET debits_posted=debits_posted+$1, updated_at=now() WHERE id=$2 and debits_pending=$3 and debits_posted=$4"
		if state == pacioli.TransferStatePending {
			sql = "UPDATE ledger_accounts SET debits_pending=debits_pending+$1, updated_at=now() WHERE id=$2 and debits_pending=$3 and debits_posted=$4"
		}

		rows, err = tx.ExecContext(ctx, sql, args.Amount, args.DebitAccountID, debitAcc.DebitsPending, debitAcc.DebitsPosted)
		if err != nil {
			return fmt.Errorf("%s %w", err, pacioli.ErrInternal)
		}
		updateCnt, err = rows.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s %w", err, pacioli.ErrInternal)
		}
		if updateCnt != 1 {
			return fmt.Errorf("%s %s %w", "unable to update debit account", args.DebitAccountID, pacioli.ErrInternal)
		}

		// All updates, inserts and validations passed
		return nil
	})

	return 0, err
}

type ledgerTransfer struct {
	pacioli.Transfer
	PendingID sql.NullString        `db:"pending_id"`
	State     pacioli.TransferState `db:"state"`
	Timeout   sql.NullTime          `db:"timeout_at"`
	CreatedAt time.Time             `db:"created_at"`
	UpdatedAt time.Time             `db:"updated_at"`
}

func getTransfer(ctx context.Context, b Backends, id string) (*ledgerTransfer, error) {
	var tr ledgerTransfer
	err := b.DB().GetContext(ctx, &tr, "SELECT * FROM ledger_transfers WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pacioli.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &tr, nil
}

func GetTransfer(ctx context.Context, b Backends, id string) (*pacioli.Transfer, error) {
	tr, err := getTransfer(ctx, b, id)
	if err != nil {
		return nil, err
	}

	return &pacioli.Transfer{
		ID:              tr.ID,
		LedgerID:        tr.LedgerID,
		DebitAccountID:  tr.DebitAccountID,
		CreditAccountID: tr.CreditAccountID,
		Amount:          tr.Amount,
		State:           tr.State,
		Code:            tr.Code,
		Timeout:         uint64(tr.Timeout.Time.UnixNano()),
	}, nil
}

func ListTransfers(ctx context.Context, b Backends, ids []string) ([]pacioli.Transfer, error) {
	var transfers []ledgerTransfer
	query, args, err := sqlx.In("SELECT * FROM ledger_transfers WHERE id IN (?);", ids)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}
	err = b.DB().SelectContext(ctx, &transfers, b.DB().Rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	resp := make([]pacioli.Transfer, len(transfers))
	for i, tr := range transfers {
		resp[i] = pacioli.Transfer{
			ID:              tr.ID,
			LedgerID:        tr.LedgerID,
			DebitAccountID:  tr.DebitAccountID,
			CreditAccountID: tr.CreditAccountID,
			Amount:          tr.Amount,
			State:           tr.State,
			Code:            tr.Code,
			Timeout:         uint64(tr.Timeout.Time.UnixNano()),
		}
	}

	return resp, nil
}

func ListTimedoutTransferIDs(ctx context.Context, b Backends) ([]string, error) {
	var transfers []string
	err := b.DB().SelectContext(ctx, &transfers,
		"SELECT id FROM ledger_transfers WHERE state=$1 and timeout_at < now()", pacioli.TransferStatePending)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	return transfers, err
}

// TryTimeoutTransfers will attemt to update the state of each of the transfers associated with the ids provided.
// Should a single transfer state fail to be updated the rest will still be updated. A list of successfully updated transfer IDs is returned.
func TryTimeoutTransfers(ctx context.Context, b Backends, ids []string) ([]string, error) {
	var success []string
	transfers, err := ListTransfers(ctx, b, ids)
	if err != nil {
		return nil, err
	}

	if len(transfers) != len(ids) {
		return nil, fmt.Errorf("list transfers did not load all expired transfers expected len(%d) actual len (%d)", len(transfers), len(ids))
	}

	for _, ex := range transfers {
		err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {

			rows, err := tx.ExecContext(ctx, "UPDATE ledger_transfers SET state=$1, updated_at=now() WHERE id=$2 and state=$3 and timeout_at<now()",
				pacioli.TransferStateTimeout, ex.ID, pacioli.TransferStatePending)
			if err != nil {
				return fmt.Errorf("%s %w", err, pacioli.ErrInternal)
			}
			count, err := rows.RowsAffected()
			if err != nil {
				return fmt.Errorf("%s %w", err, pacioli.ErrInternal)
			}
			if count != 1 {
				return fmt.Errorf("incorrect number of transfers timed out %w", pacioli.ErrInternal)
			}

			// Update Credit account balance
			rows, err = tx.ExecContext(ctx, "UPDATE ledger_accounts SET credits_pending=credits_pending-$1, updated_at=now() WHERE id=$2",
				ex.Amount, ex.CreditAccountID)
			if err != nil {
				return err
			}
			updateCnt, err := rows.RowsAffected()
			if err != nil {
				return err
			}
			if updateCnt != 1 {
				return fmt.Errorf("%s %s %w", "unable to update credit account balances for timeout", ex.CreditAccountID, pacioli.ErrInternal)
			}

			// Update Debit account balance
			rows, err = tx.ExecContext(ctx, "UPDATE ledger_accounts SET debits_pending=debits_pending-$1, updated_at=now() WHERE id=$2", ex.Amount, ex.DebitAccountID)
			if err != nil {
				return err
			}
			updateCnt, err = rows.RowsAffected()
			if err != nil {
				return err
			}
			if updateCnt != 1 {
				return fmt.Errorf("%s %s %w", "unable to update debit account balances for timeout", ex.CreditAccountID, pacioli.ErrInternal)
			}

			return nil
		})
		if err != nil {
			log.Error("transfer timeout failed", zap.Error(err))
		} else {
			success = append(success, ex.ID)
		}
	}

	return success, nil
}

func transferExists(ctx context.Context, b Backends, args pacioli.CreateTransferArgs) (pacioli.TransferResultCode, error) {
	ex, err := GetTransfer(ctx, b, args.ID)
	if err != nil {
		return 0, err
	}

	if ex.DebitAccountID != args.DebitAccountID {
		return pacioli.TransferExistsWithDifferentDebitAccountId, nil
	}
	if ex.CreditAccountID != args.CreditAccountID {
		return pacioli.TransferExistsWithDifferentCreditAccountId, nil
	}
	if args.Pending {
		// Compare timeouts with a minute grace period.
		newTimeout := time.Now().Add(time.Millisecond * time.Duration(args.Timeout))
		existingTimeout := time.Unix(0, int64(ex.Timeout))
		if newTimeout.Before(existingTimeout.Add(time.Minute)) &&
			newTimeout.After(existingTimeout.Add(time.Minute*-1)) {
			return pacioli.TransferExistsWithDifferentTimeout, nil
		}
	}

	amount := uint64(args.Amount)
	if ex.Amount != amount {
		return pacioli.TransferExistsWithDifferentAmount, nil
	}
	if ex.Code != args.Code {
		return pacioli.TransferExistsWithDifferentCode, nil
	}

	return pacioli.TransferExists, nil
}

func PostTransfers(ctx context.Context, b Backends, ids []string) (pendingPostedIDs map[string]string, res []pacioli.TransferResult, err error) {
	resMap := make(map[int]pacioli.TransferResult)
	pendingPostedIDs = make(map[string]string)
	for i, tid := range ids {
		ex, err := getTransfer(ctx, b, tid)
		if errors.Is(err, pacioli.ErrNotFound) {
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  pacioli.TransferPendingTransferNotFound,
			}
			continue
		}
		if err != nil {
			return nil, nil, err
		}

		if !validStateTransition(ex.State, pacioli.TransferStatePosted) {
			if ex.State == pacioli.TransferStatePosted {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  pacioli.TransferPendingTransferAlreadyPosted,
				}
				continue
			}
			if ex.State == pacioli.TransferStateVoided {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  pacioli.TransferPendingTransferAlreadyVoided,
				}
				continue
			}
			if ex.State != pacioli.TransferStatePending {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  pacioli.TransferPendingTransferNotPending,
				}
				continue
			}
			if ex.State == pacioli.TransferStateTimeout || ex.Timeout.Time.Before(time.Now().UTC()) {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  pacioli.TransferPendingTransferExpired,
				}
				continue
			}
			// Default catch all shouldn't execute
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  pacioli.TransferPendingTransferNotPending,
			}
			continue
		}

		err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
			// Update the old transaction's state
			rows, err := tx.ExecContext(ctx, "UPDATE ledger_transfers SET state=$1, updated_at=now() WHERE id=$2 and state=$3", pacioli.TransferStatePosted, ex.ID, pacioli.TransferStatePending)
			if err != nil {
				return err
			}
			updatedCnt, err := rows.RowsAffected()
			if err != nil {
				return err
			}
			if updatedCnt != 1 {
				return fmt.Errorf("%s %s %w", "failed to update pending tranfer state", ex.ID, pacioli.ErrInternal)
			}

			// Update Credit account balance
			rows, err = tx.ExecContext(ctx, "UPDATE ledger_accounts SET credits_posted=credits_posted+$1, credits_pending=credits_pending-$1, updated_at=now() WHERE id=$2", ex.Amount, ex.CreditAccountID)
			if err != nil {
				return err
			}
			updateCnt, err := rows.RowsAffected()
			if err != nil {
				return err
			}
			if updateCnt != 1 {
				return fmt.Errorf("%s %s %w", "unable to update credit account balances", ex.CreditAccountID, pacioli.ErrInternal)
			}

			// Update Debit account balance
			rows, err = tx.ExecContext(ctx, "UPDATE ledger_accounts SET debits_posted=debits_posted+$1, debits_pending=debits_pending-$1, updated_at=now() WHERE id=$2", ex.Amount, ex.DebitAccountID)
			if err != nil {
				return err
			}
			updateCnt, err = rows.RowsAffected()
			if err != nil {
				return err
			}
			if updateCnt != 1 {
				return fmt.Errorf("%s %s %w", "unable to update debit account balances", ex.CreditAccountID, pacioli.ErrInternal)
			}
			pendingPostedIDs[ex.ID] = ex.ID

			return nil
		})
		if err != nil {
			return nil, nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
		}
	}

	for _, v := range resMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Index < res[j].Index
	})
	return pendingPostedIDs, res, nil
}

func VoidTransfers(ctx context.Context, b Backends, ids []string) ([]pacioli.TransferResult, error) {
	resMap := make(map[int]pacioli.TransferResult)

	for i, tid := range ids {
		ex, err := getTransfer(ctx, b, tid)
		if errors.Is(err, pacioli.ErrNotFound) {
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  pacioli.TransferPendingTransferNotFound,
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		if !validStateTransition(ex.State, pacioli.TransferStateVoided) {
			if ex.State == pacioli.TransferStateVoided {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  pacioli.TransferPendingTransferAlreadyVoided,
				}
				continue
			}
			if ex.State == pacioli.TransferStatePosted {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  pacioli.TransferPendingTransferAlreadyPosted,
				}
				continue
			}
			if ex.State == pacioli.TransferStateTimeout || ex.Timeout.Time.Before(time.Now().UTC()) {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  pacioli.TransferPendingTransferExpired,
				}
				continue
			}
			if ex.State != pacioli.TransferStatePending {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  pacioli.TransferPendingTransferNotPending,
				}
				continue
			}

			// Catch all error, shouldn't happen.
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  pacioli.TransferPendingTransferNotPending,
			}
			continue
		}

		err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {

			// Update the old transaction's state
			rows, err := tx.ExecContext(ctx, "UPDATE ledger_transfers SET state=$1, updated_at=now() WHERE id=$2 and state=$3", pacioli.TransferStateVoided, ex.ID, pacioli.TransferStatePending)
			if err != nil {
				return err
			}
			updatedCnt, err := rows.RowsAffected()
			if err != nil {
				return err
			}
			if updatedCnt != 1 {
				return fmt.Errorf("%s %s %w", "failed to update pending tranfer state", ex.ID, pacioli.ErrInternal)
			}

			// Update Credit account balance
			rows, err = tx.ExecContext(ctx, "UPDATE ledger_accounts SET credits_pending=credits_pending-$1, updated_at=now() WHERE id=$2", ex.Amount, ex.CreditAccountID)
			if err != nil {
				return err
			}
			updateCnt, err := rows.RowsAffected()
			if err != nil {
				return err
			}
			if updateCnt != 1 {
				return fmt.Errorf("%s %s %w", "unable to update credit account balances", ex.CreditAccountID, pacioli.ErrInternal)
			}

			// Update Debit account balance
			rows, err = tx.ExecContext(ctx, "UPDATE ledger_accounts SET debits_pending=debits_pending-$1, updated_at=now() WHERE id=$2", ex.Amount, ex.DebitAccountID)
			if err != nil {
				return err
			}
			updateCnt, err = rows.RowsAffected()
			if err != nil {
				return err
			}
			if updateCnt != 1 {
				return fmt.Errorf("%s %s %w", "unable to update debit account balances", ex.CreditAccountID, pacioli.ErrInternal)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
		}
	}

	var res []pacioli.TransferResult
	for _, v := range resMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Index < res[j].Index
	})
	return res, nil
}
