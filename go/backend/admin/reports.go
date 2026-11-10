package admin

import (
	"context"
	"time"

	"github.com/interledger/interledger-app/go/backend/reports"
	adminv1 "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
)

// GetXagoUsersReport returns the Xago and South African users report as a CSV or PDF download.
func (s *AdminRpcService) GetXagoUsersReport(ctx context.Context, req *adminv1.UsersReportRequest) (*adminv1.ReportFile, error) {
	return reportFile(reports.XagoUsers(ctx, s.b, reportFormat(req.GetFormat()), time.Now()))
}

// GetGatehubUsersReport returns the GateHub users report as a CSV or PDF download.
func (s *AdminRpcService) GetGatehubUsersReport(ctx context.Context, req *adminv1.UsersReportRequest) (*adminv1.ReportFile, error) {
	return reportFile(reports.GatehubUsers(ctx, s.b, reportFormat(req.GetFormat()), time.Now()))
}

func reportFormat(f adminv1.ReportFormat) reports.Format {
	if f == adminv1.ReportFormat_REPORT_FORMAT_PDF {
		return reports.FormatPDF
	}
	return reports.FormatCSV
}

func reportFile(f *reports.File, err error) (*adminv1.ReportFile, error) {
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &adminv1.ReportFile{
		Filename:    f.Filename,
		Content:     f.Content,
		ContentType: f.ContentType,
	}, nil
}
