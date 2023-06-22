package workflows

import (
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var tabapayBucketName = "fynbos-tabapay"

// ProcessReportsWorkflow reads files in the TabaPay S3 bucket. Processes each individual report
func ProcessReportsWorkflow(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Checking for new tabapay reports to process")

	var reports []string
	err := workflow.ExecuteActivity(ctx, a.GetNewReportNames).Get(ctx, &reports)
	if err != nil {
		return err
	}

	if len(reports) == 0 {
		return nil
	}

	for _, r := range reports {
		// Load the report based on the filename
		if strings.Contains(r, "Monthly") {
			if strings.Contains(r, "invoice") && strings.HasSuffix(r, ".xlsx") {
				err = workflow.ExecuteActivity(ctx, a.MailReport, r).Get(ctx, nil)
			} else if strings.Contains(r, "interchange") {
				err = workflow.ExecuteActivity(ctx, a.ProcessInterchangeReport, r, "tabapay_report_monthly_interchange").Get(ctx, nil)
			} else if strings.Contains(r, "transactions") {
				err = workflow.ExecuteActivity(ctx, a.ProcessTransactionsReport, r, "tabapay_monthly_report_transactions").Get(ctx, nil)
			} else if strings.Contains(r, "networkfees") {
				err = workflow.ExecuteActivity(ctx, a.ProcessMonthlyNetworkFees, r).Get(ctx, nil)
			} else if strings.Contains(r, "processingfees") {
				err = workflow.ExecuteActivity(ctx, a.ProcessMonthlyProcessingFee, r).Get(ctx, nil)
			} else {
				logger.Error("Unhandled Monthly Report", "report_file", r)
				continue
			}
		} else if strings.Contains(r, "chargebacks") {
			err = workflow.ExecuteActivity(ctx, a.ProcessChargebacksReports, r).Get(ctx, nil)
		} else if strings.Contains(r, "AMLtransactions") {
			err = workflow.ExecuteActivity(ctx, a.ProcessAMLTransactionsReport, r).Get(ctx, nil)
		} else if strings.Contains(r, "AMLSummary") {
			err = workflow.ExecuteActivity(ctx, a.ProcessAMLSummaryReport, r).Get(ctx, nil)
		} else if strings.Contains(r, "exceptions") {
			err = workflow.ExecuteActivity(ctx, a.ProcessExceptionsReports, r).Get(ctx, nil)
		} else if strings.Contains(r, "interchange") {
			err = workflow.ExecuteActivity(ctx, a.ProcessInterchangeReport, r, "tabapay_report_interchange").Get(ctx, nil)
		} else if strings.Contains(r, "summary") {
			err = workflow.ExecuteActivity(ctx, a.ProcessSummaryReport, r).Get(ctx, nil)
		} else if strings.Contains(r, "transactions") {
			err = workflow.ExecuteActivity(ctx, a.ProcessTransactionsReport, r, "tabapay_report_transactions").Get(ctx, nil)
		} else {
			logger.Error("Unhandled Report", "report_file", r)
			continue
		}

		if err != nil {
			return err
		}

		// Insert the filename to mark it as processed
		err = workflow.ExecuteActivity(ctx, a.MarkReportAsProcessed, r).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
