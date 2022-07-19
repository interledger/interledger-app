package client

import (
	"context"

	tb_types "github.com/coilhq/tigerbeetle-go/pkg/types"

	pb "gitlab.com/fynbos/proto/pacioli/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gitlab.com/fynbos/pacioli"
)

var _ pacioli.Client = client{}

type client struct {
	c pb.PacioliServiceClient
}

func Make(grpcAddress string) (pacioli.Client, error) {
	conn, err := grpc.Dial(grpcAddress, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	pClient := pb.NewPacioliServiceClient(conn)

	return &client{c: pClient}, nil
}

func (c client) ConfigureLedgers(ctx context.Context, args []pacioli.ConfigureLedgerArgs) ([]pacioli.EventResult, error) {
	la := make([]*pb.Ledger, len(args))
	for i, a := range args {
		la[i] = &pb.Ledger{
			Id:    a.ID,
			Name:  a.Name,
			Asset: a.Asset,
			Scale: uint32(a.Scale),
		}
	}

	res, err := c.c.ConfigureLedgers(ctx, &pb.ConfigureLedgersRequest{Args: la})
	if err != nil {
		return nil, err
	}

	resp := make([]pacioli.EventResult, len(res.Errors))
	for i, le := range res.Errors {
		resp[i] = pacioli.EventResult{
			Index: le.Index,
			Code:  le.Code,
		}
	}

	return resp, nil
}

func (c client) GetLedgers(ctx context.Context, ledgerIDs []uint32) ([]pacioli.Ledger, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) ConfigureAccounts(ctx context.Context, args []pacioli.ConfigureAccountArgs) ([]pacioli.AccountResult, error) {
	aa := make([]*pb.ConfigureAccountsArgs, len(args))
	for i, a := range args {
		aa[i] = &pb.ConfigureAccountsArgs{
			Id:       a.ID,
			LedgerId: a.LedgerID,
			Code:     uint32(a.Code),
			Flags: &pb.AccountFlags{
				Linked:                     a.Flags.Linked,
				DebitsMustNotExceedCredits: a.Flags.DebitsMustNotExceedCredits,
				CreditsMustNotExceedDebits: a.Flags.CreditsMustNotExceedDebits,
			},
		}
	}

	res, err := c.c.ConfigureAccounts(ctx, &pb.ConfigureAccountsRequest{Args: aa})
	if err != nil {
		return nil, err
	}

	resp := make([]pacioli.AccountResult, len(res.Errors))
	for i, le := range res.Errors {
		resp[i] = pacioli.AccountResult{
			Index: le.Index,
			Code:  tb_types.CreateAccountResult(le.Code),
		}
	}

	return resp, nil
}

func (c client) GetAccounts(ctx context.Context, accountIDs []string) ([]pacioli.Account, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) CreateTransfers(ctx context.Context, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) GetTransfers(ctx context.Context, transferIDs []string) ([]pacioli.Transfer, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) CommitTransfers(ctx context.Context, transferIDs []string) ([]pacioli.TransferResult, error) {
	//TODO implement me
	panic("implement me")
}

func (c client) VoidTransfers(ctx context.Context, transferIDs []string) ([]pacioli.TransferResult, error) {
	//TODO implement me
	panic("implement me")
}
