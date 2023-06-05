package workflows

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/providers/tabapay"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var tabapayBucketName = "tabapayreports"

// ProcessReports reads files in the TabaPay S3 bucket. Processes each individual report
func ProcessReports(ctx workflow.Context, args tabapay.CreateCardArgs) error {
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
	err := workflow.ExecuteActivity(ctx, a.b).Get(ctx, &reports)
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
		} else {
			logger.Error("Unhandled Report", "report_file", r)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func computeHash(line []string, filename string) (string, error) {
	// Compute the line hash so we don't insert duplicates
	var buf bytes.Buffer
	buf.WriteString(filename)
	for _, le := range line {
		_, err := buf.WriteString(le)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", md5.Sum(buf.Bytes())), nil
}

func (a *Activity) ProcessChargebacksReports(ctx context.Context, filename string) error {
	data, err := a.b.AWS().S3GetObjectData(ctx, tabapayBucketName, filename)
	if err != nil {
		return err
	}
	defer data.Close()

	csvReader := csv.NewReader(data)
	var i int
	for {
		i++
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// Ignore the heading column
		if i == 1 {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		exceptionDate, _ := time.Parse("02-01-2006", line[7])
		statusDate, _ := time.Parse("02-01-2006", line[9])
		origCreatedDate, _ := time.Parse("02-01-2006", line[12])
		origProcessedDate, _ := time.Parse("02-01-2006", line[12])
		settledAmount, _ := strconv.Atoi(line[19])
		exceptionSettledAmount, _ := strconv.Atoi(line[20])
		tabapayFee, _ := strconv.Atoi(line[20])
		networkFee, _ := strconv.Atoi(line[21])

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO tabapay_report_chargebacks "+
			"(hash, filename, iso, mid, merchant_ref, transaction_id, exception_id, "+
			"exception_code, exception_decription, exception_date, action_status, "+
			"status_date, days_open, network_transaction_id, original_date_created, "+
			"original_date_processed, original_transaction_type, exception_source, "+
			"exception_destination, exception_network, last_four, original_settled_amount, "+
			"exception_settled_amount, tabapay_fee, network_fee, interchange, memo, cb_id, "+
			"first_name, last_name, network_id) "+
			"VALUES "+
			"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, "+
			"$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31) "+
			"ON CONFLICT DO NOTHING",
			lineHash, filename, line[0], line[1], line[2], line[3], line[4],
			line[5], line[6], exceptionDate, line[8],
			statusDate, line[10], line[11], origCreatedDate,
			origProcessedDate, line[14], line[15],
			line[16], line[17], line[18], settledAmount,
			exceptionSettledAmount, tabapayFee, networkFee, line[23], line[24], line[25],
			line[26], line[27], line[28])
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) GetNewReportNames(ctx context.Context) ([]string, error) {
	// Get all files from S3
	pl := a.b.AWS().S3ListObjects(tabapayBucketName)

	// Load all files and mark them as "unprocessed" i.e. false
	var s3Files map[string]bool

	for pl.HasMorePages() {
		page, err := pl.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			s3Files[*obj.Key] = false
		}
	}

	// Now all files from
	var processed []string
	err := a.b.DB().GetContext(ctx, &processed, "SELECT filename FROM tabapay_report_files")
	if err != nil {
		return nil, err
	}

	for _, p := range processed {
		s3Files[p] = true
	}

	var unprocessed []string

	for fn, p := range s3Files {
		if p {
			continue
		}

		unprocessed = append(unprocessed, fn)
	}

	return unprocessed, nil
}
