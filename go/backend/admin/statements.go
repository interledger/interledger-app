package admin

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/email"
	adminv1 "gitlab.com/fynbos/proto/backend/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *AdminRpcService) EmailWalletStatement(
	ctx context.Context, req *adminv1.EmailWalletStatementRequest,
) (*adminv1.Empty, error) {
	pdf, err := s.b.Machnet().GetStatement(ctx, req.GetWalletID(), req.GetPeriod())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	periodStart, err := time.Parse("2006-01-02", req.GetPeriod())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &adminv1.Empty{}, nil
}
