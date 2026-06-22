package seed

import (
	"context"
	"fmt"
	"os"

	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/interledger/interledger-app/go/pacioli/ledger"
	"gopkg.in/yaml.v2"
)

func Seed(b ledger.Backends, filePath string) error {
	ctx := context.Background()
	confRaw, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = ledgers(ctx, b, confRaw)
	if err != nil {
		return err
	}

	err = accounts(ctx, b, confRaw)
	if err != nil {
		return err
	}

	return nil
}

type Ledgers struct {
	Ledgers []pacioli.ConfigureLedgerArgs `yaml:"ledgers"`
}

func ledgers(ctx context.Context, b ledger.Backends, confRaw []byte) error {
	var conf Ledgers
	err := yaml.Unmarshal(confRaw, &conf)
	if err != nil {
		return err
	}

	el, err := ledger.ConfigureLedgers(ctx, b, conf.Ledgers)
	if err != nil {
		return err
	}

	for _, el := range el {
		return fmt.Errorf("error at index:%d code:%d", el.Index, el.Code)
	}

	ids := make([]uint32, len(conf.Ledgers))
	for i, lc := range conf.Ledgers {
		ids[i] = lc.ID
	}

	ll, err := ledger.GetLedgers(ctx, b, ids)
	if err != nil {
		return err
	}

	for _, lc := range conf.Ledgers {
		var found, match bool
		for _, al := range ll {
			if lc.ID != al.ID {
				continue
			}
			found = true

			if lc.Name == al.Name &&
				lc.Asset == al.Asset &&
				lc.Scale == al.Scale {
				match = true
			}
			break
		}

		// The ledger is created and the config matches.
		if found && match {
			continue
		}

		return fmt.Errorf("ledger not created or does not match config. ledgerID: %d, found:%t, match:%t", lc.ID, found, match)
	}

	return nil
}

type AccountConf struct {
	Accounts []Accounts `yaml:"accounts"`
}

type Accounts struct {
	ID                         string `yaml:"id"`
	LedgerID                   uint32 `yaml:"ledger_id"`
	Code                       uint16 `yaml:"code"`
	Linked                     bool   `yaml:"linked"`
	DebitsMustNotExceedCredits bool   `yaml:"debits_must_not_exceed_credits"`
	CreditsMustNotExceedDebits bool   `yaml:"credits_must_not_exceed_debits"`
}

func accounts(ctx context.Context, b ledger.Backends, confRaw []byte) error {
	var conf AccountConf
	err := yaml.Unmarshal(confRaw, &conf)
	if err != nil {
		return err
	}

	accArgs := make([]pacioli.ConfigureAccountArgs, len(conf.Accounts))
	accIds := make([]string, len(conf.Accounts))
	for i, accConf := range conf.Accounts {
		accArgs[i] = pacioli.ConfigureAccountArgs{
			ID:                         accConf.ID,
			LedgerID:                   accConf.LedgerID,
			Code:                       accConf.Code,
			DebitsMustNotExceedCredits: accConf.DebitsMustNotExceedCredits,
			CreditsMustNotExceedDebits: accConf.CreditsMustNotExceedDebits,
		}
		accIds[i] = accConf.ID
	}

	el, err := ledger.ConfigureAccounts(ctx, b, accArgs)
	if err != nil {
		return err
	}

	for _, el := range el {
		switch el.Code {
		case pacioli.AccountExistsWithDifferentDebitsPending,
			pacioli.AccountExistsWithDifferentDebitsPosted,
			pacioli.AccountExistsWithDifferentCreditsPending,
			pacioli.AccountExistsWithDifferentCreditsPosted:
			continue
		}
		return fmt.Errorf("error at index:%d code:%d", el.Index, el.Code)
	}

	// Check that accounts created successfully.
	accs, err := ledger.GetAccounts(ctx, b, accIds)
	if err != nil {
		return err
	}

	for _, acc := range accs {
		var found, match bool
		for _, accConf := range conf.Accounts {
			if accConf.ID != acc.ID {
				continue
			}
			found = true

			if acc.LedgerID == accConf.LedgerID &&
				acc.Code == accConf.Code &&
				acc.DebitsMustNotExceedCredits == accConf.DebitsMustNotExceedCredits &&
				acc.CreditsMustNotExceedDebits == accConf.CreditsMustNotExceedDebits {
				match = true
			}
			break
		}

		// The account is created and the config matches.
		if found && match {
			continue
		}

		return fmt.Errorf("account not created or does not match config. accID: %s, found:%t, match:%t, actual:%+v", acc.ID, found, match, acc)
	}

	return nil
}
