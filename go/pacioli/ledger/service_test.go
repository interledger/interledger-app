package ledger

import (
	"context"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	tb_types "gitlab.com/fynbos/tigerbeetle_go/pkg/types"
)

func TestPacioli(s *testing.T) {
	ctx := context.Background()
	c, err := NewTestContainer(ctx, s)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		err = c.Cleanup()
		if err != nil {
			s.Fatal(err)
		}
	})

	s.Run("creating tenants is idempotent", func(t *testing.T) {
		tenant := faker.Name()
		err := c.Ls.ConfigureTenant(ctx, tenant)
		assert.Nil(t, err)

		err = c.Ls.ConfigureTenant(ctx, tenant)
		assert.Nil(t, err)
	})

	s.Run("creating ledgers is idempotent", func(t *testing.T) {
		tenant := faker.Name()
		err := c.Ls.ConfigureTenant(ctx, tenant)
		assert.Nil(t, err)

		ledgerID := uint16(0)
		name := faker.Name()
		asset := "840"
		scale := uint8(2)
		ledger2ID := ledgerID + 1

		args := []ConfigureLedgerArgs{
			{
				ID:    ledgerID,
				Name:  name,
				Asset: asset,
				Scale: scale,
			},
			{
				ID:    ledgerID,
				Name:  faker.Name(), // this will fail because the name is different
				Asset: asset,
				Scale: scale,
			},
			{
				ID:    ledgerID,
				Name:  name,
				Asset: "740", // this will fail because the asset is different
				Scale: scale,
			},
			{
				ID:    ledgerID,
				Name:  name,
				Asset: asset,
				Scale: 3, // this will fail because the scale is different
			},
			{
				ID:    ledger2ID, // this will succeed because the id is different
				Name:  name,
				Asset: asset,
				Scale: scale,
			},
		}

		eventErrors, err := c.Ls.ConfigureLedgers(ctx, tenant, args)
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, eventErrors, 3)
		assert.Equal(t, eventErrors[0].Index, uint32(1))
		assert.Equal(t, eventErrors[0].Code, uint32(LEDGER_EXISTS_WITH_DIFFERENT_NAME))
		assert.Equal(t, eventErrors[1].Index, uint32(2))
		assert.Equal(t, eventErrors[1].Code, uint32(LEDGER_EXISTS_WITH_DIFFERENT_ASSET))
		assert.Equal(t, eventErrors[2].Index, uint32(3))
		assert.Equal(t, eventErrors[2].Code, uint32(LEDGER_EXISTS_WITH_DIFFERENT_SCALE))

		ledgers, err := c.Ls.GetLedgers(ctx, tenant, []uint32{uint32(ledgerID), uint32(ledger2ID)})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, ledgers, 2)
		assert.Equal(t, ledgers[0].Name, name)
		assert.Equal(t, ledgers[0].Asset, asset)
		assert.Equal(t, ledgers[0].Scale, scale)
		assert.Equal(t, ledgers[1].Name, name)
		assert.Equal(t, ledgers[1].Asset, asset)
		assert.Equal(t, ledgers[1].Scale, scale)
	})

	s.Run("creating accounts is idempotent", func(t *testing.T) {
		tenant := faker.Name()
		err := c.Ls.ConfigureTenant(ctx, tenant)
		if err != nil {
			t.Fatal(err)
		}
		ledgerID := uint16(2)
		eventErrors, err := c.Ls.ConfigureLedgers(ctx, tenant, []ConfigureLedgerArgs{{
			ID:    ledgerID,
			Name:  faker.Name(), // this will fail because the name is different
			Asset: "840",
			Scale: 2,
		}})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, eventErrors, 0)

		account1ID := uuid.NewString()
		account2ID := uuid.NewString()
		args := []ConfigureAccountArgs{
			{
				ID:       account1ID,
				LedgerID: ledgerID,
			},
			// we're not exhaustively testing the fields as this is a pass through to TB. We
			// just want to check that errors are correctly passed and that accounts are created
			// if ledger exists.
			{
				ID:       account1ID,
				LedgerID: ledgerID,
				Code:     1, // this will fail because the code is different
			},
			{
				ID:       account2ID,
				LedgerID: 3, // this will fail because the ledger doesn't exist
			},
			{
				ID:       account2ID,
				LedgerID: ledgerID,
				Code:     1,
				Flags: AccountFlags{
					DebitsMustNotExceedCredits: true,
				},
			},
		}
		eventErrors, err = c.Ls.ConfigureAccounts(ctx, tenant, args)
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, eventErrors, 2)
		// the error events may not correspond to the order the []CreateAccountArgs
		for _, err := range eventErrors {
			switch err.Code {
			case uint32(tb_types.AccountExistsWithDifferentCode):
				assert.Equal(t, uint32(1), err.Index, "The create account error mapping is broken.")
			case uint32(ACCOUNT_LEDGER_DOES_NOT_EXIST):
				assert.Equal(t, uint32(2), err.Index, "The create account error mapping is broken.")
			default:
				t.Fatal("The error mapping is broken.")
			}
		}

		accounts, err := c.Ls.GetAccounts(ctx, tenant, []string{account1ID, account2ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, accounts, 2)
		assert.Equal(t, accounts[0].ID, account1ID)
		assert.Equal(t, accounts[0].Code, uint16(0))
		assert.Equal(t, accounts[0].LedgerID, ledgerID)
		assert.Equal(t, accounts[0].DebitsAccepted, uint64(0))
		assert.Equal(t, accounts[0].DebitsReserved, uint64(0))
		assert.Equal(t, accounts[0].CreditsAccepted, uint64(0))
		assert.Equal(t, accounts[0].CreditsReserved, uint64(0))
		assert.Equal(t, accounts[0].Flags, AccountFlags{})
		assert.Equal(t, accounts[1].ID, account2ID)
		assert.Equal(t, accounts[1].Code, uint16(1))
		assert.Equal(t, accounts[1].LedgerID, ledgerID)
		assert.Equal(t, accounts[1].DebitsAccepted, uint64(0))
		assert.Equal(t, accounts[1].DebitsReserved, uint64(0))
		assert.Equal(t, accounts[1].CreditsAccepted, uint64(0))
		assert.Equal(t, accounts[1].CreditsReserved, uint64(0))
		assert.Equal(t, AccountFlags{DebitsMustNotExceedCredits: true}, accounts[1].Flags)
	})

	s.Run("creating transfers is idempotent", func(t *testing.T) {
		tenant := faker.Name()
		err := c.Ls.ConfigureTenant(ctx, tenant)
		if err != nil {
			t.Fatal(err)
		}
		ledgerID := uint16(3)
		eventErrors, err := c.Ls.ConfigureLedgers(ctx, tenant, []ConfigureLedgerArgs{{
			ID:    ledgerID,
			Name:  faker.Name(), // this will fail because the name is different
			Asset: "840",
			Scale: 2,
		}})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, eventErrors, 0)
		accountA := ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
		}
		accountB := ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
		}
		eventErrors, err = c.Ls.ConfigureAccounts(ctx, tenant, []ConfigureAccountArgs{
			accountA,
			accountB,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, eventErrors, 0)

		transfer1ID := uuid.NewString()
		createTransfers := []CreateTransferArgs{
			{
				ID:              transfer1ID,
				Amount:          10,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
			},
			{ // this will fail as ID already exists
				ID:              transfer1ID,
				Amount:          11,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
			},
			{
				ID:              uuid.NewString(),
				Amount:          13,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
				Flags: TransferFlags{
					TwoPhaseCommit: true,
				},
				Timeout: uint64(10 * time.Microsecond),
			},
		}

		eventErrors, err = c.Ls.CreateTransfers(ctx, tenant, createTransfers)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, eventErrors, 1)
		assert.Equal(t, eventErrors[0].Index, uint32(1))
		assert.Equal(t, eventErrors[0].Code, uint32(tb_types.TransferExistsWithDifferentAmount))

		accounts, err := c.Ls.GetAccounts(ctx, tenant, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, accounts, 2)
		assert.Equal(t, accounts[0].ID, accountA.ID)
		assert.Equal(t, accounts[0].DebitsAccepted, uint64(10))
		assert.Equal(t, accounts[0].DebitsReserved, uint64(13))
		assert.Equal(t, accounts[1].ID, accountB.ID)
		assert.Equal(t, accounts[1].CreditsAccepted, uint64(10))
		assert.Equal(t, accounts[1].CreditsReserved, uint64(13))
	})
}
