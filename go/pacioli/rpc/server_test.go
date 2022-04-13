package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	tb_types "github.com/coilhq/tigerbeetle-go/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestRpc(s *testing.T) {
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

	ledgerID := uint32(2)
	accountAID := uuid.NewString()
	accountBID := uuid.NewString()

	s.Run("can perform a health check", func(t *testing.T) {
		healthClient := grpc_health_v1.NewHealthClient(c.Connection)

		response, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
			Service: "pacioli",
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, response.Status)
	})

	// TODO: tests to configure a tenant. At the moment DEV_TENANT is hardcoded.

	s.Run("can configure ledgers", func(t *testing.T) {
		name := faker.Name()
		scale := uint8(2)
		configureResponse, err := c.Client.ConfigureLedgers(ctx, &pacioliv1.ConfigureLedgersRequest{
			Args: []*pacioliv1.Ledger{
				{
					Id:    uint32(ledgerID),
					Name:  name,
					Asset: "840",
					Scale: uint32(scale),
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, configureResponse.GetErrors(), 0)

		response, err := c.Client.GetLedgers(ctx, &pacioliv1.GetLedgersRequest{
			Ids: []uint32{uint32(ledgerID)},
		})
		if err != nil {
			t.Fatal(err)
		}
		ledgers := response.GetLedgers()
		assert.Len(t, ledgers, 1)
		assert.Equal(t, name, ledgers[0].Name)
		assert.Equal(t, scale, uint8(ledgers[0].Scale))
		assert.Equal(t, "840", ledgers[0].Asset)
	})

	s.Run("can configure accounts", func(t *testing.T) {
		configureResponse, err := c.Client.ConfigureAccounts(ctx, &pacioliv1.ConfigureAccountsRequest{
			Args: []*pacioliv1.ConfigureAccountsArgs{
				{
					Id:       accountAID,
					LedgerId: ledgerID,
				},
				{
					Id:       accountBID,
					LedgerId: ledgerID,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, configureResponse.GetErrors(), 0)

		response, err := c.Client.GetAccounts(ctx, &pacioliv1.GetAccountsRequest{
			Ids: []string{accountAID, accountBID},
		})
		if err != nil {
			t.Fatal(err)
		}
		accounts := response.GetAccounts()
		assert.Len(t, accounts, 2)
		assert.Equal(t, accountAID, accounts[0].Id)
		assert.Equal(t, ledgerID, accounts[0].LedgerId)
		assert.Equal(t, accountBID, accounts[1].Id)
		assert.Equal(t, ledgerID, accounts[1].LedgerId)
	})

	s.Run("can create transfers", func(t *testing.T) {
		transferA := uuid.NewString()
		transferB := uuid.NewString()

		createTransfersResponse, err := c.Client.CreateTransfers(ctx, &pacioliv1.CreateTransfersRequest{
			Transfers: []*pacioliv1.Transfer{
				{
					Id:              transferA,
					DebitAccountId:  accountAID,
					CreditAccountId: accountBID,
					Amount:          100,
					Code:            1,
					Flags: &pacioliv1.TransferFlags{
						TwoPhaseCommit: true,
					},
					Timeout: uint64(10 * time.Millisecond),
				},
				{ // This one will fail as the account IDs are the same.
					Id:              transferB,
					DebitAccountId:  accountAID,
					CreditAccountId: accountAID,
					Amount:          101,
					Code:            2,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		errors := createTransfersResponse.GetErrors()
		assert.Len(t, errors, 1)
		assert.Equal(t, uint32(1), errors[0].Index)
		assert.Equal(t, tb_types.TransferAccountsAreTheSame, errors[0].Code)

		response, err := c.Client.GetTransfers(ctx, &pacioliv1.GetTransfersRequest{
			Ids: []string{transferA, transferB},
		})
		if err != nil {
			t.Fatal(err)
		}
		transfers := response.GetTransfers()
		assert.Len(t, transfers, 1)
		assert.Equal(t, accountAID, transfers[0].GetDebitAccountId())
		assert.Equal(t, accountBID, transfers[0].GetCreditAccountId())
		assert.Equal(t, uint64(100), transfers[0].GetAmount())
		assert.Equal(t, uint32(1), transfers[0].GetCode())

		flags := transfers[0].GetFlags()
		assert.Equal(t, false, flags.GetCondition())
		assert.Equal(t, false, flags.GetLinked())
		assert.Equal(t, true, flags.GetTwoPhaseCommit())
	})

	s.Run("can commit transfers", func(t *testing.T) {
		scenarios := []struct {
			Name       string
			TransferID string
			Reject     bool
		}{
			{
				Name:       "can commit transfer",
				TransferID: uuid.NewString(),
				Reject:     false,
			},
			{
				Name:       "can reject transfer",
				TransferID: uuid.NewString(),
				Reject:     true,
			},
		}

		for _, scenario := range scenarios {
			createTransfersResponse, err := c.Client.CreateTransfers(ctx, &pacioliv1.CreateTransfersRequest{
				Transfers: []*pacioliv1.Transfer{
					{
						Id:              scenario.TransferID,
						DebitAccountId:  accountAID,
						CreditAccountId: accountBID,
						Amount:          100,
						Code:            1,
						Flags: &pacioliv1.TransferFlags{
							TwoPhaseCommit: true,
						},
						Timeout: uint64(1 * time.Second),
					},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			errors := createTransfersResponse.GetErrors()
			assert.Len(t, errors, 0)

			commitTransfersResponse, err := c.Client.CommitTransfers(ctx, &pacioliv1.CommitTransfersRequest{
				Commits: []*pacioliv1.Commit{
					{
						Id:   scenario.TransferID,
						Code: 1,
						Flags: &pacioliv1.CommitFlags{
							Reject: scenario.Reject,
						},
					},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			errors = commitTransfersResponse.GetErrors()
			assert.Len(t, errors, 0)
		}
	})

}
