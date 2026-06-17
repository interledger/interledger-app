package ledger_test

import (
	"context"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/pacioli"
	"github.com/interledger/interledger-app/go/pacioli/ledger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		ledgerID := uint32(1)
		name := faker.Name()
		asset := "840"
		scale := uint8(2)
		ledger2ID := ledgerID + 1

		args := []pacioli.ConfigureLedgerArgs{
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

		eventErrors, err := ledger.ConfigureLedgers(ctx, c.b, args)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, eventErrors, 3)
		assert.Equal(t, eventErrors[0].Index, uint32(1))
		assert.Equal(t, eventErrors[0].Code, pacioli.LedgerExistsWithDifferentName)
		assert.Equal(t, eventErrors[1].Index, uint32(2))
		assert.Equal(t, eventErrors[1].Code, pacioli.LedgerExistsWithDifferentAsset)
		assert.Equal(t, eventErrors[2].Index, uint32(3))
		assert.Equal(t, eventErrors[2].Code, pacioli.LedgerExistsWithDifferentScale)

		ledgers, err := ledger.GetLedgers(ctx, c.b, []uint32{ledgerID, ledger2ID})
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
		ledgerID := uint32(4)
		confLedgerErrs, err := ledger.ConfigureLedgers(ctx, c.b, []pacioli.ConfigureLedgerArgs{{
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
		args := []pacioli.ConfigureAccountArgs{
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
				ID:                         account2ID,
				LedgerID:                   ledgerID,
				Code:                       1,
				DebitsMustNotExceedCredits: true,
			},
			{ // this shouldn't fail as it will exist
				ID:       account1ID,
				LedgerID: ledgerID,
				Code:     1,
			},
		}
		confAccountErrs, err := ledger.ConfigureAccounts(ctx, c.b, args)
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confAccountErrs, 2)
		// the error events may not correspond to the order the []CreateAccountArgs
		for _, err := range confAccountErrs {
			switch err.Code {
			case pacioli.AccountExistsWithDifferentCode:
				assert.Equal(t, uint32(1), err.Index, "The create account error mapping is broken.")
			case pacioli.AccountLedgerDoesNotExist:
				assert.Equal(t, uint32(2), err.Index, "The create account error mapping is broken.")
			default:
				t.Fatal("The error mapping is broken.")
			}
		}
		accounts, err := ledger.GetAccounts(ctx, c.b, []string{account1ID, account2ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, accounts, 2)
		for i := range accounts {
			if accounts[i].ID == account1ID {
				assert.Equal(t, accounts[i].ID, account1ID)
				assert.Equal(t, accounts[i].Code, uint16(1))
				assert.Equal(t, accounts[i].LedgerID, ledgerID)
				assert.Equal(t, accounts[i].DebitsPosted, uint64(0))
				assert.Equal(t, accounts[i].DebitsPending, uint64(0))
				assert.Equal(t, accounts[i].CreditsPosted, uint64(0))
				assert.Equal(t, accounts[i].CreditsPending, uint64(0))
			} else if accounts[i].ID == account2ID {
				assert.Equal(t, accounts[i].ID, account2ID)
				assert.Equal(t, accounts[i].Code, uint16(1))
				assert.Equal(t, accounts[i].LedgerID, ledgerID)
				assert.Equal(t, accounts[i].DebitsPosted, uint64(0))
				assert.Equal(t, accounts[i].DebitsPending, uint64(0))
				assert.Equal(t, accounts[i].CreditsPosted, uint64(0))
				assert.Equal(t, accounts[i].CreditsPending, uint64(0))
				assert.True(t, accounts[i].DebitsMustNotExceedCredits)
			} else {
				assert.Fail(t, "unkown account in results")
			}
		}
	})

	s.Run("creating transfers is idempotent", func(t *testing.T) {
		ledgerID := uint32(5)
		confLedgerErrs, err := ledger.ConfigureLedgers(ctx, c.b, []pacioli.ConfigureLedgerArgs{{
			ID:    ledgerID,
			Name:  faker.Name(), // this will fail because the name is different
			Asset: "840",
			Scale: 2,
		}})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confLedgerErrs, 0)
		accountA := pacioli.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}
		accountB := pacioli.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}

		confAccountErrs, err := ledger.ConfigureAccounts(ctx, c.b, []pacioli.ConfigureAccountArgs{
			accountA,
			accountB,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confAccountErrs, 0)

		transfer1ID := uuid.NewString()
		createTransfers := []pacioli.CreateTransferArgs{
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
				Pending:         true,
				Timeout:         uint64(10 * time.Microsecond),
				Ledger:          ledgerID,
				Code:            1,
			},
		}

		createTransErrs, err := ledger.CreateTransfers(ctx, c.b, createTransfers)
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, createTransErrs, 1)
		assert.Equal(t, createTransErrs[0].Index, uint32(1))
		assert.Equal(t, createTransErrs[0].Code, pacioli.TransferExistsWithDifferentAmount)

		accounts, err := ledger.GetAccounts(ctx, c.b, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, accounts, 2)
		for i := range accounts {
			if accounts[i].ID == accountA.ID {
				assert.Equal(t, accounts[i].ID, accountA.ID)
				assert.Equal(t, accounts[i].DebitsPosted, uint64(10))
				assert.Equal(t, accounts[i].DebitsPending, uint64(13))
			} else if accounts[i].ID == accountB.ID {
				assert.Equal(t, accounts[i].ID, accountB.ID)
				assert.Equal(t, accounts[i].CreditsPosted, uint64(10))
				assert.Equal(t, accounts[i].CreditsPending, uint64(13))
			} else {
				assert.Fail(t, "unknown account in result set")
			}
		}
	})

	s.Run("transfer commit is idempotent", func(t *testing.T) {
		ledgerID := uint32(6)
		confLedgerErrs, err := ledger.ConfigureLedgers(ctx, c.b, []pacioli.ConfigureLedgerArgs{{
			ID:    ledgerID,
			Name:  faker.Name(), // this will fail because the name is different
			Asset: "840",
			Scale: 2,
		}})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, confLedgerErrs, 0)

		accountA := pacioli.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}
		accountB := pacioli.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}

		confAccErrs, err := ledger.ConfigureAccounts(ctx, c.b, []pacioli.ConfigureAccountArgs{
			accountA,
			accountB,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confAccErrs, 0)

		transfer1ID := uuid.NewString()
		createTransfers := []pacioli.CreateTransferArgs{
			{
				ID:              transfer1ID,
				Amount:          13,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
				Pending:         true,
				Timeout:         uint64(time.Second),
				Ledger:          ledgerID,
				Code:            1,
			},
		}

		transErrs, err := ledger.CreateTransfers(ctx, c.b, createTransfers)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, transErrs, 0)

		accounts, err := ledger.GetAccounts(ctx, c.b, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, accounts, 2)
		for i := range accounts {
			if accounts[i].ID == accountA.ID {
				assert.Equal(t, uint64(0), accounts[i].DebitsPosted)
				assert.Equal(t, uint64(13), accounts[i].DebitsPending)
			} else if accounts[i].ID == accountB.ID {
				assert.Equal(t, uint64(0), accounts[i].CreditsPosted)
				assert.Equal(t, uint64(13), accounts[i].CreditsPending)
			} else {
				assert.Fail(t, "unknown account in result")
			}
		}

		erList, err := ledger.PostTransfers(ctx, c.b, []string{transfer1ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, erList, 0)

		// Check that the commit went through.
		accounts, err = ledger.GetAccounts(ctx, c.b, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		require.Len(t, accounts, 2)
		for i := range accounts {
			if accounts[i].ID == accountA.ID {
				assert.Equal(t, accounts[i].ID, accountA.ID)
				assert.Equal(t, accounts[i].DebitsPosted, uint64(13))
				assert.Equal(t, accounts[i].DebitsPending, uint64(0))
			} else if accounts[i].ID == accountB.ID {
				assert.Equal(t, accounts[i].ID, accountB.ID)
				assert.Equal(t, accounts[i].CreditsPosted, uint64(13))
				assert.Equal(t, accounts[i].CreditsPending, uint64(0))
			} else {
				assert.Fail(t, "unknown account in result")
			}
		}

		// Commit again
		erList, err = ledger.PostTransfers(ctx, c.b, []string{transfer1ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, erList, 1)
		assert.Equal(t, erList[0].Index, uint32(0))
		assert.Equal(t, erList[0].Code, pacioli.TransferPendingTransferAlreadyPosted)
	})

	s.Run("transfer void is idempotent", func(t *testing.T) {
		ledgerID := uint32(7)
		confLedgerErrs, err := ledger.ConfigureLedgers(ctx, c.b, []pacioli.ConfigureLedgerArgs{{
			ID:    ledgerID,
			Name:  faker.Name(), // this will fail because the name is different
			Asset: "840",
			Scale: 2,
		}})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confLedgerErrs, 0)
		accountA := pacioli.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}
		accountB := pacioli.ConfigureAccountArgs{
			ID:       uuid.NewString(),
			LedgerID: ledgerID,
			Code:     1,
		}

		confAccErrs, err := ledger.ConfigureAccounts(ctx, c.b, []pacioli.ConfigureAccountArgs{
			accountA,
			accountB,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, confAccErrs, 0)

		transfer1ID := uuid.NewString()
		createTransfers := []pacioli.CreateTransferArgs{
			{
				ID:              transfer1ID,
				Amount:          13,
				DebitAccountID:  accountA.ID,
				CreditAccountID: accountB.ID,
				Pending:         true,
				Timeout:         uint64(time.Second),
				Ledger:          ledgerID,
				Code:            1,
			},
		}

		transErrs, err := ledger.CreateTransfers(ctx, c.b, createTransfers)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, transErrs, 0)

		accounts, err := ledger.GetAccounts(ctx, c.b, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		require.Len(t, accounts, 2)
		for i := range accounts {
			if accounts[i].ID == accountA.ID {
				assert.Equal(t, accountA.ID, accounts[i].ID)
				assert.Equal(t, uint64(0), accounts[i].DebitsPosted)
				assert.Equal(t, uint64(13), accounts[i].DebitsPending)
			} else if accounts[i].ID == accountB.ID {
				assert.Equal(t, accountB.ID, accounts[i].ID)
				assert.Equal(t, uint64(0), accounts[i].CreditsPosted)
				assert.Equal(t, uint64(13), accounts[i].CreditsPending)
			} else {
				assert.Fail(t, "unknown account in results")
			}
		}

		erList, err := ledger.VoidTransfers(ctx, c.b, []string{transfer1ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, erList, 0)

		// Check that the void went through.
		accounts, err = ledger.GetAccounts(ctx, c.b, []string{accountA.ID, accountB.ID})
		if err != nil {
			t.Fatal(err)
		}

		require.Len(t, accounts, 2)
		for i := range accounts {
			if accounts[i].ID == accountA.ID {
				assert.Equal(t, accounts[i].ID, accountA.ID)
				assert.Equal(t, accounts[i].DebitsPosted, uint64(0))
				assert.Equal(t, accounts[i].DebitsPending, uint64(0))
			} else if accounts[i].ID == accountB.ID {
				assert.Equal(t, accounts[i].ID, accountB.ID)
				assert.Equal(t, accounts[i].CreditsPosted, uint64(0))
				assert.Equal(t, accounts[i].CreditsPending, uint64(0))
			} else {
				assert.Fail(t, "unknown account in results")
			}
		}

		// Void again
		erList, err = ledger.VoidTransfers(ctx, c.b, []string{transfer1ID})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, erList, 1)
		assert.Equal(t, erList[0].Index, uint32(0))
		assert.Equal(t, erList[0].Code, pacioli.TransferPendingTransferAlreadyVoided)
	})
}
