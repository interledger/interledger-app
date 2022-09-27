package ops

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/fakecash"
	"gitlab.com/fynbos/pacioli"
)

func Create(
	ctx context.Context, b Backends, args fakecash.CreateArgs,
) (*fakecash.Account, error) {
	id := args.ID
	if id == "" {
		id = uuid.NewString()
	}

	ledgerAccErrors, err := b.Pacioli().ConfigureAccounts(ctx, []pacioli.ConfigureAccountArgs{
		{
			ID:       id,
			Code:     1,
			LedgerID: b.LedgerID(),
			Flags: pacioli.AccountFlags{
				CreditsMustNotExceedDebits: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", fakecash.ErrInternal, err)
	}
	if len(ledgerAccErrors) > 0 {
		return nil, fmt.Errorf("%w %+v", fakecash.ErrInternal, ledgerAccErrors)
	}

	return &fakecash.Account{ID: args.ID}, nil
}

func Get(ctx context.Context, b Backends, id string) (*fakecash.Account, error) {
	accounts, err := b.Pacioli().GetAccounts(ctx, []string{id})
	if err != nil {
		return nil, fmt.Errorf("%w %s", fakecash.ErrInternal, err)
	}
	if len(accounts) < 1 {
		return nil, fakecash.ErrNotFound
	}

	return &fakecash.Account{
		ID:               accounts[0].ID,
		AvailableBalance: calculateAvailbleBalance(accounts[0]),
	}, nil
}

func calculateAvailbleBalance(account pacioli.Account) uint64 {
	// not checking overflows as credits must not exceed debits.
	return account.DebitsPosted - (account.CreditsPending + account.CreditsPosted)
}
