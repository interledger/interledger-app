package tigerroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	tb_types "github.com/coilhq/tigerbeetle-go/pkg/types"
	"gitlab.com/fynbos/pacioli"
)

type ledgerAccount struct {
	pacioli.Account
	FlagsRaw  uint16    `db:"flags"`
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
			return nil, fmt.Errorf("%s %w", err.Error(), pacioli.ErrInvalidArg)
		}

		// Return an error here as Tigerbeetle doesn't have any validation for this case so
		// can't return an AccountResult with a code that makes sense.
		_, err = GetLedger(ctx, b, aa.LedgerID)
		if errors.Is(err, pacioli.ErrNotFound) {
			return nil, fmt.Errorf("%s %d %s %w", "unknown ledger index: ", i, err.Error(), pacioli.ErrNotFound)
		} else if err != nil {
			return nil, fmt.Errorf("%s %d %s %w", "index: ", i, err.Error(), pacioli.ErrInternal)
		}

		if aa.Flags.DebitsMustNotExceedCredits && aa.Flags.CreditsMustNotExceedDebits {
			resMap[i] = pacioli.AccountResult{
				Index: uint32(i),
				Code:  tb_types.AccountMutuallyExclusiveFlags,
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
			return nil, fmt.Errorf("%s %d %s %w", "index: ", i, err.Error(), pacioli.ErrInternal)
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
		return 0, err
	}
	if ex != nil {
		if ex.Flags.ToUint16() != args.Flags.ToUint16() {
			return tb_types.AccountExistsWithDifferentFlags, nil
		} else if ex.LedgerID != args.LedgerID {
			return tb_types.AccountExistsWithDifferentLedger, nil
		} else if ex.Code != args.Code {
			return tb_types.AccountExistsWithDifferentCode, nil
		}

		// Account exists with all the same params, do nothing.
		return pacioli.AccountOK, nil
	}

	_, err = b.DB().ExecContext(ctx,
		"INSERT INTO ledger_accounts (id, ledger_id, code, flags)VALUES ($1, $2, $3, $4)",
		args.ID, args.LedgerID, args.Code, args.Flags.ToUint16())

	return pacioli.AccountOK, err
}

func GetAccount(ctx context.Context, b Backends, id string) (*pacioli.Account, error) {
	var acc ledgerAccount
	err := b.DB().GetContext(ctx, &acc, "SELECT * FROM ledger_accounts WHERE id=$1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pacioli.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &pacioli.Account{
		ID:             acc.ID,
		LedgerID:       acc.LedgerID,
		Flags:          pacioli.ToAccountFlags(acc.FlagsRaw),
		Code:           acc.Code,
		DebitsPending:  acc.DebitsPending,
		DebitsPosted:   acc.DebitsPosted,
		CreditsPending: acc.CreditsPending,
		CreditsPosted:  acc.CreditsPosted,
	}, nil
}
