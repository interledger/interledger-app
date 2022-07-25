package ledger_test

import (
	"context"
	"testing"
	"time"

	"gitlab.com/fynbos/pacioli/ledger"

	"github.com/bxcodec/faker/v3"
	tb_types "github.com/coilhq/tigerbeetle-go/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

	s.Run("creating ledgers is idempotent", func(t *testing.T) {

		ledgerID := uint32(0)
		name := faker.Name()
		asset := "840"
		scale := uint8(2)
		ledger2ID := ledgerID + 1

		args := []ledger.ConfigureLedgerArgs{
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

		eventErrors, err := c.Ls.ConfigureLedgers(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, eventErrors, 3)
		assert.Equal(t, eventErrors[0].Index, uint32(1))
		assert.Equal(t, eventErrors[0].Code, uint32(ledger.LEDGER_EXISTS_WITH_DIFFERENT_NAME))
		assert.Equal(t, eventErrors[1].Index, uint32(2))
		assert.Equal(t, eventErrors[1].Code, uint32(ledger.LEDGER_EXISTS_WITH_DIFFERENT_ASSET))
		assert.Equal(t, eventErrors[2].Index, uint32(3))
		assert.Equal(t, eventErrors[2].Code, uint32(ledger.LEDGER_EXISTS_WITH_DIFFERENT_SCALE))

		ledgers, err := c.Ls.GetLedgers(ctx, []uint32{uint32(ledgerID), uint32(ledger2ID)})
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
		ledgerID := uint32(2)
		confLedgerErrs, err := c.Ls.ConfigureLedgers(ctx, []ledger.ConfigureLedgerArgs{{
			ID:    ledgerID,
			Name:  faker.Name(), // this will fail because the name is different
			Asset: "840",
			Scale: 2,
		}})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, confLedgerErrs, 0)

		account1ID := uuid.NewString()
		account2ID := uuid.NewString()
		args := []ledger.ConfigureAccountArgs{
			{
				ID:       account1ID,
				LedgerID: ledgerID,
				Code:     1,
			},
			// we're not exhaustively testing the fields as this is a pass through to TB. We
			// just want to check that errors are correctly passed and that accounts are created
			// if ledger exists.
			{
				ID:       account1ID,
				LedgerID: ledgerID,
				Code:     2, // this will fail because the code is different
			},
			{
				ID:       account2ID,
				LedgerID: 3, // this will fail because the ledger doesn't exist
				Code:     1,
			},
			{
				ID:       account2ID,
				LedgerID: ledgerID,
				Code:     1,
				Flags: ledger.AccountFlags{
					DebitsMustNotExceedCredits: true,
				},
			},
			{ // this shouldn't fail as it will exist
				ID:       account1ID,
				LedgerID: ledgerID,
				Code:     1,
			},
		}
		confAccountErrs, err := c.Ls.ConfigureAccounts(ctx, args)
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confAccountErrs, 2)
		// the error events may not correspond to the order the []CreateAccountArgs
		for _, err := range confAccountErrs {
			switch err.Code {
			case tb_types.AccountExistsWithDifferentCode:
				assert.Equal(t, uint32(1), err.Index, "The create account error mapping is broken.")
			case tb_types.CreateAccountResult(ledger.ACCOUNT_LEDGER_DOES_NOT_EXIST):
				assert.Equal(t, uint32(2), err.Index, "The create account error mapping is broken.")
			default:
				t.Fatal("The error mapping is broken.")
			}
		}
		accounts, err := c.Ls.GetAccounts(ctx, []string{account1ID, account2ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, accounts, 2)
		assert.Equal(t, accounts[0].ID, account1ID)
		assert.Equal(t, accounts[0].Code, uint16(1))
		assert.Equal(t, accounts[0].LedgerID, ledgerID)
		assert.Equal(t, accounts[0].DebitsPosted, uint64(0))
		assert.Equal(t, accounts[0].DebitsPending, uint64(0))
		assert.Equal(t, accounts[0].CreditsPosted, uint64(0))
		assert.Equal(t, accounts[0].CreditsPending, uint64(0))
		assert.Equal(t, accounts[0].Flags, ledger.AccountFlags{})
		assert.Equal(t, accounts[1].ID, account2ID)
		assert.Equal(t, accounts[1].Code, uint16(1))
		assert.Equal(t, accounts[1].LedgerID, ledgerID)
		assert.Equal(t, accounts[1].DebitsPosted, uint64(0))
		assert.Equal(t, accounts[1].DebitsPending, uint64(0))
		assert.Equal(t, accounts[1].CreditsPosted, uint64(0))
		assert.Equal(t, accounts[1].CreditsPending, uint64(0))
		assert.Equal(t, ledger.AccountFlags{DebitsMustNotExceedCredits: true}, accounts[1].Flags)
	})

	s.Run("creating transfers is idempotent", func(t *testing.T) {
		ledgerID := uint32(3)
		confLedgerErrs, err := c.Ls.ConfigureLedgers(ctx, []ledger.ConfigureLedgerArgs{{
			ID:    ledgerID,
			Name:  faker.Name(), // this will fail because the name is different
			Asset: "840",
			Scale: 2,
		}})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confLedgerErrs, 0)
		accountA := ledger.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}
		accountB := ledger.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}

		confAccountErrs, err := c.Ls.ConfigureAccounts(ctx, []ledger.ConfigureAccountArgs{
			accountA,
			accountB,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confAccountErrs, 0)

		transfer1ID := uuid.NewString()
		createTransfers := []ledger.CreateTransferArgs{
			{
				ID:              transfer1ID,
				Amount:          10,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
				Ledger:          ledgerID,
				Code:            1,
			},
			{ // this will fail as ID already exists
				ID:              transfer1ID,
				Amount:          11,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
				Ledger:          ledgerID,
				Code:            1,
			},
			{
				ID:              uuid.NewString(),
				Amount:          13,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
				Flags: ledger.TransferFlags{
					Pending: true,
				},
				Timeout: uint64(10 * time.Microsecond),
				Ledger:  ledgerID,
				Code:    1,
			},
		}

		createTransErrs, err := c.Ls.CreateTransfers(ctx, createTransfers)
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, createTransErrs, 1)
		assert.Equal(t, createTransErrs[0].Index, uint32(1))
		assert.Equal(t, createTransErrs[0].Code, tb_types.TransferExistsWithDifferentAmount)

		accounts, err := c.Ls.GetAccounts(ctx, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, accounts, 2)
		assert.Equal(t, accounts[0].ID, accountA.ID)
		assert.Equal(t, accounts[0].DebitsPosted, uint64(10))
		assert.Equal(t, accounts[0].DebitsPending, uint64(13))
		assert.Equal(t, accounts[1].ID, accountB.ID)
		assert.Equal(t, accounts[1].CreditsPosted, uint64(10))
		assert.Equal(t, accounts[1].CreditsPending, uint64(13))
	})

	s.Run("transfer commit is idempotent", func(t *testing.T) {
		ledgerID := uint32(4)
		confLedgerErrs, err := c.Ls.ConfigureLedgers(ctx, []ledger.ConfigureLedgerArgs{{
			ID:    ledgerID,
			Name:  faker.Name(), // this will fail because the name is different
			Asset: "840",
			Scale: 2,
		}})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, confLedgerErrs, 0)

		accountA := ledger.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}
		accountB := ledger.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}

		confAccErrs, err := c.Ls.ConfigureAccounts(ctx, []ledger.ConfigureAccountArgs{
			accountA,
			accountB,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confAccErrs, 0)

		transfer1ID := uuid.NewString()
		createTransfers := []ledger.CreateTransferArgs{
			{
				ID:              transfer1ID,
				Amount:          13,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
				Flags: ledger.TransferFlags{
					Pending: true,
				},
				Timeout: uint64(time.Second),
				Ledger:  ledgerID,
				Code:    1,
			},
		}

		transErrs, err := c.Ls.CreateTransfers(ctx, createTransfers)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, transErrs, 0)

		accounts, err := c.Ls.GetAccounts(ctx, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, accounts, 2)
		assert.Equal(t, accountA.ID, accounts[0].ID)
		assert.Equal(t, uint64(0), accounts[0].DebitsPosted)
		assert.Equal(t, uint64(13), accounts[0].DebitsPending)
		assert.Equal(t, accountB.ID, accounts[1].ID)
		assert.Equal(t, uint64(0), accounts[1].CreditsPosted)
		assert.Equal(t, uint64(13), accounts[1].CreditsPending)

		erList, err := c.Ls.CommitTransfers(ctx, []string{transfer1ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, erList, 0)

		// Check that the commit went through.
		accounts, err = c.Ls.GetAccounts(ctx, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, accounts, 2)
		assert.Equal(t, accounts[0].ID, accountA.ID)
		assert.Equal(t, accounts[0].DebitsPosted, uint64(13))
		assert.Equal(t, accounts[0].DebitsPending, uint64(0))
		assert.Equal(t, accounts[1].ID, accountB.ID)
		assert.Equal(t, accounts[1].CreditsPosted, uint64(13))
		assert.Equal(t, accounts[1].CreditsPending, uint64(0))

		// Commit again
		erList, err = c.Ls.CommitTransfers(ctx, []string{transfer1ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, erList, 1)
		assert.Equal(t, erList[0].Index, uint32(0))
		assert.Equal(t, erList[0].Code, tb_types.TransferPendingTransferAlreadyPosted)
	})

	s.Run("transfer void is idempotent", func(t *testing.T) {
		ledgerID := uint32(5)
		confLedgerErrs, err := c.Ls.ConfigureLedgers(ctx, []ledger.ConfigureLedgerArgs{{
			ID:    ledgerID,
			Name:  faker.Name(), // this will fail because the name is different
			Asset: "840",
			Scale: 2,
		}})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confLedgerErrs, 0)
		accountA := ledger.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}
		accountB := ledger.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}

		confAccErrs, err := c.Ls.ConfigureAccounts(ctx, []ledger.ConfigureAccountArgs{
			accountA,
			accountB,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confAccErrs, 0)

		transfer1ID := uuid.NewString()
		createTransfers := []ledger.CreateTransferArgs{
			{
				ID:              transfer1ID,
				Amount:          13,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
				Flags: ledger.TransferFlags{
					Pending: true,
				},
				Timeout: uint64(time.Second),
				Ledger:  ledgerID,
				Code:    1,
			},
		}

		transErrs, err := c.Ls.CreateTransfers(ctx, createTransfers)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, transErrs, 0)

		accounts, err := c.Ls.GetAccounts(ctx, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, accounts, 2)
		assert.Equal(t, accountA.ID, accounts[0].ID)
		assert.Equal(t, uint64(0), accounts[0].DebitsPosted)
		assert.Equal(t, uint64(13), accounts[0].DebitsPending)
		assert.Equal(t, accountB.ID, accounts[1].ID)
		assert.Equal(t, uint64(0), accounts[1].CreditsPosted)
		assert.Equal(t, uint64(13), accounts[1].CreditsPending)

		erList, err := c.Ls.VoidTransfers(ctx, []string{transfer1ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, erList, 0)

		// Check that the void went through.
		accounts, err = c.Ls.GetAccounts(ctx, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, accounts, 2)
		assert.Equal(t, accounts[0].ID, accountA.ID)
		assert.Equal(t, accounts[0].DebitsPosted, uint64(0))
		assert.Equal(t, accounts[0].DebitsPending, uint64(0))
		assert.Equal(t, accounts[1].ID, accountB.ID)
		assert.Equal(t, accounts[1].CreditsPosted, uint64(0))
		assert.Equal(t, accounts[1].CreditsPending, uint64(0))

		// Void again
		erList, err = c.Ls.VoidTransfers(ctx, []string{transfer1ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, erList, 1)
		assert.Equal(t, erList[0].Index, uint32(0))
		assert.Equal(t, erList[0].Code, tb_types.TransferPendingTransferAlreadyVoided)
	})
}
