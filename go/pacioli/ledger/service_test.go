package pacioli

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/pacioli/utils"
	"gitlab.com/fynbos/tigerbeetle_go"
	tigerbeetle_types "gitlab.com/fynbos/tigerbeetle_go/pkg/types"
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
		// tb.Container.Terminate(ctx)

		err := db.Close()
		if err != nil {
			return
		}
		err = crdb.Container.Terminate(ctx)
		if err != nil {
			return
		}
	})

	s.Run("ledger", func(t *testing.T) {
		t.Cleanup(func() {
			err := test_utils.TruncateDb(ctx, db)
			if err != nil {
				return
			}
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
			assert.Equal(tt, ErrDuplicate{Err: "Ledger exists."}, err)
		})

		t.Run("can get ledger by code", func(tt *testing.T) {
			name := faker.Name()
			code := uint16(rand.Intn(65535))

			ledger, err := ps.GetLedgerByCode(code)
			assert.Equal(tt, "Ledger not found.", err.Error())
			assert.Nil(tt, ledger)

			_, err = ps.CreateLedger(name, code)
			if err != nil {
				t.Fatal(err)
			}
			if err != nil {
				tt.Fatal(err)
			}

			freshLedger, err := ps.GetLedgerByCode(code)
			if err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, code, freshLedger.Code)
			assert.Equal(tt, name, freshLedger.Name)
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

	s.Run("transfers", func(t *testing.T) {
		ledger, err := ps.CreateLedger(faker.Name(), uint16(rand.Intn(65535)))
		if err != nil {
			t.Fatal(err)
		}
		accA := CreateAccountArgs{
			ID:   uuid.NewString(),
			Code: uint16(rand.Intn(65535)),
		}
		accB := CreateAccountArgs{
			ID:   uuid.NewString(),
			Code: uint16(rand.Intn(65535)),
		}
		eventErrors, err := ps.CreateAccounts(ledger.ID, []CreateAccountArgs{accA, accB})
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, eventErrors)

		t.Run("can send a transfer between two accounts on the same ledger", func(tt *testing.T) {
			transferID := uuid.NewString()
			transferArgs := CreateTransferArgs{
				ID:              transferID,
				DebitAccountID:  accA.ID,
				CreditAccountID: accB.ID,
				Amount:          100,
				Code:            5,
			}

			eventErrors, err := ps.CreateTransfers(ledger.ID, []CreateTransferArgs{transferArgs})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Empty(tt, eventErrors)

			results, err := ps.GetTransfers(ledger.ID, []string{transferID})
			if err != nil {
				tt.Fatal(err)
			}

			assert.Len(tt, results, 1)
			assert.Equal(tt, transferID, results[0].ID)           // make sure uuid encoding is correct
			assert.Equal(tt, accA.ID, results[0].DebitAccountID)  // make sure uuid encoding is correct
			assert.Equal(tt, accB.ID, results[0].CreditAccountID) // make sure uuid encoding is correct
			assert.Equal(tt, uint64(100), results[0].Amount)
			assert.Equal(tt, uint32(5), results[0].Code)
		})

		t.Run("prohibits cross-ledger transfers", func(tt *testing.T) {
			otherLedger, err := ps.CreateLedger("other ledger", 1)
			if err != nil {
				tt.Fatal(err)
			}
			otherAcc := CreateAccountArgs{
				ID:   uuid.NewString(),
				Code: uint16(rand.Intn(65535)),
			}
			results, err := ps.CreateAccounts(otherLedger.ID, []CreateAccountArgs{otherAcc})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Empty(tt, results)

			transferID := uuid.NewString()
			transferArgs := CreateTransferArgs{
				ID:              transferID,
				DebitAccountID:  accA.ID,
				CreditAccountID: otherAcc.ID,
				Amount:          100,
				Code:            5,
			}

			eventErrors, err := ps.CreateTransfers(ledger.ID, []CreateTransferArgs{transferArgs})
			if err != nil {
				tt.Fatal(err)
			}
			assert.Len(tt, eventErrors, 1)
			assert.Equal(tt, uint32(0), eventErrors[0].Index)
			// ledger code is used as `Unit` on the account. TB is planning to rename `Unit` to `Ledger`.
			assert.Equal(tt, tigerbeetle_types.TransferAccountsHaveDifferentUnits, eventErrors[0].Code)
		})
	})
}
