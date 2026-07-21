package rpcserver

import (
	"context"

	"github.com/interledger/interledger-app/go/pacioli"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"github.com/interledger/interledger-app/go/pacioli/healthcheck"
	"github.com/interledger/interledger-app/go/pacioli/ledger"
	pacioliv1 "github.com/interledger/interledger-app/go/proto/pacioli/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func NewServer(b ledger.Backends, hs healthcheck.Service) *grpc.Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()))
	pacioliv1.RegisterPacioliServiceServer(server, &rpcServer{b: b})
	grpc_health_v1.RegisterHealthServer(server, hs)
	reflection.Register(server)
	return server
}

type rpcServer struct {
	pacioliv1.UnimplementedPacioliServiceServer
	b ledger.Backends
}

// TODO: wire up grpc auth

func (s *rpcServer) ConfigureLedgers(ctx context.Context, req *pacioliv1.ConfigureLedgersRequest) (*pacioliv1.ConfigureLedgersResponse, error) {
	ledgers := make([]pacioli.ConfigureLedgerArgs, len(req.Args))
	for i, ledger := range req.Args {
		ledgers[i] = pacioli.ConfigureLedgerArgs{
			ID:    ledger.GetId(),
			Name:  ledger.GetName(),
			Asset: ledger.GetAsset(),
			Scale: uint8(ledger.GetScale()),
		}
	}
	eventErrors, err := ledger.ConfigureLedgers(ctx, s.b, ledgers)
	// This error will be due to connection / io / validation  issues
	if err != nil {
		switch err.(type) {
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	ret := make([]*pacioliv1.EventError, len(eventErrors))
	for i, err := range eventErrors {
		ret[i] = &pacioliv1.EventError{
			Index: err.Index,
			Code:  uint32(err.Code),
		}
	}

	return &pacioliv1.ConfigureLedgersResponse{
		Errors: ret,
	}, nil
}

func (s *rpcServer) GetLedgers(ctx context.Context, req *pacioliv1.GetLedgersRequest) (*pacioliv1.GetLedgersResponse, error) {
	ledgers, err := ledger.GetLedgers(ctx, s.b, req.GetIds())
	if err != nil {
		switch err.(type) {
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	ret := make([]*pacioliv1.Ledger, len(ledgers))
	for i, ledger := range ledgers {
		ret[i] = &pacioliv1.Ledger{
			Id:    ledger.ID,
			Name:  ledger.Name,
			Asset: ledger.Asset,
			Scale: uint32(ledger.Scale),
		}
	}

	return &pacioliv1.GetLedgersResponse{
		Ledgers: ret,
	}, nil
}

func (s *rpcServer) ConfigureAccounts(ctx context.Context, req *pacioliv1.ConfigureAccountsRequest) (*pacioliv1.ConfigureAccountsResponse, error) {
	accounts := make([]pacioli.ConfigureAccountArgs, len(req.Args))
	for i, account := range req.Args {
		accounts[i] = pacioli.ConfigureAccountArgs{
			ID:                         account.GetId(),
			LedgerID:                   account.GetLedgerId(),
			Code:                       uint16(account.GetCode()),
			DebitsMustNotExceedCredits: account.GetDebitsMustNotExceedCredits(),
			CreditsMustNotExceedDebits: account.GetCreditsMustNotExceedDebits(),
		}
	}
	eventErrors, err := ledger.ConfigureAccounts(ctx, s.b, accounts)
	// This error will be due to connection / io / validation  issues
	if err != nil {
		switch err.(type) {
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	ret := make([]*pacioliv1.EventError, len(eventErrors))
	for i, err := range eventErrors {
		ret[i] = &pacioliv1.EventError{
			Index: err.Index,
			Code:  uint32(err.Code),
		}
	}

	return &pacioliv1.ConfigureAccountsResponse{
		Errors: ret,
	}, nil
}

func (s *rpcServer) GetAccounts(ctx context.Context, req *pacioliv1.GetAccountsRequest) (*pacioliv1.GetAccountsResponse, error) {
	accounts, err := ledger.GetAccounts(ctx, s.b, req.GetIds())
	// This error will be due to connection / io / validation  issues
	if err != nil {
		switch err.(type) {
		default:
			return nil, status.Error(codes.Internal, "Failed to get account.")
		}
	}

	ret := make([]*pacioliv1.Account, len(accounts))
	for i, account := range accounts {
		ret[i] = &pacioliv1.Account{
			Id:                         account.ID,
			LedgerId:                   account.LedgerID,
			Code:                       uint32(account.Code),
			DebitsReserved:             account.DebitsPending,
			DebitsAccepted:             account.DebitsPosted,
			CreditsAccepted:            account.CreditsPosted,
			CreditsReserved:            account.CreditsPending,
			DebitsMustNotExceedCredits: account.DebitsMustNotExceedCredits,
			CreditsMustNotExceedDebits: account.CreditsMustNotExceedDebits,
		}
	}

	return &pacioliv1.GetAccountsResponse{
		Accounts: ret,
	}, nil
}

func (s *rpcServer) CreateTransfers(ctx context.Context, req *pacioliv1.CreateTransfersRequest) (*pacioliv1.CreateTransfersResponse, error) {
	transfers := req.GetTransfers()
	transferArgs := make([]pacioli.CreateTransferArgs, len(transfers))
	for i, transfer := range transfers {
		transferArgs[i] = pacioli.CreateTransferArgs{
			ID:              transfer.GetId(),
			Amount:          int64(transfer.GetAmount()),
			DebitAccountID:  transfer.GetDebitAccountId(),
			CreditAccountID: transfer.GetCreditAccountId(),
			Code:            uint16(transfer.GetCode()),
			Timeout:         transfer.GetTimeout(),
			Ledger:          transfer.Ledger,
			Pending:         transfer.Pending,
		}
	}

	errors, err := ledger.CreateTransfers(ctx, s.b, transferArgs)
	if err != nil {
		switch err.(type) {
		default:
			return nil, status.Error(codes.Internal, "Failed to create transfer."+err.Error())
		}
	}

	errorsToReturn := make([]*pacioliv1.EventError, len(errors))
	for i, err := range errors {
		errorsToReturn[i] = &pacioliv1.EventError{
			Index: err.Index,
			Code:  uint32(err.Code),
		}
	}

	return &pacioliv1.CreateTransfersResponse{
		Errors: errorsToReturn,
	}, nil
}

func (s *rpcServer) GetTransfers(ctx context.Context, req *pacioliv1.GetTransfersRequest) (*pacioliv1.GetTransfersResponse, error) {
	transfers, err := ledger.GetTransfers(ctx, s.b, req.GetIds())
	if err != nil {
		switch err.(type) {
		default:
			return nil, status.Error(codes.Internal, "Failed to get transfers.")
		}
	}

	transfersToReturn := make([]*pacioliv1.Transfer, len(transfers))
	for i, transfer := range transfers {
		transfersToReturn[i] = &pacioliv1.Transfer{
			Id:              transfer.ID,
			DebitAccountId:  transfer.DebitAccountID,
			CreditAccountId: transfer.CreditAccountID,
			Amount:          transfer.Amount,
			Code:            uint32(transfer.Code),
			Pending:         transfer.State == pacioli.TransferStatePending,
		}
	}

	return &pacioliv1.GetTransfersResponse{
		Transfers: transfersToReturn,
	}, nil
}

func (s *rpcServer) PostTransfers(ctx context.Context, req *pacioliv1.PostTransfersRequest) (*pacioliv1.PostTransfersResponse, error) {

	errors, err := ledger.PostTransfers(ctx, s.b, req.TransferIds)
	if err != nil {
		switch err.(type) {
		default:
			return nil, status.Error(codes.Internal, "Failed to commit transfer."+err.Error())
		}
	}

	errorsToReturn := make([]*pacioliv1.EventError, len(errors))
	for i, err := range errors {
		errorsToReturn[i] = &pacioliv1.EventError{
			Index: err.Index,
			Code:  uint32(err.Code),
		}
	}

	return &pacioliv1.PostTransfersResponse{
		Errors: errorsToReturn,
	}, nil
}

func (s *rpcServer) VoidTransfers(ctx context.Context, req *pacioliv1.VoidTransfersRequest) (*pacioliv1.VoidTransfersResponse, error) {

	errors, err := ledger.VoidTransfers(ctx, s.b, req.TransferIds)
	if err != nil {
		switch err.(type) {
		default:
			return nil, status.Error(codes.Internal, "Failed to commit transfer."+err.Error())
		}
	}

	errorsToReturn := make([]*pacioliv1.EventError, len(errors))
	for i, err := range errors {
		errorsToReturn[i] = &pacioliv1.EventError{
			Index: err.Index,
			Code:  uint32(err.Code),
		}
	}

	return &pacioliv1.VoidTransfersResponse{
		Errors: errorsToReturn,
	}, nil
}
