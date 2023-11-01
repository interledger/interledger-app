package grpc

import (
	"context"
	"errors"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"gitlab.com/fynbos/backend/providers/xago"

	"gitlab.com/fynbos/backend/user"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

func (s *rpcService) AddXagoBankAccount(ctx context.Context, req *pb.AddXagoBankAccountRequest) (*pb.LinkedAccount, error) {
	_, err := s.b.Users().UserForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, UnauthenticatedError("Unauthenticated.")
	}

	w, err := s.b.Wallets().ForContext(ctx)
	if err != nil && !errors.Is(err, user.ErrNoUserFound) {
		return nil, ForbiddenError("Unauthenticated.")
	}

	await, err := s.b.Xago().CreateBeneficiary(ctx, xago.CreateBankAccountArgs{
		WalletID:      w.ID,
		AccountNumber: req.AccountNumber,
		BranchCode:    req.BranchCode,
		BankName:      req.BankName,
		IBAN:          req.Iban,
		BIC:           req.Bic,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	var la linkedaccounts.LinkedAccount
	err = await(ctx, &la)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return transformLinkedAccount(la), nil
}
