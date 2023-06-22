package ops

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var gmtBucketName = "fynbos-gmt"

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
		err = workflow.ExecuteActivity(ctx, a.ProcessDailyReport, r).Get(ctx, nil)
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

func isEmptyLine(line []string) bool {
	if len(line) == 0 {
		return true
	}
	for _, le := range line {
		if strings.TrimSpace(le) != "" {
			return false
		}
	}

	return true
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

func parseReportDate(input string) time.Time {
	dt, err := time.Parse("1/2/2006 3:4:5 PM", input)
	if err == nil {
		return dt
	}

	return dt
}

func parseReportAmount(input string) float64 {
	res, _ := strconv.ParseFloat(strings.TrimSpace(input), 64)

	return res
}

func (a *Activity) ProcessDailyReport(ctx context.Context, filename string) error {
	data, err := a.b.AWS().S3GetObjectData(ctx, gmtBucketName, filename)
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
		if i == 1 || isEmptyLine(line) {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO gmt_daily_report "+
			"(hash, filename, receipt, third_party, transaction_date, "+
			"status, method_of_payment, correspondent, sender, money_sent, "+
			"total_amount, gmt_rate, agency_rate, agency_com, agency_fee, "+
			"gain_by_exchange_rate, fees) "+
			"VALUES "+
			"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) "+
			"ON CONFLICT DO NOTHING",
			lineHash, filename, line[0], line[1], parseReportDate(line[2]),
			line[3], line[4], line[5], line[6], parseReportAmount(line[7]),
			parseReportAmount(line[8]), parseReportAmount(line[9]), parseReportAmount(line[10]), parseReportAmount(line[11]), parseReportAmount(line[12]),
			parseReportAmount(line[13]), parseReportAmount(line[14]),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) GetNewReportNames(ctx context.Context) ([]string, error) {
	// Get all files from S3
	pl := a.b.AWS().S3ListObjects(gmtBucketName)

	// Load all files and mark them as "unprocessed" i.e. false
	s3Files := make(map[string]bool)

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
	err := a.b.DB().SelectContext(ctx, &processed, "SELECT filename FROM gmt_report_files")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
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

func (a *Activity) MarkReportAsProcessed(ctx context.Context, filename string) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO gmt_report_files (filename) VALUES ($1)", filename)
	return err
}
