package admin

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	adminv1 "gitlab.com/fynbos/proto/backend/admin/v1"
)

type validateStatementsArgs struct {
	Period   string `validate:"required"`
	WalletID string `validate:"required,uuid"`
}

func (s *AdminRpcService) EmailWalletStatement(
	ctx context.Context, req *adminv1.EmailWalletStatementRequest,
) (*adminv1.Empty, error) {
	if err := s.b.Validator().StructCtx(ctx, &validateStatementsArgs{
		Period:   req.GetPeriod(),
		WalletID: req.GetWalletID(),
	}); err != nil {
		return nil, toGRPCError(err)
	}

	// look up wallet linked account to get the machnet provider id
	las, err := s.b.LinkedAccounts().ListByWalletId(ctx, req.GetWalletID())
	if err != nil {
		return nil, toGRPCError(err)
	}

	var walletLinkedAccount *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == machnet.ProviderName && la.Type == machnet.TypeWallet {
			walletLinkedAccount = &la
			break
		}
	}

	if walletLinkedAccount == nil {
		return nil, NotFoundError("Machnet wallet found.")
	}

	pdf, err := s.b.Machnet().GetStatement(ctx, walletLinkedAccount.ProviderID, req.GetPeriod())
	if err != nil {
		return nil, toGRPCError(err)
	}

	periodStart, err := time.Parse("2006-01-02", req.GetPeriod())
	if err != nil {
		return nil, toGRPCError(err)
	}

	err = s.b.Email().SendMailTemplate(ctx, req.GetWalletID(), email.StatementTemplateID,
		map[string]interface{}{
			"period":  fmt.Sprintf("%s -%s", periodStart.Format("02 Jan 2006"), periodStart.AddDate(0, 1, 0).Format("02 Jan 2006")),
			"subject": email.StatementTemplateID.Subject(),
		},
		[]email.Attachment{
			{
				Content:     pdf,
				ContentType: "application/pdf",
				Name:        "statement.pdf",
			},
		},
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &adminv1.Empty{}, nil
}
