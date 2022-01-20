package pacioli

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/pacioli/utils"
	"gitlab.com/fynbos/tigerbeetle_go"
)

func TestPacioliService(s *testing.T) {
	ctx := context.Background()
	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		s.Fatal(err)
	}

	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		s.Fatal(err)
	}
	defer db.Close()

	var tbClusterID uint32 = 0
	tb, err := test_utils.SetupTigerBeetle(ctx, tbClusterID)
	if err != nil {
		s.Fatal(err)
	}
	fmt.Println("URI", tb.URI)
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

		t.Run("tenant must exist to create ledger", func(tt *testing.T) {
			ledger, err := ps.CreateLedger("my first ledger")
			if err == nil {
				tt.Fatal("Tenant must exist to create ledger.")
			}

			assert.Nil(tt, ledger)
			assert.Equal(tt, "Tenant not found.", err.Error())
		})

		t.Run("tenant can create a ledger", func(tt *testing.T) {
			ledger, err := ps.CreateLedger("my first ledger")
			if err != nil {
				t.Fatal(err)
			}

			assert.NotNil(tt, ledger.ID)
			assert.Equal(tt, "my first ledger", ledger.Name)
		})
	})

	s.Run("accounts and transfers", func(t *testing.T) {
		t.Cleanup(func() {
			test_utils.TruncateDb(ctx, db)
		})
		ledger, err := ps.CreateLedger(faker.Name())
		if err != nil {
			t.Fatal(err)
		}

		t.Run("can create accounts", func(tt *testing.T) {
			acc, err := ps.CreateAccount(CreateAccountArgs{
				LedgerID: ledger.ID,
				Unit:     1,
			})
			if err != nil {
				tt.Fatal(err)
			}

			assert.Equal(tt, ledger.ID, acc.LedgerID)
			assert.Equal(tt, uint16(1), acc.Unit)
			assert.Equal(tt, uint64(0), acc.DebitsAccepted)
			assert.Equal(tt, uint64(0), acc.DebitsReserved)
			assert.Equal(tt, uint64(0), acc.CreditsAccepted)
			assert.Equal(tt, uint64(0), acc.CreditsReserved)
		})

		t.Run("tenant can create a transfer", func(tt *testing.T) {
			acc1, err := ps.CreateAccount(CreateAccountArgs{
				LedgerID: ledger.ID,
				Unit:     1,
			})
			acc2, err := ps.CreateAccount(CreateAccountArgs{
				LedgerID: ledger.ID,
				Unit:     1,
			})
			if err != nil {
				tt.Fatal(err)
			}
			transfer, err := ps.CreateTransfer(CreateTransferArgs{
				Amount:          100,
				DebitAccountID:  acc1.ID,
				CreditAccountID: acc2.ID,
			})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, uint64(100), transfer.Amount)
			assert.Equal(tt, acc1.ID, transfer.DebitAccountID)
			assert.Equal(tt, acc2.ID, transfer.CreditAccountID)

			freshAcc1, err := ps.GetAccount(acc1.ID)
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, uint64(100), freshAcc1.DebitsAccepted)
			assert.Equal(tt, uint64(0), freshAcc1.DebitsReserved)
			assert.Equal(tt, uint64(0), freshAcc1.CreditsAccepted)
			assert.Equal(tt, uint64(0), freshAcc1.CreditsReserved)

			freshAcc2, err := ps.GetAccount(acc2.ID)
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, uint64(0), freshAcc2.DebitsAccepted)
			assert.Equal(tt, uint64(0), freshAcc2.DebitsReserved)
			assert.Equal(tt, uint64(100), freshAcc2.CreditsAccepted)
			assert.Equal(tt, uint64(0), freshAcc2.CreditsReserved)
		})

		t.Run("can only transfer between accounts in the same ledger", func(tt *testing.T) {
			otherLedger, err := ps.CreateLedger(faker.Name())
			if err != nil {
				tt.Fatal(err)
			}
			acc1, err := ps.CreateAccount(CreateAccountArgs{
				LedgerID: otherLedger.ID,
				Unit:     1,
			})
			acc2, err := ps.CreateAccount(CreateAccountArgs{
				LedgerID: ledger.ID,
				Unit:     1,
			})
			if err != nil {
				tt.Fatal(err)
			}
			transfer, err := ps.CreateTransfer(CreateTransferArgs{
				Amount:          100,
				DebitAccountID:  acc1.ID,
				CreditAccountID: acc2.ID,
			})
			if err == nil {
				tt.Fatal("Should fail if trying to do cross ledger transfer.")
			}

			assert.Nil(tt, transfer)
			assert.Error(tt, err, "Accounts don't belong to the same ledger.")
		})

		t.Run("account codes must match those described in transaction type", func(tt *testing.T) {
			acc1, err := ps.CreateAccount(CreateAccountArgs{
				LedgerID: ledger.ID,
				Code:     33,
				Unit:     1,
			})
			acc2, err := ps.CreateAccount(CreateAccountArgs{
				LedgerID: ledger.ID,
				Unit:     1,
			})
			if err != nil {
				tt.Fatal(err)
			}
			transfer, err := ps.CreateTransfer(CreateTransferArgs{
				Amount:          100,
				DebitAccountID:  acc1.ID,
				CreditAccountID: acc2.ID,
			})
			if err == nil {
				tt.Fatal("Should fail if account codes don't match those in transaction type.")
			}

			assert.Nil(tt, transfer)
			assert.Error(tt, err, "Incorrect debit account category for transfer.")
		})
	})
}
