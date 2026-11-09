package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/interledger/interledger-app/go/backend/reports"
	adminv1 "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetXagoUsersReport returns the Xago users report as a CSV download.
func (s *AdminRpcService) GetXagoUsersReport(ctx context.Context, _ *emptypb.Empty) (*adminv1.CsvReport, error) {
	csv, err := reports.XagoUsersCSV(ctx, s.b)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &adminv1.CsvReport{
		Filename: fmt.Sprintf("xago-users-%s.csv", time.Now().UTC().Format("2006-01-02")),
		Content:  []byte(csv),
	}, nil
}
