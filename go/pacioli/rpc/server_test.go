package rpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/jmoiron/sqlx"
	ledger "gitlab.com/fynbos/pacioli/ledger"
	test_utils "gitlab.com/fynbos/pacioli/utils"
	"gitlab.com/fynbos/tigerbeetle_go"
	"gotest.tools/assert"

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
		ledger, err := client.CreateLedger(ctx, &pacioliv1.CreateLedgerRequest{
			Name: name,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, name, ledger.Name)
	})

	s.Run("can create an account", func(t *testing.T) {

		ledger, err := client.CreateLedger(ctx, &pacioliv1.CreateLedgerRequest{
			Name: faker.Name(),
		})
		if err != nil {
			t.Fatal(err)
		}

		account, err := client.CreateAccount(ctx, &pacioliv1.CreateAccountRequest{
			LedgerId: ledger.Id,
			Unit:     1,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint32(1), account.Unit)
		assert.Equal(t, uint32(1), account.Code)
		assert.Equal(t, ledger.Id, account.LedgerId)
		assert.Equal(t, uint64(0), account.DebitsAccepted)
		assert.Equal(t, uint64(0), account.DebitsReserved)
		assert.Equal(t, uint64(0), account.CreditsAccepted)
		assert.Equal(t, uint64(0), account.CreditsReserved)

		freshAccount, err := client.GetAccount(ctx, &pacioliv1.GetAccountRequest{
			Id: account.Id,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint32(1), freshAccount.Unit)
		assert.Equal(t, uint32(1), freshAccount.Code)
		assert.Equal(t, ledger.Id, freshAccount.LedgerId)
		assert.Equal(t, uint64(0), freshAccount.DebitsAccepted)
		assert.Equal(t, uint64(0), freshAccount.DebitsReserved)
		assert.Equal(t, uint64(0), freshAccount.CreditsAccepted)
		assert.Equal(t, uint64(0), freshAccount.CreditsReserved)
	})

	s.Run("can create transfer", func(t *testing.T) {
		ledger, err := client.CreateLedger(ctx, &pacioliv1.CreateLedgerRequest{
			Name: faker.Name(),
		})
		if err != nil {
			t.Fatal(err)
		}

		acc1, err := client.CreateAccount(ctx, &pacioliv1.CreateAccountRequest{
			LedgerId: ledger.Id,
			Unit:     1,
		})
		if err != nil {
			t.Fatal(err)
		}
		acc2, err := client.CreateAccount(ctx, &pacioliv1.CreateAccountRequest{
			LedgerId: ledger.Id,
			Unit:     1,
		})
		if err != nil {
			t.Fatal(err)
		}

		transfer, err := client.CreateTransfer(ctx, &pacioliv1.CreateTransferRequest{
			DebitAccountId:  acc1.Id,
			CreditAccountId: acc2.Id,
			Amount:          100,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint64(100), transfer.Amount)
		assert.Equal(t, acc1.Id, transfer.DebitAccountId)
		assert.Equal(t, acc2.Id, transfer.CreditAccountId)

		freshAcc1, err := client.GetAccount(ctx, &pacioliv1.GetAccountRequest{
			Id: acc1.Id,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint64(100), freshAcc1.DebitsAccepted)
		assert.Equal(t, uint64(0), freshAcc1.DebitsReserved)
		assert.Equal(t, uint64(0), freshAcc1.CreditsAccepted)
		assert.Equal(t, uint64(0), freshAcc1.CreditsReserved)

		freshAcc2, err := client.GetAccount(ctx, &pacioliv1.GetAccountRequest{
			Id: acc2.Id,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint64(0), freshAcc2.DebitsAccepted)
		assert.Equal(t, uint64(0), freshAcc2.DebitsReserved)
		assert.Equal(t, uint64(100), freshAcc2.CreditsAccepted)
		assert.Equal(t, uint64(0), freshAcc2.CreditsReserved)
	})
}
