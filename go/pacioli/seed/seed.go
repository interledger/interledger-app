package seed

import (
	"context"
	"fmt"
	"io/ioutil"

	"gitlab.com/fynbos/pacioli/ledger"
	"gopkg.in/yaml.v2"
)

func TigerBeetle(l ledger.Service, filePath string) error {
	ctx := context.Background()
	confRaw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = ledgers(ctx, l, confRaw)
	if err != nil {
		return err
	}

	err = accounts(ctx, l, confRaw)
	if err != nil {
		return err
	}

	return nil
}

type Ledgers struct {
	Ledgers []ledger.ConfigureLedgerArgs `yaml:"ledgers"`
}

func ledgers(ctx context.Context, l ledger.Service, confRaw []byte) error {
	var conf Ledgers
	err := yaml.Unmarshal(confRaw, &conf)
	if err != nil {
		return err
	}

	el, err := l.ConfigureLedgers(ctx, conf.Ledgers)
	if err != nil {
		return err
	}

	for _, el := range el {
		return fmt.Errorf("error at index:%d code:%d", el.Index, el.Code)
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

func accounts(ctx context.Context, l ledger.Service, confRaw []byte) error {
	var conf AccountConf
	err := yaml.Unmarshal(confRaw, &conf)
	if err != nil {
		return err
	}

	accArgs := make([]ledger.ConfigureAccountArgs, len(conf.Accounts))
	for i, accConf := range conf.Accounts {
		accArgs[i] = ledger.ConfigureAccountArgs{
			ID:       accConf.ID,
			LedgerID: accConf.LedgerID,
			Code:     accConf.Code,
			Flags: ledger.AccountFlags{
				Linked:                     accConf.Linked,
				DebitsMustNotExceedCredits: accConf.DebitsMustNotExceedCredits,
				CreditsMustNotExceedDebits: accConf.CreditsMustNotExceedDebits,
			},
		}
	}

	el, err := l.ConfigureAccounts(ctx, accArgs)
	if err != nil {
		return err
	}

	for _, el := range el {
		return fmt.Errorf("error at index:%d code:%d", el.Index, el.Code)
	}
	return nil
}
