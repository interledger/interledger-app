package tigerroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"

	tb_types "github.com/coilhq/tigerbeetle-go/pkg/types"

	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/pacioli"
)

/*
 * Pending(init)───────────┬────────────►Voided(final)
 *                         │
 *                         ├────────────►Replaced(final)
 *                         │
 *                         └────────────►Timedout(final)
 * Posted(init/final)
 */

type transferState int

const (
	transferStateUnknown  transferState = 0
	transferStatePending  transferState = 1
	transferStateVoided   transferState = 2
	transferStateReplaced transferState = 3
	transferStateTimeout  transferState = 4
	transferStatePosted   transferState = 5
	transferStateSentinel transferState = 6 // Sanity check value
)

func (ts transferState) IsValid() bool {
	return ts > transferStateUnknown && ts < transferStateSentinel
}

func validStateTransition(init, target transferState) bool {
	if !init.IsValid() || !target.IsValid() {
		return false
	}

	// Check for no-ops
	if init == target {
		return false
	}

	// Only state the machine can transition from
	if init != transferStatePending {
		return false
	}

	return target == transferStateVoided ||
		target == transferStateReplaced ||
		target == transferStateTimeout
}

func CreateTransfers(ctx context.Context, b Backends, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error) {

	resMap := make(map[int]pacioli.TransferResult)
	// Validations outside of DB lookups and covered by `validate` tags and tb logic copy
	for i, ta := range args {
		err := b.Validator().Struct(ta)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err, pacioli.ErrInvalidArg)
		}

		if ta.Flags.Pending && ta.Timeout == 0 {
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  tb_types.TransferPendingTransferMustTimeout,
			}
			continue
		}
	}

	for i, ta := range args {
		if _, ok := resMap[i]; ok {
			// Already failed validation
			continue
		}

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
		return tb_types.TransferDebitAccountNotFound, nil
	}
	if err != nil {
		return 0, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	creditAcc, err := GetAccount(ctx, b, args.CreditAccountID)
	if errors.Is(err, pacioli.ErrNotFound) {
		return tb_types.TransferCreditAccountNotFound, nil
	} else if err != nil {
		return 0, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	if debitAcc.ID == creditAcc.ID {
		return tb_types.TransferAccountsMustBeDifferent, nil
	}

	if debitAcc.LedgerID != creditAcc.LedgerID {
		return tb_types.TransferAccountsMustHaveTheSameLedger, nil
	}

	if debitAcc.LedgerID != args.Ledger {
		return tb_types.TransferTransferMustHaveTheSameLedgerAsAccounts, nil
	}

	code, err := transferExists(ctx, b, args)
	if err != nil && !errors.Is(err, pacioli.ErrNotFound) {
		return 0, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}
	if code == tb_types.TransferExists {
		return 0, nil
	}
	if code != 0 {
		return code, nil
	}

	// Transfer with that ID doesn't exist.

	if debitAcc.Flags.DebitsMustNotExceedCredits &&
		debitAcc.DebitsPosted+debitAcc.DebitsPending+args.Amount > debitAcc.CreditsPosted {
		return tb_types.TransferExceedsCredits, nil
	}

	if creditAcc.Flags.CreditsMustNotExceedDebits &&
		debitAcc.CreditsPending+debitAcc.CreditsPosted+args.Amount > debitAcc.DebitsPosted {
		return tb_types.TransferExceedsDebits, nil
	}

	// All validation passed, create entry and update account values
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		state := transferStatePosted
		var timeoutAt sql.NullTime
		if args.Flags.Pending {
			state = transferStatePending

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
		sql := "UPDATE ledger_accounts SET credits_posted=credits_posted+$1, updated_at=now() WHERE id=$2"
		if state == transferStatePending {
			sql = "UPDATE ledger_accounts SET credits_pending=credits_pending+$1, updated_at=now() WHERE id=$2"
		}

		rows, err := tx.ExecContext(ctx, sql, args.Amount, args.CreditAccountID)
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
		sql = "UPDATE ledger_accounts SET debits_posted=debits_posted+$1, updated_at=now() WHERE id=$2"
		if state == transferStatePending {
			sql = "UPDATE ledger_accounts SET debits_pending=debits_pending+$1, updated_at=now() WHERE id=$2"
		}

		rows, err = tx.ExecContext(ctx, sql, args.Amount, args.DebitAccountID)
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
	PendingID sql.NullString `db:"pending_id"`
	State     transferState  `db:"state"`
	Timeout   sql.NullTime   `db:"timeout_at"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
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
		PendingID:       tr.PendingID.String,
		LedgerID:        tr.LedgerID,
		DebitAccountID:  tr.DebitAccountID,
		CreditAccountID: tr.CreditAccountID,
		Amount:          tr.Amount,
		Flags: pacioli.TransferFlags{
			Pending: tr.State == transferStatePending,
		},
		Code:    tr.Code,
		Timeout: uint64(tr.Timeout.Time.UnixNano()),
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
			PendingID:       tr.PendingID.String,
			LedgerID:        tr.LedgerID,
			DebitAccountID:  tr.DebitAccountID,
			CreditAccountID: tr.CreditAccountID,
			Amount:          tr.Amount,
			Flags: pacioli.TransferFlags{
				Pending: tr.State == transferStatePending,
			},
			Code:    tr.Code,
			Timeout: uint64(tr.Timeout.Time.UnixNano()),
		}
	}

	return resp, nil
}

func transferExists(ctx context.Context, b Backends, args pacioli.CreateTransferArgs) (pacioli.TransferResultCode, error) {
	ex, err := GetTransfer(ctx, b, args.ID)
	if err != nil {
		return 0, err
	}

	if ex.Flags.ToUint16() != args.Flags.ToUint16() {
		return tb_types.TransferExistsWithDifferentFlags, nil
	}
	if ex.DebitAccountID != args.DebitAccountID {
		return tb_types.TransferExistsWithDifferentDebitAccountId, nil
	}
	if ex.CreditAccountID != args.CreditAccountID {
		return tb_types.TransferExistsWithDifferentCreditAccountId, nil
	}
	if args.Flags.Pending {
		// Compare timeouts with a minute grace period.
		newTimeout := time.Now().Add(time.Millisecond * time.Duration(args.Timeout))
		existingTimeout := time.Unix(0, int64(ex.Timeout))
		if newTimeout.Before(existingTimeout.Add(time.Minute)) &&
			newTimeout.After(existingTimeout.Add(time.Minute*-1)) {
			return tb_types.TransferExistsWithDifferentTimeout, nil
		}
	}

	if ex.Amount != args.Amount {
		return tb_types.TransferExistsWithDifferentAmount, nil
	}
	if ex.Code != args.Code {
		return tb_types.TransferExistsWithDifferentCode, nil
	}

	return tb_types.TransferExists, nil
}

func PostTransfers(ctx context.Context, b Backends, ids []string) (map[string]string, []pacioli.TransferResult, error) {
	resMap := make(map[int]pacioli.TransferResult)
	existing := make(map[string]ledgerTransfer)
	newUUIDs := make(map[string]string)
	for i, tid := range ids {
		ex, err := getTransfer(ctx, b, tid)
		if errors.Is(err, pacioli.ErrNotFound) {
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  tb_types.TransferPendingTransferNotFound,
			}
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		if validStateTransition(ex.State, transferStateReplaced) {
			if ex.State == transferStateReplaced {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  tb_types.TransferPendingTransferAlreadyPosted,
				}
				continue
			}
			if ex.State == transferStateVoided {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  tb_types.TransferPendingTransferAlreadyVoided,
				}
				continue
			}
			if ex.State != transferStatePending {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  tb_types.TransferPendingTransferNotPending,
				}
				continue
			}
			if ex.State == transferStateTimeout || ex.Timeout.Time.Before(time.Now().UTC()) {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  tb_types.TransferPendingTransferExpired,
				}
				continue
			}
			// Default catch all shouldn't execute
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  tb_types.TransferPendingTransferNotPending,
			}
			continue
		}
		existing[tid] = *ex
	}

	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		// Loop over actual existing transactions
		for _, ex := range existing {
			// Insert the new transaction
			newID := uuid.NewString()
			_, err := tx.ExecContext(ctx, "INSERT INTO ledger_transfers (id, ledger_id, code, debit_account_id, credit_account_id, amount, state, pending_id) "+
				"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", newID, ex.LedgerID, ex.Code, ex.DebitAccountID, ex.CreditAccountID, ex.Amount, transferStatePosted, ex.ID)
			if err != nil {
				return err
			}

			// Update the old transaction's state
			rows, err := tx.ExecContext(ctx, "UPDATE ledger_transfers SET state=$1, updated_at=now() WHERE id=$2 and state=$3", transferStateReplaced, ex.ID, transferStatePending)
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
				return fmt.Errorf("%s %s %w", "unable to update credit account balances", ex.CreditAccountID, pacioli.ErrInternal)
			}
			newUUIDs[ex.ID] = newID
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	var res []pacioli.TransferResult
	for _, v := range resMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Index < res[j].Index
	})
	return newUUIDs, res, nil
}

func VoidTransactions(ctx context.Context, b Backends, ids []string) ([]pacioli.TransferResult, error) {
	resMap := make(map[int]pacioli.TransferResult)
	for i, tid := range ids {
		ex, err := getTransfer(ctx, b, tid)
		if errors.Is(err, pacioli.ErrNotFound) {
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  tb_types.TransferPendingTransferNotFound,
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		if validStateTransition(ex.State, transferStateVoided) {
			if ex.State == transferStateVoided {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  tb_types.TransferPendingTransferAlreadyVoided,
				}
				continue
			}
			if ex.State == transferStateReplaced {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  tb_types.TransferPendingTransferAlreadyPosted,
				}
				continue
			}
			if ex.State == transferStateTimeout || ex.Timeout.Time.Before(time.Now().UTC()) {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  tb_types.TransferPendingTransferExpired,
				}
				continue
			}
			if ex.State != transferStatePending {
				resMap[i] = pacioli.TransferResult{
					Index: uint32(i),
					Code:  tb_types.TransferPendingTransferNotPending,
				}
				continue
			}

			// Catch all error, shouldn't happen.
			resMap[i] = pacioli.TransferResult{
				Index: uint32(i),
				Code:  tb_types.TransferPendingTransferNotPending,
			}
		}

		err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
			// Update the old transaction's state
			rows, err := tx.ExecContext(ctx, "UPDATE ledger_transfers SET state=$1, updated_at=now() WHERE id=$2 and state=$3", transferStateVoided, ex.ID, transferStatePending)
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
				return fmt.Errorf("%s %s %w", "unable to update credit account balances", ex.CreditAccountID, pacioli.ErrInternal)
			}

			return nil
		})

		if err != nil {
			return newUUIDs, nil, err
		}
	}

	var res []pacioli.TransferResult
	for _, v := range resMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Index < res[j].Index
	})
	return newUUIDs, res, nil
}
