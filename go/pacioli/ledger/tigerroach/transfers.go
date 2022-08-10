package tigerroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

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
			return nil, fmt.Errorf("%s %w", err.Error(), pacioli.ErrInvalidArg)
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

	}
	/*
		err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
			// Check that all transfer accounts are in the same ledger
			for i, ta := range args {
				// Skip entries that have already failed validation
				if skipIndexes[i] {
					continue
				}

				debitAcc, err := GetAccountTX(ctx, tx, ta.DebitAccountID)
				if errors.Is(err, pacioli.ErrNotFound) {
					results = append(results, pacioli.TransferResult{
						Index: uint32(i),
						Code:  tb_types.TransferDebitAccountNotFound,
					})
					continue
				} else if err != nil {
					return err
				}

				creditAcc, err := GetAccountTX(ctx, tx, ta.CreditAccountID)
				if errors.Is(err, pacioli.ErrNotFound) {
					results = append(results, pacioli.TransferResult{
						Index: uint32(i),
						Code:  tb_types.TransferCreditAccountNotFound,
					})
					continue
				} else if err != nil {
					return err
				}

				if debitAcc.LedgerID != creditAcc.LedgerID {
					results = append(results, pacioli.TransferResult{
						Index: uint32(i),
						Code:  tb_types.TransferAccountsMustHaveTheSameLedger,
					})
					continue
				}

				if debitAcc.LedgerID != ta.Ledger {
					results = append(results, pacioli.TransferResult{
						Index: uint32(i),
						Code:  tb_types.TransferTransferMustHaveTheSameLedgerAsAccounts,
					})
					continue
				}

				te, err := transferExists(ctx, tx, ta)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return err
				}
				if te == tb_types.TransferExists {
					noop[i] = true
				} else if te != 0 {
					results = append(results, pacioli.TransferResult{
						Index: uint32(i),
						Code:  te,
					})
					continue
				}

				if debitAcc.Flags.DebitsMustNotExceedCredits &&
					debitAcc.DebitsPosted+debitAcc.DebitsPending+ta.Amount > debitAcc.CreditsPosted {
					results = append(results, pacioli.TransferResult{
						Index: uint32(i),
						Code:  tb_types.TransferExceedsCredits,
					})
					continue
				}

				if creditAcc.Flags.CreditsMustNotExceedDebits &&
					debitAcc.CreditsPending+debitAcc.CreditsPosted+ta.Amount > debitAcc.DebitsPosted {
					results = append(results, pacioli.TransferResult{
						Index: uint32(i),
						Code:  tb_types.TransferExceedsDebits,
					})
					continue
				}
			}

			// Return validation errors and do not attempt to post transfers to the DB
			if len(results) > 0 {
				return nil
			}

			// Write transfers to the DB.
			stmt, err := tx.PrepareContext(ctx,
				"INSERT INTO ledger_transfers (id, ledger_id, debit_account_id, credit_account_id, amount, state, timeout_at) "+
					"VALUES ($1, $2, $3, $4, $5, $6, $7)")
			if err != nil {
				return err
			}
			defer stmt.Close()

			for i, ta := range args {
				if noop[i] {
					// Transfer exists, do nothing.
					continue
				}

				state := transferStatePosted
				var timeout sql.NullTime
				if ta.Flags.Pending {
					state = transferStatePending
					timeout = sql.NullTime{
						Time:  time.Now().Add(time.Second), // TODO Remember how to do this and check what the timeout is in value (ms or s)
						Valid: true,
					}
				}

				_, err = stmt.ExecContext(ctx, ta.ID, ta.Ledger, ta.DebitAccountID, ta.CreditAccountID, ta.Amount, state, timeout)
				if err != nil {
					return err
				}
			}

			return nil
		})
	*/
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

}

type ledgerTransfer struct {
	pacioli.Transfer
	State     transferState `db:"state"`
	Timeout   time.Time     `db:"timeout_at"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}

func GetTransferTX(ctx context.Context, tx *sqlx.Tx, id string) (*pacioli.Transfer, error) {
	var tr ledgerTransfer
	err := tx.GetContext(ctx, &tr, "SELECT * FROM ledger_transfers WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pacioli.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &pacioli.Transfer{
		ID:              tr.ID,
		PendingID:       tr.PendingID,
		LedgerID:        tr.LedgerID,
		DebitAccountID:  tr.DebitAccountID,
		CreditAccountID: tr.CreditAccountID,
		Amount:          tr.Amount,
		Flags: pacioli.TransferFlags{
			Pending: tr.State == transferStatePending,
		},
		Code:    tr.Code,
		Timeout: uint64(tr.Timeout.UnixMilli()), // TODO check
	}, nil
}

func transferExists(ctx context.Context, tx *sqlx.Tx, args pacioli.CreateTransferArgs) (pacioli.TransferResultCode, error) {
	var ex pacioli.Transfer
	err := tx.GetContext(ctx, &ex, "SELECT * FROM ledger_transfers WHERE id=$1", args.ID)
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
	if ex.Timeout != args.Timeout {
		return tb_types.TransferExistsWithDifferentTimeout, nil
	}
	if ex.Amount != args.Amount {
		return tb_types.TransferExistsWithDifferentAmount, nil
	}
	if ex.Code != args.Code {
		return tb_types.TransferPendingTransferHasDifferentCode, nil
	}

	return tb_types.TransferExists, nil
}
