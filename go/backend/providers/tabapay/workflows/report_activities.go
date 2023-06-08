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
)

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

		exceptionDate, _ := time.Parse("02-01-2006", line[8])
		statusDate, _ := time.Parse("02-01-2006", line[10])
		origCreatedDate, _ := time.Parse("02-01-2006", line[13])
		origProcessedDate, _ := time.Parse("02-01-2006", line[14])
		settledAmount, _ := strconv.ParseFloat(line[20], 64)
		exceptionSettledAmount, _ := strconv.ParseFloat(line[21], 64)
		tabapayFee, _ := strconv.ParseFloat(line[22], 64)
		networkFee, _ := strconv.ParseFloat(line[23], 64)
		daysOpen, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(line[11], "days", ""), "day", "")), 64)

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO tabapay_report_chargebacks "+
			"(hash, filename, iso, mid, merchant_ref, transaction_id, exception_id, exception_type, "+
			"exception_code, exception_description, exception_date, action_status, "+
			"status_date, days_open, network_transaction_id, original_date_created, "+
			"original_date_processed, original_transaction_type, exception_source, "+
			"exception_destination, exception_network, last_four, original_settled_amount, "+
			"exception_settled_amount, tabapay_fee, network_fee, interchange, memo, cb_id, "+
			"first_name, last_name, network_id) "+
			"VALUES "+
			"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, "+
			"$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32) "+
			"ON CONFLICT DO NOTHING",
			lineHash, filename, line[0], line[1], line[2], line[3], line[4], line[5],
			line[6], line[7], exceptionDate, line[9],
			statusDate, daysOpen, line[12], origCreatedDate,
			origProcessedDate, line[15], line[16],
			line[17], line[18], line[19], settledAmount*10,
			exceptionSettledAmount*10, tabapayFee*10, networkFee*10, line[24], line[25], line[26],
			line[27], line[28], line[29])
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) ProcessAMLTransactionsReport(ctx context.Context, filename string) error {
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

		settleDate, _ := time.Parse("20060102", line[11])
		reportDate, _ := time.Parse("01/02/2006", line[16])
		txAmount, _ := strconv.ParseFloat(line[12], 64)
		settleAmount, _ := strconv.ParseFloat(line[13], 64)

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO tabapay_report_aml_transaction "+
			"(hash, filename, aml_id, aml_code, aml_description, iso, iso_name, mid, "+
			"merchant_name, caid, bin_last_four, transaction_type, transaction_id,"+
			"settle_date, transaction_amount, settle_amount, fn, ln, report_date, avs,"+
			"cvv_cav, type) "+
			"VALUES "+
			"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, "+
			"$18, $19, $20, $21, $22) "+
			"ON CONFLICT DO NOTHING",
			lineHash, filename, line[0], line[1], line[2], line[3], line[4], line[5],
			line[6], line[7], line[8], line[9], line[10],
			settleDate, txAmount*10, settleAmount*10, line[14], line[15], reportDate, line[17],
			line[18], line[19],
		)
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
