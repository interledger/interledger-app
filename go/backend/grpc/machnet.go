package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/currency"

	"go.temporal.io/api/enums/v1"

	"gitlab.com/fynbos/backend/db"
	temporal_utils "gitlab.com/fynbos/backend/temporal/utils"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) GetMachnetWidgetToken(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.MachnetWidgetToken, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	token, err := s.b.Machnet().GetWidgetToken(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.MachnetWidgetToken{
		Value:            token.Value,
		ExpiresInMinutes: int64(token.ExpiresInMinutes),
		UserId:           token.UserID,
	}, nil
}

func (s *rpcService) StartMachnetKYC(
	ctx context.Context, _ *backendv1.Empty,
) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = s.b.Machnet().StartSendUserKYC(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}

func (s *rpcService) HasSendUser(
	ctx context.Context, req *backendv1.Empty,
) (*backendv1.HasSendUserResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	_, err = s.b.Machnet().GetUserByWalletID(ctx, wallet.ID)
	if err != nil {
		if errors.Is(err, machnet.ErrNotFound) {
			return &backendv1.HasSendUserResponse{
				HasSendUser: false,
			}, nil
		}
		return nil, toGRPCError(err)
	}

	return &backendv1.HasSendUserResponse{
		HasSendUser: true,
	}, nil
}

func (s *rpcService) KYCStatus(ctx context.Context, _ *backendv1.Empty) (*backendv1.KYCStatusResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthorized.")
	}

	kyc, err := s.b.Machnet().GetKYCStatus(ctx, wallet.ID)
	if errors.Is(err, machnet.ErrNotFound) {
		// Check the temporal workflow exists to create a KYC user
		wfd, err := s.b.Temporal().DescribeWorkflowExecution(ctx, "machnet_create_send_user_"+wallet.ID, "")
		if temporal_utils.IsNotFoundError(err) {
			return &backendv1.KYCStatusResponse{
				HasSendUser: false,
			}, nil
		}
		if err != nil {
			return nil, toGRPCError(err)
		}

		// If the workflow is in any of these states it failed and needs to be retried.
		switch wfd.GetWorkflowExecutionInfo().GetStatus() {
		case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
			enums.WORKFLOW_EXECUTION_STATUS_FAILED,
			enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
			enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
			enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
			return &backendv1.KYCStatusResponse{
				HasSendUser: false,
			}, nil
		}

		return &backendv1.KYCStatusResponse{
			HasSendUser: true,
			KycStatus:   int32(machnet.KYCStatusInProgress),
		}, nil
	}
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.KYCStatusResponse{
		HasSendUser:  true,
		KycStatus:    int32(kyc.User.KYCStatus),
		FailedFields: kyc.FailedFields,
	}, nil
}

func (s *rpcService) CreateWallet(
	ctx context.Context, req *backendv1.CreateWalletRequest,
) (*backendv1.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if err = s.b.Validator().VarCtx(ctx, req.GetNickname(), "required"); err != nil {
		return nil, toGRPCError(err)
	}

	sendUser, err := s.b.Machnet().GetUserByWalletID(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	la, err := s.b.Machnet().CreateWallet(ctx, machnet.CreateWalletArgs{
		Nickname:   req.GetNickname(),
		SendUserID: sendUser.ID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.LinkedAccount{
		Id:   la.ID,
		Type: la.Type,
		Name: la.Name,
		Mask: la.Mask,
	}, nil
}

func (s *rpcService) GetWalletBalance(ctx context.Context, _ *backendv1.Empty) (*backendv1.WalletBalance, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, ForbiddenError("Unauthenticated.")
	}

	lal, err := s.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	var found *linkedaccounts.LinkedAccount
	for _, la := range lal {
		if la.Provider != machnet.ProviderName || la.Type != machnet.TypeWallet {
			continue
		}
		found = &la
		break
	}
	if found == nil {
		return nil, NotFoundError("machnet wallet not found")
	}

	mw, err := s.b.Machnet().GetWallet(ctx, found.ProviderID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.WalletBalance{
		Balance:   mw.Balance,
		Available: mw.AvailableBalance,
	}, nil
}

type validateWithdrawFromMachnetWalletArgs struct {
	ToLinkedAccount string `validate:"required,uuid"`
	Amount          uint64 `validate:"gt=0"`
	IpAddress       string `validate:"ip_addr"`
}

func (s *rpcService) StartWithdrawFromMachnetWallet(
	ctx context.Context, req *backendv1.WithdrawFromMachnetWalletRequest,
) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if err = s.b.Validator().StructCtx(ctx, validateWithdrawFromMachnetWalletArgs{
		ToLinkedAccount: req.GetToLinkedAccountId(),
		Amount:          req.GetAmount(),
		IpAddress:       req.GetIpAddress(),
	}); err != nil {
		return nil, toGRPCError(err)
	}

	toLinkedAcc, err := s.b.LinkedAccounts().Get(ctx, req.GetToLinkedAccountId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if toLinkedAcc.WalletID != wallet.ID {
		return nil, NotFoundError("Linked account not found.")
	}

	linkedAccounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	var linkedWallet *linkedaccounts.LinkedAccount
	for _, la := range linkedAccounts {
		if la.Provider == machnet.ProviderName && la.Type == machnet.TypeWallet {
			linkedWallet = &la
			break
		}
	}
	if linkedWallet == nil {
		return nil, toGRPCError(errors.New("Machnet wallet not found."))
	}

	await, err := s.b.Machnet().WithdrawFromWallet(ctx, machnet.WithdrawFromWalletArgs{
		IdempotencyKey:        req.IdempotencyKey,
		WalletID:              wallet.ID,
		Amount:                req.GetAmount(),
		WalletLinkedAccountID: linkedWallet.ID,
		ToLinkedAccountID:     req.GetToLinkedAccountId(),
		IpAddress:             req.GetIpAddress(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = await(ctx, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}

type validateStartMachnetWalletTopupArgs struct {
	FromLinkedAccount string `validate:"required,uuid"`
	Amount            uint64 `validate:"gt=0"`
	IpAddress         string `validate:"ip_addr"`
	Currency          string `validate:"iso4217"`
}

func (s *rpcService) StartMachnetWalletTopup(
	ctx context.Context, req *backendv1.StartMachnetWalletTopupRequest,
) (*backendv1.Empty, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	if err = s.b.Validator().StructCtx(ctx, validateStartMachnetWalletTopupArgs{
		FromLinkedAccount: req.GetFromLinkedAccountId(),
		Amount:            req.GetAmount(),
		IpAddress:         req.GetIpAddress(),
		Currency:          req.GetCurrency(),
	}); err != nil {
		return nil, toGRPCError(err)
	}

	fromLinkedAcc, err := s.b.LinkedAccounts().Get(ctx, req.GetFromLinkedAccountId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	if fromLinkedAcc.WalletID != wallet.ID {
		return nil, NotFoundError("Linked account not found.")
	}
	if fromLinkedAcc.Type != machnet.TypeSendCard {
		return nil, InternalError("Cannot fund wallet from this linked account.")
	}

	linkedAccounts, err := s.b.LinkedAccounts().ListByWalletId(ctx, wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	var linkedWallet *linkedaccounts.LinkedAccount
	for _, la := range linkedAccounts {
		if la.Provider == machnet.ProviderName && la.Type == machnet.TypeWallet {
			linkedWallet = &la
			break
		}
	}
	if linkedWallet == nil {
		return nil, toGRPCError(errors.New("Machnet wallet not found."))
	}

	await, err := s.b.Machnet().StartWalletTopup(ctx, machnet.StartWalletTopupArgs{
		IdempotencyKey:        req.IdempotencyKey,
		WalletID:              wallet.ID,
		Amount:                currency.FromUInt64(req.GetAmount(), currency.ParseCurrency(req.GetCurrency())),
		WalletLinkedAccountID: linkedWallet.ID,
		FromLinkedAccountID:   fromLinkedAcc.ID,
		IpAddress:             req.GetIpAddress(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = await(ctx, nil)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.Empty{}, nil
}

func (s *rpcService) ListStatements(
	ctx context.Context, req *backendv1.PaginationRequest,
) (*backendv1.ListStatementsResponse, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	statementPeriods, err := s.b.Machnet().ListStatementPeriods(ctx, db.PaginationFromPB(req), wallet.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.ListStatementsResponse{
		Periods: statementPeriods,
	}, nil
}

func (s *rpcService) GetStatementPDF(
	ctx context.Context, req *backendv1.GetStatementPDFRequest,
) (*backendv1.StatementPDF, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	wallet, err := s.b.Users().WalletForContext(ctx)
	if err != nil {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	pdf, err := s.b.Machnet().GetStatement(ctx, wallet.ID, req.GetPeriod())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &backendv1.StatementPDF{
		Chunks: pdf,
	}, nil
}
