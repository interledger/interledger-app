package rpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"gitlab.com/fynbos/pacioli/pacioli"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

func NewServer(ps pacioli.Service) *grpc.Server {
	server := grpc.NewServer()
	pacioliv1.RegisterPacioliServiceServer(server, &rpcServer{ps: ps})
	reflection.Register(server)
	return server
}

type rpcServer struct {
	pacioliv1.UnimplementedPacioliServiceServer
	ps pacioli.Service
}

func (s *rpcServer) CreateLedger(ctx context.Context, req *pacioliv1.CreateLedgerRequest) (*pacioliv1.Ledger, error) {
	ledger, err := s.ps.CreateLedger(req.GetName())
	if err != nil {
		// TODO: switch on err
		return nil, status.Error(codes.Internal, "Failed to create ledger.")
	}

	return &pacioliv1.Ledger{
		Id:   ledger.ID,
		Name: ledger.Name,
	}, nil
}

func (s *rpcServer) CreateAccount(ctx context.Context, req *pacioliv1.CreateAccountRequest) (*pacioliv1.Account, error) {
	account, err := s.ps.CreateAccount(pacioli.CreateAccountArgs{
		LedgerID: req.GetLedgerId(),
		Code:     uint16(req.GetCode()),
		Unit:     uint16(req.GetUnit()),
	})
	if err != nil {
		// TODO: switch on err
		return nil, status.Error(codes.Internal, "Failed to create account category.")
	}

	return &pacioliv1.Account{
		Id:              account.ID,
		LedgerId:        account.LedgerID,
		Unit:            uint32(account.Unit),
		Code:            uint32(account.Code),
		DebitsReserved:  account.DebitsReserved,
		DebitsAccepted:  account.DebitsAccepted,
		CreditsReserved: account.CreditsReserved,
		CreditsAccepted: account.CreditsAccepted,
	}, nil
}

func (s *rpcServer) GetAccount(ctx context.Context, req *pacioliv1.GetAccountRequest) (*pacioliv1.Account, error) {
	account, err := s.ps.GetAccount(req.GetId())
	if err != nil {
		switch err.(type) {
		case *pacioli.ErrNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			// TODO: exhaustive switch on err
			return nil, status.Error(codes.Internal, "Failed to get account.")
		}
	}

	return &pacioliv1.Account{
		Id:              account.ID,
		LedgerId:        account.LedgerID,
		Unit:            uint32(account.Unit),
		Code:            uint32(account.Code),
		DebitsReserved:  account.DebitsReserved,
		DebitsAccepted:  account.DebitsAccepted,
		CreditsReserved: account.CreditsReserved,
		CreditsAccepted: account.CreditsAccepted,
	}, nil
}

func (s *rpcServer) CreateTransfer(ctx context.Context, req *pacioliv1.CreateTransferRequest) (*pacioliv1.Transfer, error) {
	transfer, err := s.ps.CreateTransfer(pacioli.CreateTransferArgs{
		Amount:            req.GetAmount(),
		DebitAccountID:    req.GetDebitAccountId(),
		CreditAccountID:   req.GetCreditAccountId(),
		TransactionTypeID: req.GetTransactionTypeId(),
	})
	if err != nil {
		// TODO: switch on err
		return nil, status.Error(codes.Internal, "Failed to create transfer.")
	}

	return &pacioliv1.Transfer{
		Id:              transfer.ID,
		DebitAccountId:  transfer.DebitAccountID,
		CreditAccountId: transfer.CreditAccountID,
		Amount:          transfer.Amount,
	}, nil
}
