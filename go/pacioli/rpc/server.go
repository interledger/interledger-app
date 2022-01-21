package rpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	ledger "gitlab.com/fynbos/pacioli/ledger"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

func NewServer(ps ledger.Service) *grpc.Server {
	server := grpc.NewServer()
	pacioliv1.RegisterPacioliServiceServer(server, &rpcServer{ledger: ps})
	reflection.Register(server)
	return server
}

type rpcServer struct {
	pacioliv1.UnimplementedPacioliServiceServer
	ledger ledger.Service
}

func (s *rpcServer) CreateLedger(ctx context.Context, req *pacioliv1.CreateLedgerRequest) (*pacioliv1.Ledger, error) {
	ledger, err := s.ledger.CreateLedger(req.GetName(), uint16(req.GetCode()))
	if err != nil {
		// TODO: switch on err
		return nil, status.Error(codes.Internal, "Failed to create ledger.")
	}

	return &pacioliv1.Ledger{
		Id:   ledger.ID,
		Name: ledger.Name,
		Code: uint32(ledger.Code),
	}, nil
}

func (s *rpcServer) CreateAccount(ctx context.Context, req *pacioliv1.CreateAccountRequest) (*pacioliv1.Account, error) {
	accountId := uuid.NewString()
	errors, err := s.ledger.CreateAccounts(req.GetLedgerID(), []ledger.CreateAccountArgs{
		{
			ID:   accountId,
			Code: uint16(req.GetCode()),
		},
	})
	// This error will be due to connection / io / validation  issues
	if err != nil {
		switch err.(type) {
		case ledger.ErrInvalidArg:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "Failed to create account.")
		}
	}
	// This will be application errors
	if len(errors) != 0 {
		// TODO: exhaustive switch on err
		return nil, status.Error(codes.Internal, "Failed to create account.")
	}

	accounts, err := s.ledger.GetAccounts(req.GetLedgerID(), []string{accountId})
	if err != nil || len(accounts) != 1 {
		return nil, status.Error(codes.Internal, "Failed to create account.")
	}

	return &pacioliv1.Account{
		Id:              accountId,
		LedgerCode:      uint32(accounts[0].LedgerCode),
		Code:            uint32(accounts[0].Code),
		DebitsReserved:  accounts[0].DebitsReserved,
		DebitsAccepted:  accounts[0].DebitsAccepted,
		CreditsReserved: accounts[0].CreditsReserved,
		CreditsAccepted: accounts[0].CreditsAccepted,
	}, nil
}

func (s *rpcServer) GetAccount(ctx context.Context, req *pacioliv1.GetAccountRequest) (*pacioliv1.Account, error) {
	accounts, err := s.ledger.GetAccounts(req.GetLedgerID(), []string{req.GetId()})
	// This error will be due to connection / io / validation  issues
	if err != nil {
		switch err.(type) {
		case ledger.ErrInvalidArg:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case ledger.ErrNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "Failed to get account.")
		}
	}
	if len(accounts) != 1 {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pacioliv1.Account{
		Id:              accounts[0].ID,
		LedgerCode:      uint32(accounts[0].LedgerCode),
		Code:            uint32(accounts[0].Code),
		DebitsReserved:  accounts[0].DebitsReserved,
		DebitsAccepted:  accounts[0].DebitsAccepted,
		CreditsReserved: accounts[0].CreditsReserved,
		CreditsAccepted: accounts[0].CreditsAccepted,
	}, nil
}

func (s *rpcServer) CreateTransfer(ctx context.Context, req *pacioliv1.CreateTransferRequest) (*pacioliv1.Transfer, error) {
	transfer, err := s.ledger.CreateTransfer(ledger.CreateTransferArgs{
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
