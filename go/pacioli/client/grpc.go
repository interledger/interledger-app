package client

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"github.com/interledger/interledger-app/go/pacioli"
	pb "github.com/interledger/interledger-app/go/proto/pacioli/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ pacioli.Client = client{}

type client struct {
	c pb.PacioliServiceClient
}

func New(grpcAddress string) (pacioli.Client, error) {
	conn, err := grpc.Dial(grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		return nil, err
	}

	pClient := pb.NewPacioliServiceClient(conn)

	return &client{c: pClient}, nil
}

func (c client) ConfigureLedgers(ctx context.Context, args []pacioli.ConfigureLedgerArgs) ([]pacioli.LedgerResult, error) {
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

	resp := make([]pacioli.LedgerResult, len(res.Errors))
	for i, le := range res.Errors {
		resp[i] = pacioli.LedgerResult{
			Index: le.Index,
			Code:  pacioli.LedgerResultCode(le.Code),
		}
	}

	return resp, nil
}

func (c client) GetLedgers(ctx context.Context, ledgerIDs []uint32) ([]pacioli.Ledger, error) {
	lresp, err := c.c.GetLedgers(ctx, &pb.GetLedgersRequest{Ids: ledgerIDs})
	if err != nil {
		return nil, err
	}

	res := make([]pacioli.Ledger, len(lresp.Ledgers))
	for i, l := range lresp.Ledgers {
		res[i] = pacioli.Ledger{
			ID:    l.Id,
			Name:  l.Name,
			Asset: l.Asset,
			Scale: uint8(l.Scale),
		}
	}

	return res, nil
}

func (c client) ConfigureAccounts(ctx context.Context, args []pacioli.ConfigureAccountArgs) ([]pacioli.AccountResult, error) {
	aa := make([]*pb.ConfigureAccountsArgs, len(args))
	for i, a := range args {
		aa[i] = &pb.ConfigureAccountsArgs{
			Id:                         a.ID,
			LedgerId:                   a.LedgerID,
			Code:                       uint32(a.Code),
			DebitsMustNotExceedCredits: a.DebitsMustNotExceedCredits,
			CreditsMustNotExceedDebits: a.CreditsMustNotExceedDebits,
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
			Code:  pacioli.AccountResultCode(le.Code),
		}
	}

	return resp, nil
}

func (c client) GetAccounts(ctx context.Context, accountIDs []string) ([]pacioli.Account, error) {
	aresp, err := c.c.GetAccounts(ctx, &pb.GetAccountsRequest{Ids: accountIDs})
	if err != nil {
		return nil, err
	}

	res := make([]pacioli.Account, len(aresp.Accounts))
	for i, l := range aresp.Accounts {
		res[i] = pacioli.Account{
			ID:                         l.Id,
			LedgerID:                   l.LedgerId,
			Code:                       uint16(l.Code),
			DebitsPending:              l.DebitsReserved,
			DebitsPosted:               l.DebitsAccepted,
			CreditsPending:             l.CreditsReserved,
			CreditsPosted:              l.CreditsAccepted,
			DebitsMustNotExceedCredits: l.DebitsMustNotExceedCredits,
			CreditsMustNotExceedDebits: l.CreditsMustNotExceedDebits,
		}
	}

	return res, nil
}

func (c client) CreateTransfers(ctx context.Context, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error) {
	ta := make([]*pb.Transfer, len(args))
	for i, a := range args {
		ta[i] = &pb.Transfer{
			Id:              a.ID,
			DebitAccountId:  a.DebitAccountID,
			CreditAccountId: a.CreditAccountID,
			Amount:          uint64(a.Amount),
			Code:            uint32(a.Code),
			Timeout:         a.Timeout,
			Ledger:          a.Ledger,
			Pending:         a.Pending,
		}
	}

	res, err := c.c.CreateTransfers(ctx, &pb.CreateTransfersRequest{Transfers: ta})
	if err != nil {
		return nil, err
	}

	resp := make([]pacioli.TransferResult, len(res.Errors))
	for i, le := range res.Errors {
		resp[i] = pacioli.TransferResult{
			Index: le.Index,
			Code:  pacioli.TransferResultCode(le.Code),
		}
	}

	return resp, nil
}

func (c client) GetTransfers(ctx context.Context, transferIDs []string) ([]pacioli.Transfer, error) {
	tresp, err := c.c.GetTransfers(ctx, &pb.GetTransfersRequest{Ids: transferIDs})
	if err != nil {
		return nil, err
	}

	res := make([]pacioli.Transfer, len(tresp.Transfers))
	for i, l := range tresp.Transfers {
		res[i] = pacioli.Transfer{
			ID:              l.Id,
			LedgerID:        l.Ledger,
			DebitAccountID:  l.DebitAccountId,
			CreditAccountID: l.CreditAccountId,
			Amount:          l.Amount,
			Code:            uint16(l.Code),
			Timeout:         l.Timeout,
		}
	}

	return res, nil
}

func (c client) PostTransfers(ctx context.Context, transferIDs []string) ([]pacioli.TransferResult, error) {
	presp, err := c.c.PostTransfers(ctx, &pb.PostTransfersRequest{TransferIds: transferIDs})
	if err != nil {
		return nil, err
	}

	resp := make([]pacioli.TransferResult, len(presp.Errors))
	for i, le := range presp.Errors {
		resp[i] = pacioli.TransferResult{
			Index: le.Index,
			Code:  pacioli.TransferResultCode(le.Code),
		}
	}

	return resp, nil
}

func (c client) VoidTransfers(ctx context.Context, transferIDs []string) ([]pacioli.TransferResult, error) {
	vresp, err := c.c.VoidTransfers(ctx, &pb.VoidTransfersRequest{TransferIds: transferIDs})
	if err != nil {
		return nil, err
	}

	resp := make([]pacioli.TransferResult, len(vresp.Errors))
	for i, le := range vresp.Errors {
		resp[i] = pacioli.TransferResult{
			Index: le.Index,
			Code:  pacioli.TransferResultCode(le.Code),
		}
	}

	return resp, nil
}
