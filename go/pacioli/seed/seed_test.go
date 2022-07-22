package seed_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/pacioli/ledger"
	"gitlab.com/fynbos/pacioli/ledger/mock"
	"gitlab.com/fynbos/pacioli/seed"
)

func TestTigerBeetle(t *testing.T) {
	ctrl := gomock.NewController(t)
	lc := mock.NewMockService(ctrl)

	lc.EXPECT().ConfigureLedgers(gomock.Any(), []ledger.ConfigureLedgerArgs{
		{
			ID:    123,
			Name:  "LocalUSD",
			Asset: "USD",
			Scale: 2,
		},
		{
			ID:    124,
			Name:  "LocalZAR",
			Asset: "ZAR",
			Scale: 2,
		},
	}).Times(1)
	lc.EXPECT().ConfigureAccounts(gomock.Any(), []ledger.ConfigureAccountArgs{
		{
			ID:       "46d4a2bd-e29b-4a63-9aa8-7990776c714e",
			LedgerID: 124,
			Code:     2,
			Flags: ledger.AccountFlags{
				DebitsMustNotExceedCredits: true,
			},
		},
		{
			ID:       "c54aa8a9-b303-4b75-9bf4-203a9cf15f68",
			LedgerID: 123,
			Code:     2,
			Flags: ledger.AccountFlags{
				CreditsMustNotExceedDebits: true,
			},
		},
		{
			ID:       "29e5aa54-0dc8-4e92-a9dd-b99a373525f0",
			LedgerID: 123,
			Code:     2,
			Flags: ledger.AccountFlags{
				Linked:                     true,
				DebitsMustNotExceedCredits: true,
				CreditsMustNotExceedDebits: true,
			},
		},
	}).Times(1)

	err := seed.TigerBeetle(lc, "example.yml")
	assert.NoError(t, err)
}
