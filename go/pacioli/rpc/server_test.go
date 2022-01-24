package rpc

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	ledger "gitlab.com/fynbos/pacioli/ledger"
	test_utils "gitlab.com/fynbos/pacioli/utils"
	"gitlab.com/fynbos/tigerbeetle_go"
	tigerbeetleTypes "gitlab.com/fynbos/tigerbeetle_go/pkg/types"

	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"google.golang.org/grpc"
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

	ps, err := ledger.NewLedgerService(db, tbClient)
	if err != nil {
		s.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		s.Fatal(err)
	}
	server := NewServer(ps)
	go func() {
		if err := server.Serve(listener); err != nil {
			panic(err)
		}
	}()

	connectTo := "127.0.0.1:8081"
	conn, err := grpc.Dial(connectTo, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		s.Fatal(err)
	}
	client := pacioliv1.NewPacioliServiceClient(conn)

	s.Cleanup(func() {
		fmt.Println("cleaning up")
		server.Stop()
		// tbClient.Deinit()
		os.RemoveAll(tb.DataDir)
		// tb.Container.Terminate(ctx)

		db.Close()
		crdb.Container.Terminate(ctx)
	})

	s.Run("can create a ledger", func(t *testing.T) {
		name := faker.Name()
		code := uint32(rand.Intn(65535))
		ledger, err := client.CreateLedger(ctx, &pacioliv1.CreateLedgerRequest{
			Name: name,
			Code: code,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, name, ledger.Name)
		assert.Equal(t, code, uint32(ledger.Code))
	})

	s.Run("can create an account on a ledger", func(t *testing.T) {
		ledger, err := client.CreateLedger(ctx, &pacioliv1.CreateLedgerRequest{
			Name: faker.Name(),
			Code: uint32(rand.Intn(65535)),
		})
		if err != nil {
			t.Fatal(err)
		}

		account, err := client.CreateAccount(ctx, &pacioliv1.CreateAccountRequest{
			LedgerID: ledger.Id,
			Unit:     1,
			Code:     1,
		})
		if err != nil {
			t.Fatal(err)
		}
		uuid.MustParse(account.Id)
		assert.Equal(t, uint32(1), account.Code)
		assert.Equal(t, ledger.Code, account.LedgerCode)
		assert.Equal(t, uint64(0), account.DebitsAccepted)
		assert.Equal(t, uint64(0), account.DebitsReserved)
		assert.Equal(t, uint64(0), account.CreditsAccepted)
		assert.Equal(t, uint64(0), account.CreditsReserved)

		freshAccount, err := client.GetAccount(ctx, &pacioliv1.GetAccountRequest{
			Id:       account.Id,
			LedgerID: ledger.Id,
		})
		if err != nil {
			t.Fatal(err)
		}
		uuid.MustParse(freshAccount.Id)
		assert.Equal(t, uint32(1), freshAccount.Code)
		assert.Equal(t, ledger.Code, freshAccount.LedgerCode)
		assert.Equal(t, uint64(0), freshAccount.DebitsAccepted)
		assert.Equal(t, uint64(0), freshAccount.DebitsReserved)
		assert.Equal(t, uint64(0), freshAccount.CreditsAccepted)
		assert.Equal(t, uint64(0), freshAccount.CreditsReserved)
	})

	s.Run("can create a batch of transfers", func(t *testing.T) {
		ledger, err := client.CreateLedger(ctx, &pacioliv1.CreateLedgerRequest{
			Name: faker.Name(),
			Code: uint32(rand.Intn(65535)),
		})
		if err != nil {
			t.Fatal(err)
		}
		accountA, err := client.CreateAccount(ctx, &pacioliv1.CreateAccountRequest{
			LedgerID: ledger.Id,
			Unit:     1,
			Code:     1,
		})
		if err != nil {
			t.Fatal(err)
		}
		accountB, err := client.CreateAccount(ctx, &pacioliv1.CreateAccountRequest{
			LedgerID: ledger.Id,
			Unit:     1,
			Code:     1,
		})
		if err != nil {
			t.Fatal(err)
		}
		transferA := uuid.NewString()
		transferB := uuid.NewString()

		response, err := client.CreateTransfers(ctx, &pacioliv1.CreateTransfersRequest{
			LedgerID: ledger.Id,
			Transfers: []*pacioliv1.Transfer{
				{
					Id:              transferA,
					DebitAccountId:  accountA.Id,
					CreditAccountId: accountB.Id,
					Amount:          100,
					Code:            1,
				},
				{ // This one will fail as the account IDs are the same.
					Id:              transferB,
					DebitAccountId:  accountA.Id,
					CreditAccountId: accountA.Id,
					Amount:          101,
					Code:            2,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		errors := response.GetErrors()
		assert.Len(t, errors, 1)
		assert.Equal(t, uint32(1), errors[0].Index)
		assert.Equal(t, tigerbeetleTypes.TransferAccountsAreTheSame, errors[0].Code)

		lookupResponse, err := client.GetTransfers(ctx, &pacioliv1.GetTransfersRequest{
			LedgerID:    ledger.Id,
			TransferIDs: []string{transferA, transferB},
		})
		if err != nil {
			t.Fatal(err)
		}
		transfers := lookupResponse.GetTransfers()
		assert.Len(t, transfers, 1)
		assert.Equal(t, accountA.Id, transfers[0].GetDebitAccountId())
		assert.Equal(t, accountB.Id, transfers[0].GetCreditAccountId())
		assert.Equal(t, uint64(100), transfers[0].GetAmount())
		assert.Equal(t, uint32(1), transfers[0].GetCode())

		flags := transfers[0].GetFlags()
		assert.Equal(t, false, flags.GetCondition())
		assert.Equal(t, false, flags.GetLinked())
		assert.Equal(t, false, flags.GetTwoPhaseCommit())
	})

}
