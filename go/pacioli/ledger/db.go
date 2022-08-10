package ledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	tb_types "github.com/coilhq/tigerbeetle-go/pkg/types"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/pacioli"
)

/* Transfer State Machine:
 *
 * Pending(init)───────────┬────────────►Voided(final)
 *                         │
 *                         │
 *                         └────────────►Replaced(final)
 *
 * Posted(init/final)
 */

type transferState int

const (
	transferStateUnknown  transferState = 0
	transferStatePending  transferState = 1
	transferStateVoided   transferState = 2
	transferStateReplaced transferState = 3
	transferStatePosted   transferState = 4
	transferStateSentinel transferState = 5 // Sanity check value
)

func (ts transferState) IsValid() bool {
	return ts > transferStateUnknown && ts < transferStateSentinel
}

func isValidStateTransition(init, target transferState) bool {
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

	return target == transferStateVoided || target == transferStateReplaced
}

func createDBTransfers(ctx context.Context, b Backends, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error) {

	var results []pacioli.TransferResult
	skipIndexes := make(map[int]bool)
	// Validations outside of DB lookups and covered by `validate` tags and tb logic copy
	for i, ta := range args {
		err := b.Validator().Struct(ta)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err.Error(), pacioli.ErrInvalidArg)
		}

		if ta.Flags.Pending && ta.Timeout == 0 {
			skipIndexes[i] = true
			results = append(results, pacioli.TransferResult{
				Index: uint32(i),
				Code:  tb_types.TransferPendingTransferMustTimeout,
			})
			continue
		}

	}

	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		// Check that all transfer accounts are in the same ledger
		for i, ta := range args {
			// Skip entries that have already failed validation
			if skipIndexes[i] {
				continue
			}

			var creditAcc, debitAcc pacioli.Account
			err := tx.GetContext(ctx, &debitAcc, "select * from ledger_accounts where id=$1", ta.DebitAccountID)
			if errors.Is(err, sql.ErrNoRows) {
				results = append(results, pacioli.TransferResult{
					Index: uint32(i),
					Code:  tb_types.TransferDebitAccountNotFound,
				})
				continue
			} else if err != nil {
				return err
			}

			err = tx.GetContext(ctx, &creditAcc, "select * from ledger_accounts where id=$1", ta.CreditAccountID)
			if errors.Is(err, sql.ErrNoRows) {
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
		}
		return nil
	})

	return results, err
}

/*    pub fn debits_exceed_credits(self: *const Account, amount: u64) bool {
    return (self.flags.debits_must_not_exceed_credits and
        self.debits_pending + self.debits_posted + amount > self.credits_posted);
}

pub fn credits_exceed_debits(self: *const Account, amount: u64) bool {
    return (self.flags.credits_must_not_exceed_debits and
        self.credits_pending + self.credits_posted + amount > self.debits_posted);
}*/
