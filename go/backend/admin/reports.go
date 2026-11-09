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
	return csvReport(ctx, "xago-users", func(ctx context.Context) (string, error) {
		return reports.XagoUsersCSV(ctx, s.b)
	})
}

// GetGatehubUsersReport returns the GateHub users report as a CSV download.
func (s *AdminRpcService) GetGatehubUsersReport(ctx context.Context, _ *emptypb.Empty) (*adminv1.CsvReport, error) {
	return csvReport(ctx, "gatehub-users", func(ctx context.Context) (string, error) {
		return reports.GatehubUsersCSV(ctx, s.b)
	})
}

func csvReport(ctx context.Context, name string, build func(context.Context) (string, error)) (*adminv1.CsvReport, error) {
	csv, err := build(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &adminv1.CsvReport{
		Filename: fmt.Sprintf("%s-%s.csv", name, time.Now().UTC().Format("2006-01-02")),
		Content:  []byte(csv),
	}, nil
}
