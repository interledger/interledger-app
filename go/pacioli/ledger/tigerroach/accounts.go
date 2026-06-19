package tigerroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/jmoiron/sqlx"
)

type ledgerAccount struct {
	pacioli.Account
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func ConfigureAccounts(
	ctx context.Context,
	b Backends,
	args []pacioli.ConfigureAccountArgs,
) ([]pacioli.AccountResult, error) {

	resMap := make(map[int]pacioli.AccountResult)
	for i, aa := range args {
		err := b.Validator().Struct(aa)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err, pacioli.ErrInvalidArg)
		}

		// Return an error here as Tigerbeetle doesn't have any validation for this case so
		// can't return an AccountResult with a code that makes sense.
		_, err = GetLedger(ctx, b, aa.LedgerID)
		if errors.Is(err, pacioli.ErrNotFound) {
			resMap[i] = pacioli.AccountResult{
				Index: uint32(i),
				Code:  pacioli.AccountLedgerDoesNotExist,
			}
		} else if err != nil {
			return nil, fmt.Errorf("%s %d %s %w", "index: ", i, err, pacioli.ErrInternal)
		}

		if aa.DebitsMustNotExceedCredits && aa.CreditsMustNotExceedDebits {
			resMap[i] = pacioli.AccountResult{
				Index: uint32(i),
				Code:  pacioli.AccountMutuallyExclusiveFlags,
			}
		}
	}

	for i, ac := range args {
		if _, ok := resMap[i]; ok {
			// Already failed validation
			continue
		}

		code, err := configureAccount(ctx, b, ac)
		if err != nil {
			return nil, fmt.Errorf("%s %d %s %w", "index: ", i, err, pacioli.ErrInternal)
		}

		if code == pacioli.AccountOK {
			continue
		}

		resMap[i] = pacioli.AccountResult{
			Index: uint32(i),
			Code:  code,
		}
	}

	var res []pacioli.AccountResult
	for _, v := range resMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Index < res[j].Index
	})
	return res, nil
}

func configureAccount(
	ctx context.Context,
	b Backends,
	args pacioli.ConfigureAccountArgs,
) (pacioli.AccountResultCode, error) {

	ex, err := GetAccount(ctx, b, args.ID)
	if err != nil && !errors.Is(err, pacioli.ErrNotFound) {
		return 0, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}
	if ex != nil {
		if ex.LedgerID != args.LedgerID {
			return pacioli.AccountExistsWithDifferentLedger, nil
		} else if ex.Code != args.Code {
			return pacioli.AccountExistsWithDifferentCode, nil
		} else if ex.DebitsMustNotExceedCredits != args.DebitsMustNotExceedCredits ||
			ex.CreditsMustNotExceedDebits != args.CreditsMustNotExceedDebits {
			return pacioli.AccountExistsWithDifferentFlags, nil
		}

		// Account exists with all the same params, do nothing.
		return pacioli.AccountOK, nil
	}

	_, err = GetLedger(ctx, b, args.LedgerID)
	if errors.Is(err, pacioli.ErrNotFound) {
		return pacioli.AccountLedgerDoesNotExist, nil
	}
	if err != nil {
		return pacioli.AccountOK, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	_, err = b.DB().ExecContext(ctx,
		"INSERT INTO ledger_accounts (id, ledger_id, code, debits_must_not_exceed_credits, credits_must_not_exceed_debits)VALUES ($1, $2, $3, $4, $5)",
		args.ID, args.LedgerID, args.Code, args.DebitsMustNotExceedCredits, args.CreditsMustNotExceedDebits)
	if err != nil {
		return pacioli.AccountOK, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	return pacioli.AccountOK, nil
}

func GetAccount(ctx context.Context, b Backends, id string) (*pacioli.Account, error) {
	var acc ledgerAccount
	err := b.DB().GetContext(ctx, &acc, "SELECT * FROM ledger_accounts WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pacioli.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	return &pacioli.Account{
		ID:                         acc.ID,
		LedgerID:                   acc.LedgerID,
		DebitsMustNotExceedCredits: acc.DebitsMustNotExceedCredits,
		CreditsMustNotExceedDebits: acc.CreditsMustNotExceedDebits,
		Code:                       acc.Code,
		DebitsPending:              acc.DebitsPending,
		DebitsPosted:               acc.DebitsPosted,
		CreditsPending:             acc.CreditsPending,
		CreditsPosted:              acc.CreditsPosted,
	}, nil
}

func ListAccounts(ctx context.Context, b Backends, ids []string) ([]pacioli.Account, error) {
	var accs []ledgerAccount
	var resp []pacioli.Account
	query, args, err := sqlx.In("SELECT * FROM ledger_accounts WHERE id IN (?);", ids)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}
	err = b.DB().SelectContext(ctx, &accs, b.DB().Rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err, pacioli.ErrInternal)
	}

	for _, acc := range accs {
		resp = append(resp, pacioli.Account{
			ID:                         acc.ID,
			LedgerID:                   acc.LedgerID,
			DebitsMustNotExceedCredits: acc.DebitsMustNotExceedCredits,
			CreditsMustNotExceedDebits: acc.CreditsMustNotExceedDebits,
			Code:                       acc.Code,
			DebitsPending:              acc.DebitsPending,
			DebitsPosted:               acc.DebitsPosted,
			CreditsPending:             acc.CreditsPending,
			CreditsPosted:              acc.CreditsPosted,
		})
	}
	return resp, nil
}
