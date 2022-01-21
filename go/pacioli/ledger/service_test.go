package pacioli

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/pacioli/utils"
	"gitlab.com/fynbos/tigerbeetle_go"
)

func TestLedgerService(s *testing.T) {
	ctx := context.Background()
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		s.Fatal(err)
	}

	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		s.Fatal(err)
	}

	var tbClusterID uint32 = 0
	tb, err := test_utils.SetupTigerBeetle(ctx, tbClusterID)
	if err != nil {
		s.Fatal(err)
	}

	tbClient, err := tigerbeetle_go.NewClient(tbClusterID, []string{tb.URI})
	if err != nil {
		s.Fatal(err)
	}
	// drive the TB client.
	go func() {
		tick := time.Tick(20 * time.Millisecond)
		for range tick {
			tbClient.Tick()
		}
	}()

	ps, err := NewLedgerService(db, tbClient)
	if err != nil {
		s.Fatal(err)
	}

	s.Cleanup(func() {
		// tbClient.Deinit()
		os.RemoveAll(tb.DataDir)
		// tb.Container.Terminate(ctx)

		db.Close()
		crdb.Container.Terminate(ctx)
	})

	s.Run("ledger", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		if err != nil {
			t.Fatal(err)
		}

		t.Run("can create a ledger", func(tt *testing.T) {
			name := faker.Name()
			code := uint16(rand.Intn(65535))
			ledger, err := ps.CreateLedger(name, code)
			if err != nil {
				t.Fatal(err)
			}

			assert.NotNil(tt, ledger.ID)
			assert.Equal(tt, name, ledger.Name)
			assert.Equal(tt, code, ledger.Code)

			fetchedLedger, err := ps.GetLedger(ledger.ID)
			if err != nil {
				tt.Fatal(err)
			}

			assert.Equal(tt, ledger.ID, fetchedLedger.ID)
			assert.Equal(tt, ledger.Name, fetchedLedger.Name)
			assert.Equal(tt, ledger.Code, fetchedLedger.Code)
		})

		t.Run("ledger codes must be unique", func(tt *testing.T) {
			name := faker.Name()
			code := uint16(rand.Intn(65535))
			ledger, err := ps.CreateLedger(name, code)
			if err != nil {
				t.Fatal(err)
			}

			assert.NotNil(tt, ledger.ID)
			assert.Equal(tt, name, ledger.Name)
			assert.Equal(tt, code, ledger.Code)

			dup, err := ps.CreateLedger(name, code)
			if err == nil {
				tt.Fatal("Ledger codes must be unique")
			}

			assert.Nil(tt, dup)
			assert.Equal(tt, ErrInvalidArg{Err: "Ledger Code must be unique."}, err)
		})
	})

	s.Run("can create accounts for a ledger", func(t *testing.T) {
		ledger, err := ps.CreateLedger(faker.Name(), uint16(rand.Intn(65535)))
		if err != nil {
			t.Fatal(err)
		}
		accAArgs := CreateAccountArgs{
			ID:   uuid.NewString(),
			Code: uint16(rand.Intn(65535)),
		}
		accBArgs := CreateAccountArgs{
			ID:   uuid.NewString(),
			Code: uint16(rand.Intn(65535)),
		}

		eventErrors, err := ps.CreateAccounts(ledger.ID, []CreateAccountArgs{accAArgs, accBArgs})
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, eventErrors)

		accounts, err := ps.GetAccounts(ledger.ID, []string{accAArgs.ID, accBArgs.ID})
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, accounts, 2)
		accA := accounts[0]
		assert.Equal(t, accAArgs.ID, accA.ID) // make sure our uuid -> u128 conversion is correct.
		assert.Equal(t, ledger.Code, accA.LedgerCode)
		assert.Equal(t, accAArgs.Code, accA.Code)

		accB := accounts[1]
		assert.Equal(t, accBArgs.ID, accB.ID) // make sure our uuid -> u128 conversion is correct.
		assert.Equal(t, ledger.Code, accB.LedgerCode)
		assert.Equal(t, accBArgs.Code, accB.Code)
	})

	// s.Run("accounts and transfers", func(t *testing.T) {
	// 	t.Cleanup(func() {
	// 		test_utils.TruncateDb(ctx, db)
	// 	})
	// 	ledger, err := ps.CreateLedger(faker.Name())
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	t.Run("can create accounts", func(tt *testing.T) {
	// 		acc, err := ps.CreateAccount(CreateAccountArgs{
	// 			LedgerID: ledger.ID,
	// 			Unit:     1,
	// 		})
	// 		if err != nil {
	// 			tt.Fatal(err)
	// 		}

	// 		assert.Equal(tt, ledger.ID, acc.LedgerID)
	// 		assert.Equal(tt, uint16(1), acc.Unit)
	// 		assert.Equal(tt, uint64(0), acc.DebitsAccepted)
	// 		assert.Equal(tt, uint64(0), acc.DebitsReserved)
	// 		assert.Equal(tt, uint64(0), acc.CreditsAccepted)
	// 		assert.Equal(tt, uint64(0), acc.CreditsReserved)
	// 	})

	// 	t.Run("tenant can create a transfer", func(tt *testing.T) {
	// 		acc1, err := ps.CreateAccount(CreateAccountArgs{
	// 			LedgerID: ledger.ID,
	// 			Unit:     1,
	// 		})
	// 		acc2, err := ps.CreateAccount(CreateAccountArgs{
	// 			LedgerID: ledger.ID,
	// 			Unit:     1,
	// 		})
	// 		if err != nil {
	// 			tt.Fatal(err)
	// 		}
	// 		transfer, err := ps.CreateTransfer(CreateTransferArgs{
	// 			Amount:          100,
	// 			DebitAccountID:  acc1.ID,
	// 			CreditAccountID: acc2.ID,
	// 		})
	// 		if err != nil {
	// 			tt.Fatal(err)
	// 		}
	// 		assert.Equal(tt, uint64(100), transfer.Amount)
	// 		assert.Equal(tt, acc1.ID, transfer.DebitAccountID)
	// 		assert.Equal(tt, acc2.ID, transfer.CreditAccountID)

	// 		freshAcc1, err := ps.GetAccount(acc1.ID)
	// 		if err != nil {
	// 			tt.Fatal(err)
	// 		}
	// 		assert.Equal(tt, uint64(100), freshAcc1.DebitsAccepted)
	// 		assert.Equal(tt, uint64(0), freshAcc1.DebitsReserved)
	// 		assert.Equal(tt, uint64(0), freshAcc1.CreditsAccepted)
	// 		assert.Equal(tt, uint64(0), freshAcc1.CreditsReserved)

	// 		freshAcc2, err := ps.GetAccount(acc2.ID)
	// 		if err != nil {
	// 			tt.Fatal(err)
	// 		}
	// 		assert.Equal(tt, uint64(0), freshAcc2.DebitsAccepted)
	// 		assert.Equal(tt, uint64(0), freshAcc2.DebitsReserved)
	// 		assert.Equal(tt, uint64(100), freshAcc2.CreditsAccepted)
	// 		assert.Equal(tt, uint64(0), freshAcc2.CreditsReserved)
	// 	})

	// 	t.Run("can only transfer between accounts in the same ledger", func(tt *testing.T) {
	// 		otherLedger, err := ps.CreateLedger(faker.Name())
	// 		if err != nil {
	// 			tt.Fatal(err)
	// 		}
	// 		acc1, err := ps.CreateAccount(CreateAccountArgs{
	// 			LedgerID: otherLedger.ID,
	// 			Unit:     1,
	// 		})
	// 		acc2, err := ps.CreateAccount(CreateAccountArgs{
	// 			LedgerID: ledger.ID,
	// 			Unit:     1,
	// 		})
	// 		if err != nil {
	// 			tt.Fatal(err)
	// 		}
	// 		transfer, err := ps.CreateTransfer(CreateTransferArgs{
	// 			Amount:          100,
	// 			DebitAccountID:  acc1.ID,
	// 			CreditAccountID: acc2.ID,
	// 		})
	// 		if err == nil {
	// 			tt.Fatal("Should fail if trying to do cross ledger transfer.")
	// 		}

	// 		assert.Nil(tt, transfer)
	// 		assert.Error(tt, err, "Accounts don't belong to the same ledger.")
	// 	})

	// 	t.Run("account codes must match those described in transaction type", func(tt *testing.T) {
	// 		acc1, err := ps.CreateAccount(CreateAccountArgs{
	// 			LedgerID: ledger.ID,
	// 			Code:     33,
	// 			Unit:     1,
	// 		})
	// 		acc2, err := ps.CreateAccount(CreateAccountArgs{
	// 			LedgerID: ledger.ID,
	// 			Unit:     1,
	// 		})
	// 		if err != nil {
	// 			tt.Fatal(err)
	// 		}
	// 		transfer, err := ps.CreateTransfer(CreateTransferArgs{
	// 			Amount:          100,
	// 			DebitAccountID:  acc1.ID,
	// 			CreditAccountID: acc2.ID,
	// 		})
	// 		if err == nil {
	// 			tt.Fatal("Should fail if account codes don't match those in transaction type.")
	// 		}

	// 		assert.Nil(tt, transfer)
	// 		assert.Error(tt, err, "Incorrect debit account category for transfer.")
	// 	})
	// })
}
