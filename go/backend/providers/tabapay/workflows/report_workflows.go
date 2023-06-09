package workflows

import (
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var tabapayBucketName = "tabapayreports"

// ProcessReports reads files in the TabaPay S3 bucket. Processes each individual report
func ProcessReports(ctx workflow.Context) error {
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
		// TODO: Add more reports
		if strings.Contains(r, "chargebacks") {
			err = workflow.ExecuteActivity(ctx, a.ProcessChargebacksReports, r).Get(ctx, nil)
		} else if strings.Contains(r, "AMLtransactions") {
			err = workflow.ExecuteActivity(ctx, a.ProcessAMLTransactionsReport, r).Get(ctx, nil)
		} else if strings.Contains(r, "AMLSummary") {
			err = workflow.ExecuteActivity(ctx, a.ProcessAMLSummaryReport, r).Get(ctx, nil)
		} else if strings.Contains(r, "exceptions") {
			err = workflow.ExecuteActivity(ctx, a.ProcessExceptionsReports, r).Get(ctx, nil)
		} else {
			logger.Error("Unhandled Report", "report_file", r)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
