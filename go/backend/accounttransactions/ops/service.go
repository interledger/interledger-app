package ops

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	tb_types "github.com/coilhq/tigerbeetle-go/pkg/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	account_transactions "gitlab.com/fynbos/backend/accounttransactions"
	"gitlab.com/fynbos/pacioli"
)

type accountTransaction struct {
	account_transactions.AccountTransaction
	TransferIDs pq.StringArray `db:"transfer_ids"`
}

// Calls out to Pacioli first and then inserts an account transaction into CRDB.
// TODO: this should return an error array
func Create(
	ctx context.Context, b Backends,
	args *account_transactions.CreateTransactionArgs,
) (*account_transactions.AccountTransaction, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInvalidArgument, err.Error())
	}

	acc, err := b.Accounts().Get(ctx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
	}

	var transaction accountTransaction
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {

		ledgerTransfers := make([]pacioli.CreateTransferArgs, len(args.LedgerTransfers))
		transferIDs := make([]string, len(args.LedgerTransfers))
		for i, transfer := range args.LedgerTransfers {
			id := uuid.NewString()
			transferIDs[i] = id
			ledgerTransfers[i] = pacioli.CreateTransferArgs{
				ID:              id,
				DebitAccountID:  transfer.DebitAccountID,
				CreditAccountID: transfer.CreditAccountID,
				Amount:          transfer.Amount,
				Code:            transfer.Code,
				Ledger:          transfer.LedgerID,
				Flags: pacioli.TransferFlags{
					Linked:  transfer.Flags.Linked,
					Pending: false,
				},
			}
		}

		transferErrors, err := b.Pacioli().CreateTransfers(ctx, ledgerTransfers)
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		if len(transferErrors) > 0 {
			for _, err := range transferErrors {
				switch err.Code {
				case tb_types.TransferExceedsCredits:
					return fmt.Errorf("%w %+v", account_transactions.ErrExceedsCredits, err)
				case tb_types.TransferExceedsDebits:
					return fmt.Errorf("%w %+v", account_transactions.ErrExceedsDebits, err)
				default:
					return fmt.Errorf("%w %+v", account_transactions.ErrInvalidLedgerTransfer, err)
				}
			}
		}

		stmt, err := tx.PrepareNamedContext(ctx, `INSERT INTO account_transactions
		(account_id, type, description, net_amount, state, transfer_ids) VALUES
		(:accountid, :type, :description, :netamount, :state, :transfer_ids)
		RETURNING *;
		`,
		)
		if err != nil {
			return fmt.Errorf("%s %w", err.Error(), account_transactions.ErrInternal)
		}

		err = stmt.Stmt.Get(
			&transaction,
			acc.ID,
			args.Type,
			args.Description,
			args.NetAmount,
			account_transactions.Posted.String(),
			pq.StringArray(transferIDs),
		)
		if err != nil {
			return fmt.Errorf("%s %w", err.Error(), account_transactions.ErrInternal)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &account_transactions.AccountTransaction{
		ID:          transaction.ID,
		Type:        transaction.Type,
		AccountID:   transaction.AccountID,
		Description: transaction.Description,
		State:       transaction.State,
		NetAmount:   transaction.NetAmount,
		TransferIDs: transaction.TransferIDs,
		CreatedAt:   transaction.CreatedAt,
		UpdatedAt:   transaction.UpdatedAt,
	}, nil
}

func CreatePending(
	ctx context.Context,
	b Backends,
	args *account_transactions.CreatePendingTransactionArgs,
) (*account_transactions.AccountTransaction, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInvalidArgument, err.Error())
	}

	acc, err := b.Accounts().Get(ctx, args.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
	}

	// Default to 48 Hours timeout
	duration, err := time.ParseDuration("48h")
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
	}

	var transaction accountTransaction
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {

		ledgerTransfers := make([]pacioli.CreateTransferArgs, len(args.LedgerTransfers))
		transferIDs := make([]string, len(args.LedgerTransfers))
		for i, transfer := range args.LedgerTransfers {
			id := uuid.NewString()
			transferIDs[i] = id
			ledgerTransfers[i] = pacioli.CreateTransferArgs{
				ID:              id,
				DebitAccountID:  transfer.DebitAccountID,
				CreditAccountID: transfer.CreditAccountID,
				Amount:          transfer.Amount,
				Code:            transfer.Code,
				Timeout:         uint64(duration.Nanoseconds()),
				Ledger:          transfer.LedgerID,
				Flags: pacioli.TransferFlags{
					Linked:  transfer.Flags.Linked,
					Pending: true,
				},
			}
		}

		transferErrors, err := b.Pacioli().CreateTransfers(ctx, ledgerTransfers)
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		if len(transferErrors) > 0 {
			for _, err := range transferErrors {
				switch err.Code {
				case tb_types.TransferExceedsCredits:
					return fmt.Errorf("%w %+v", account_transactions.ErrExceedsCredits, err)
				case tb_types.TransferExceedsDebits:
					return fmt.Errorf("%w %+v", account_transactions.ErrExceedsDebits, err)
				default:
					return fmt.Errorf("%w %+v", account_transactions.ErrInvalidLedgerTransfer, err)
				}
			}
		}

		stmt, err := tx.PrepareNamedContext(ctx, `INSERT INTO account_transactions
		(account_id, type, description, net_amount, state, transfer_ids) VALUES
		(:accountid, :type, :description, :netamount, :state, :transfer_ids)
		RETURNING *;
		`,
		)
		if err != nil {
			return fmt.Errorf("%s %w", err.Error(), account_transactions.ErrInternal)
		}

		err = stmt.Stmt.Get(
			&transaction,
			acc.ID,
			args.Type,
			args.Description,
			args.NetAmount,
			account_transactions.Pending.String(),
			pq.StringArray(transferIDs),
		)
		if err != nil {
			return fmt.Errorf("%s %w", err.Error(), account_transactions.ErrInternal)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &account_transactions.AccountTransaction{
		ID:          transaction.ID,
		Type:        transaction.Type,
		AccountID:   transaction.AccountID,
		Description: transaction.Description,
		State:       transaction.State,
		NetAmount:   transaction.NetAmount,
		TransferIDs: transaction.TransferIDs,
		CreatedAt:   transaction.CreatedAt,
		UpdatedAt:   transaction.UpdatedAt,
	}, nil
}

func PostPending(ctx context.Context, b Backends, id string) (*account_transactions.AccountTransaction, error) {
	trx := &accountTransaction{}
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		err := b.DB().GetContext(
			ctx,
			trx,
			"SELECT * FROM account_transactions WHERE id=$1 LIMIT 1;",
			id,
		)

		if err == sql.ErrNoRows {
			return account_transactions.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		// Check if state is pending
		if trx.State != account_transactions.Pending {
			return fmt.Errorf("account transaction: trx not in pending state actual is %s", trx.State)
		}

		commitErrors, err := b.Pacioli().PostTransfers(ctx, trx.TransferIDs)
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		// TODO need to handle the case where this was retried and is safe to do so... ie committed already.
		if len(commitErrors) > 0 {
			for _, err := range commitErrors {
				switch err.Code {
				default:
					return fmt.Errorf("%w %+v", account_transactions.ErrInvalidLedgerTransfer, err)
				}
			}
		}

		_, err = b.DB().Exec("UPDATE account_transactions set state = $1 where id = $2", account_transactions.Posted.String(), id)
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		err = b.DB().GetContext(
			ctx,
			trx,
			"SELECT * FROM account_transactions WHERE id=$1 LIMIT 1;",
			id,
		)

		if err == sql.ErrNoRows {
			return account_transactions.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &account_transactions.AccountTransaction{
		ID:          trx.ID,
		Type:        trx.Type,
		AccountID:   trx.AccountID,
		Description: trx.Description,
		State:       trx.State,
		NetAmount:   trx.NetAmount,
		TransferIDs: []string(trx.TransferIDs),
		CreatedAt:   trx.CreatedAt,
		UpdatedAt:   trx.UpdatedAt,
	}, nil
}

func VoidPending(ctx context.Context, b Backends, id string) (*account_transactions.AccountTransaction, error) {
	trx := &accountTransaction{}
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		err := b.DB().GetContext(
			ctx,
			trx,
			"SELECT * FROM account_transactions WHERE id=$1 LIMIT 1;",
			id,
		)
		if err == sql.ErrNoRows {
			return account_transactions.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		// Check if state is pending
		if trx.State != account_transactions.Pending {
			return fmt.Errorf("account transaction: trx not in pending state actual is %s", trx.State)
		}

		voidErrors, err := b.Pacioli().VoidTransfers(ctx, trx.TransferIDs)
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		// TODO need to handle the case where this was retried and is safe to do so... ie voided already.
		if len(voidErrors) > 0 {
			for _, err := range voidErrors {
				switch err.Code {
				default:
					return fmt.Errorf("%w %+v", account_transactions.ErrInvalidLedgerTransfer, err)
				}
			}
		}

		_, err = b.DB().Exec("UPDATE account_transactions set state = $1 where id = $2", account_transactions.Voided.String(), id)
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		err = b.DB().GetContext(
			ctx,
			trx,
			"SELECT * FROM account_transactions WHERE id=$1 LIMIT 1;",
			id,
		)
		if err == sql.ErrNoRows {
			return account_transactions.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &account_transactions.AccountTransaction{
		ID:          trx.ID,
		Type:        trx.Type,
		AccountID:   trx.AccountID,
		Description: trx.Description,
		State:       trx.State,
		NetAmount:   trx.NetAmount,
		TransferIDs: trx.TransferIDs,
		CreatedAt:   trx.CreatedAt,
		UpdatedAt:   trx.UpdatedAt,
	}, nil
}

func GetByAccount(
	ctx context.Context,
	b Backends,
	tx *sqlx.Tx,
	args *account_transactions.GetByAccountArgs,
) ([]*account_transactions.AccountTransaction, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInvalidArgument, err.Error())
	}

	var transactions []accountTransaction
	err = tx.SelectContext(ctx, &transactions,
		fmt.Sprintf(
			"SELECT * FROM account_transactions WHERE account_id=$1 ORDER BY created_at %s LIMIT $2;",
			args.OrderBy,
		),
		args.AccountID,
		args.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
	}

	ret := make([]*account_transactions.AccountTransaction, len(transactions))
	for i, trx := range transactions {

		ret[i] = &account_transactions.AccountTransaction{
			ID:          trx.ID,
			Type:        trx.Type,
			AccountID:   trx.AccountID,
			Description: trx.Description,
			State:       trx.State,
			NetAmount:   trx.NetAmount,
			TransferIDs: trx.TransferIDs,
			CreatedAt:   trx.CreatedAt,
			UpdatedAt:   trx.UpdatedAt,
		}
	}

	return ret, nil
}

func GetPage(ctx context.Context, b Backends, args *account_transactions.PaginationArgs) ([]account_transactions.AccountTransaction, error) {
	err := b.Validator().Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInvalidArgument, err.Error())
	}
	limit := args.First
	if limit == 0 {
		limit = 10
	}

	query := "SELECT * FROM account_transactions WHERE "
	conds := make([]string, 0)
	conds = append(conds, "account_id=:accountid")
	if args.After != "" {
		conds = append(
			conds,
			// equality needs to flip when ORDERING BY ASC
			`created_at < (
				SELECT created_at FROM account_transactions WHERE account_id=:accountid AND id=:id LIMIT 1
			)`,
		)
	}

	query = query + strings.Join(conds, " AND ")
	stmt, err := b.DB().PrepareNamedContext(ctx, query+" ORDER BY created_at DESC LIMIT :limit;")
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
	}

	transactions := []accountTransaction{}
	err = stmt.SelectContext(
		ctx, &transactions,
		map[string]interface{}{
			"accountid": args.AccountID,
			"id":        args.After,
			"limit":     limit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
	}

	ret := make([]account_transactions.AccountTransaction, len(transactions))
	for i, trx := range transactions {
		ret[i] = account_transactions.AccountTransaction{
			ID:          trx.ID,
			Type:        trx.Type,
			AccountID:   trx.AccountID,
			Description: trx.Description,
			State:       trx.State,
			NetAmount:   trx.NetAmount,
			TransferIDs: trx.TransferIDs,
			CreatedAt:   trx.CreatedAt,
			UpdatedAt:   trx.UpdatedAt,
		}
	}

	return ret, nil
}

func GetPageInfo(
	ctx context.Context,
	b Backends,
	accountID string,
	edges []account_transactions.AccountTransaction,
) (hasNextPage bool, startCursor string, endCursor string, err error) {
	if len(edges) == 0 {
		return false, "", "", nil
	}

	last := edges[len(edges)-1]
	nextPageEdges, err := GetPage(ctx, b, &account_transactions.PaginationArgs{
		AccountID: accountID,
		After:     last.ID,
	})
	if err != nil {
		return false, "", "", err
	}

	return len(nextPageEdges) > 0, edges[0].ID, last.ID, nil
}

func Get(ctx context.Context, b Backends, id string) (*account_transactions.AccountTransaction, error) {
	trx := &accountTransaction{}
	err := crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		err := b.DB().GetContext(
			ctx,
			trx,
			"SELECT * FROM account_transactions WHERE id=$1 LIMIT 1;",
			id,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return account_transactions.ErrNotFound
			}

			return fmt.Errorf("%w %s", account_transactions.ErrInternal, err.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &account_transactions.AccountTransaction{
		ID:          trx.ID,
		Type:        trx.Type,
		AccountID:   trx.AccountID,
		Description: trx.Description,
		State:       trx.State,
		NetAmount:   trx.NetAmount,
		TransferIDs: trx.TransferIDs,
		CreatedAt:   trx.CreatedAt,
		UpdatedAt:   trx.UpdatedAt,
	}, nil
}
