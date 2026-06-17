package seed_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/interledger/interledger-app/go/pacioli/db"
	"github.com/interledger/interledger-app/go/pacioli/ledger"
	"github.com/interledger/interledger-app/go/pacioli/seed"
	test_utils "github.com/interledger/interledger-app/go/pacioli/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeed(t *testing.T) {
	//ctx := context.Background()

	/*ctrl := gomock.NewController(t)
	lc := mock.NewMockService(ctrl)

	c.EXPECT().ConfigureLedgers(gomock.Any(), []pacioli.ConfigureLedgerArgs{
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
	lc.EXPECT().ConfigureAccounts(gomock.Any(), []pacioli.ConfigureAccountArgs{
		{
			ID:       "46d4a2bd-e29b-4a63-9aa8-7990776c714e",
			LedgerID: 124,
			Code:     2,
			Flags: pacioli.AccountFlags{
				DebitsMustNotExceedCredits: true,
			},
		},
		{
			ID:       "c54aa8a9-b303-4b75-9bf4-203a9cf15f68",
			LedgerID: 123,
			Code:     2,
			Flags: pacioli.AccountFlags{
				CreditsMustNotExceedDebits: true,
			},
		},
		{
			ID:       "29e5aa54-0dc8-4e92-a9dd-b99a373525f0",
			LedgerID: 123,
			Code:     2,
			Flags: pacioli.AccountFlags{
				Linked:                     true,
				DebitsMustNotExceedCredits: true,
				CreditsMustNotExceedDebits: true,
			},
		},
	}).Times(1)
	lc.EXPECT().GetLedgers(gomock.Any(), gomock.Any()).Return([]pacioli.Ledger{
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
	}, nil).Times(1)
	lc.EXPECT().GetAccounts(gomock.Any(), gomock.Any()).Return([]pacioli.Account{
		{
			ID:       "46d4a2bd-e29b-4a63-9aa8-7990776c714e",
			LedgerID: 124,
			Code:     2,
			Flags: pacioli.AccountFlags{
				DebitsMustNotExceedCredits: true,
			},
		},
		{
			ID:       "c54aa8a9-b303-4b75-9bf4-203a9cf15f68",
			LedgerID: 123,
			Code:     2,
			Flags: pacioli.AccountFlags{
				CreditsMustNotExceedDebits: true,
			},
		},
		{
			ID:       "29e5aa54-0dc8-4e92-a9dd-b99a373525f0",
			LedgerID: 123,
			Code:     2,
			Flags: pacioli.AccountFlags{
				Linked:                     true,
				DebitsMustNotExceedCredits: true,
				CreditsMustNotExceedDebits: true,
			},
		},
	}, nil).Times(1)
	*/
	ctx := context.Background()
	c, err := NewTestContainer(ctx, t)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = c.Cleanup()
		require.NoError(t, err)
	})

	//return test_utils.NewBackends(t, dbc, tbClient), nil
	fmt.Println("XXXXXXX")
	err = seed.Seed(c.b, "example.yml")
	fmt.Println("YYYYYYYYYYYY")
	assert.NoError(t, err)
}

type TestContainer struct {
	b   ledger.Backends
	Ctx context.Context
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx

	_, db := db.MigrateTestDB(t, ctx)

	c.b = test_utils.NewBackends(t, db)
	return c, nil
}

func (c *TestContainer) Cleanup() error {
	return nil
}
