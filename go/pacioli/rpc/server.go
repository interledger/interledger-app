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

func (s *rpcServer) CreateTenant(ctx context.Context, req *pacioliv1.CreateTenantRequest) (*pacioliv1.Tenant, error) {
	tenant, err := s.ps.CreateTenant(req.GetIdentifier())
	if err != nil {
		// TODO: switch on err
		return nil, status.Error(codes.Internal, "Failed to create tenant.")
	}

	return &pacioliv1.Tenant{
		Id:         tenant.ID,
		Identifier: tenant.Identifier,
	}, nil
}

func (s *rpcServer) CreateLedger(ctx context.Context, req *pacioliv1.CreateLedgerRequest) (*pacioliv1.Ledger, error) {
	ledger, err := s.ps.CreateLedger(req.GetTenantId(), req.GetName())
	if err != nil {
		// TODO: switch on err
		return nil, status.Error(codes.Internal, "Failed to create ledger.")
	}

	return &pacioliv1.Ledger{
		Id:   ledger.ID,
		Name: ledger.Name,
	}, nil
}

func (s *rpcServer) CreateAccountCategory(ctx context.Context, req *pacioliv1.CreateAccountCategoryRequest) (*pacioliv1.AccountCategory, error) {
	category, err := s.ps.CreateAccountCategory(req.GetTenantId(), pacioli.AccountCategoryArgs{
		Name:        req.GetName(),
		Type:        req.GetType(),
		Description: req.GetDescription(),
		Code:        uint16(req.GetCode()),
	})
	if err != nil {
		// TODO: switch on err
		return nil, status.Error(codes.Internal, "Failed to create account category.")
	}

	return &pacioliv1.AccountCategory{
		Id:          category.ID,
		Name:        category.Name,
		Type:        category.Type,
		Description: category.Description,
		Code:        uint32(category.Code),
	}, nil
}

func (s *rpcServer) GetAccountCategoryByCode(ctx context.Context, req *pacioliv1.GetAccountCategoryByCodeRequest) (*pacioliv1.AccountCategory, error) {
	category, err := s.ps.GetAccountCategoryByCode(req.GetTenantId(), uint16(req.GetCode()))
	if err != nil {
		switch err.(type) {
		case *pacioli.ErrNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			// TODO: exhaustive switch on err
			return nil, status.Error(codes.Internal, "Failed to get account category.")
		}
	}

	return &pacioliv1.AccountCategory{
		Id:          category.ID,
		Name:        category.Name,
		Type:        category.Type,
		Description: category.Description,
		Code:        uint32(category.Code),
	}, nil
}

func (s *rpcServer) CreateAccount(ctx context.Context, req *pacioliv1.CreateAccountRequest) (*pacioliv1.Account, error) {
	account, err := s.ps.CreateAccount(req.GetTenantId(), pacioli.CreateAccountArgs{
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
	account, err := s.ps.GetAccount(req.GetTenantId(), req.GetId())
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

func (s *rpcServer) CreateTransactionType(ctx context.Context, req *pacioliv1.CreateTransactionTypeRequest) (*pacioliv1.TransactionType, error) {
	transactionType, err := s.ps.CreateTransactionType(req.GetTenantId(), pacioli.TransactionTypeArgs{
		Name:                      req.GetName(),
		Description:               req.GetDescription(),
		CreditAccountCategoryCode: uint16(req.GetCreditAccountCategoryCode()),
		DebitAccountCategoryCode:  uint16(req.GetDebitAccountCategoryCode()),
	})
	if err != nil {
		// TODO: switch on err
		return nil, status.Error(codes.Internal, "Failed to create transaction type.")
	}

	return &pacioliv1.TransactionType{
		Id:                        transactionType.ID,
		Name:                      transactionType.Name,
		Description:               transactionType.Description,
		CreditAccountCategoryCode: uint32(transactionType.CreditAccountCategoryCode),
		DebitAccountCategoryCode:  uint32(transactionType.DebitAccountCategoryCode),
	}, nil
}

func (s *rpcServer) GetTransactionType(ctx context.Context, req *pacioliv1.GetTransactionTypeRequest) (*pacioliv1.TransactionType, error) {
	transactionType, err := s.ps.GetTransactionType(req.GetTenantId(), req.GetId())
	if err != nil {
		switch err.(type) {
		case *pacioli.ErrNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			// TODO: exhaustive switch on err
			return nil, status.Error(codes.Internal, "Failed to get transaction type.")
		}
	}

	return &pacioliv1.TransactionType{
		Id:                        transactionType.ID,
		Name:                      transactionType.Name,
		Description:               transactionType.Description,
		CreditAccountCategoryCode: uint32(transactionType.CreditAccountCategoryCode),
		DebitAccountCategoryCode:  uint32(transactionType.DebitAccountCategoryCode),
	}, nil
}

func (s *rpcServer) CreateTransfer(ctx context.Context, req *pacioliv1.CreateTransferRequest) (*pacioliv1.Transfer, error) {
	transfer, err := s.ps.CreateTransfer(req.GetTenantId(), pacioli.CreateTransferArgs{
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
		Id:                transfer.ID,
		DebitAccountId:    transfer.DebitAccountID,
		CreditAccountId:   transfer.CreditAccountID,
		TransactionTypeId: transfer.TransactionTypeID,
		Amount:            transfer.Amount,
	}, nil
}
