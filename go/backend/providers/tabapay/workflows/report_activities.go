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
		if i == 1 || isEmptyLine(line) {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		exceptionDate := parseReportDate(line[8])
		statusDate := parseReportDate(line[10])
		origCreatedDate := parseReportDate(line[13])
		origProcessedDate := parseReportDate(line[14])
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
			line[17], line[18], line[19], settledAmount*100,
			exceptionSettledAmount*100, tabapayFee*100, networkFee*100, line[24], line[25], line[26],
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
		if i == 1 || isEmptyLine(line) {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		settleDate, _ := time.Parse("20060102", line[11])
		reportDate := parseReportDate(line[16])
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
			settleDate, txAmount*100, settleAmount*100, line[14], line[15], reportDate, line[17],
			line[18], line[19],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) ProcessAMLSummaryReport(ctx context.Context, filename string) error {
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
		if i == 1 || isEmptyLine(line) {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		reportDate := parseReportDate(line[12])
		txCount, _ := strconv.Atoi(line[10])
		amount, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line[11], "$")), 64)

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO tabapay_report_aml_summary "+
			"(hash, filename, aml_id, aml_code, aml_description, iso, iso_name, mid, "+
			"merchant_name, caid, bin_last_four, transaction_type, count,"+
			"total, report_date) "+
			"VALUES "+
			"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) "+
			"ON CONFLICT DO NOTHING",
			lineHash, filename, line[0], line[1], line[2], line[3], line[4], line[5],
			line[6], line[7], line[8], line[9], txCount,
			amount*100, reportDate,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) ProcessExceptionsReports(ctx context.Context, filename string) error {
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
		if i == 1 || isEmptyLine(line) {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		exceptionDate := parseReportDate(line[8])
		statusDate := parseReportDate(line[10])
		origCreatedDate := parseReportDate(line[13])
		origProcessedDate := parseReportDate(line[14])
		settledAmount, _ := strconv.ParseFloat(line[20], 64)
		exceptionSettledAmount, _ := strconv.ParseFloat(line[21], 64)
		tabapayFee, _ := strconv.ParseFloat(line[22], 64)
		networkFee, _ := strconv.ParseFloat(line[23], 64)
		daysOpen, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(line[11], "days", ""), "day", "")), 64)

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO tabapay_report_exceptions "+
			"(hash, filename, iso, mid, merchant_reference_id, original_transaction_id, exception_id, exception_type, "+
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
			line[17], line[18], line[19], settledAmount*100,
			exceptionSettledAmount*100, tabapayFee*100, networkFee*100, line[24], line[25], line[26],
			line[27], line[28], line[29])
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) ProcessInterchangeReport(ctx context.Context, filename, tableName string) error {
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
		if i == 1 || isEmptyLine(line) {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		txCount, _ := strconv.Atoi(strings.TrimSpace(line[7]))
		txDollars, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line[8], "$")), 64)
		interchangeDollars, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line[9], "$")), 64)

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO "+tableName+" "+
			"(hash, filename, iso, iso_name, mid, merchant_name, brand, card_type, "+
			"interchange_category, transaction_count, transaction_dollars, interchange_dollars) "+
			"VALUES "+
			"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) "+
			"ON CONFLICT DO NOTHING",
			lineHash, filename, line[0], line[1], line[2], line[3], line[4], line[5],
			line[6], txCount, txDollars*100, interchangeDollars*100)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) ProcessSummaryReport(ctx context.Context, filename string) error {
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
		if i == 1 || isEmptyLine(line) {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		txCount, _ := strconv.Atoi(strings.TrimSpace(line[4]))
		txAmount, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line[5], "$")), 64)
		sumDate := parseReportDate(line[2])

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO tabapay_report_summary "+
			"(hash, filename, iso, mid, type, summary_date, transactions_count, transactions_amount) "+
			"VALUES "+
			"($1, $2, $3, $4, $5, $6, $7, $8) "+
			"ON CONFLICT DO NOTHING",
			lineHash, filename, line[0], line[1], line[3], sumDate, txCount, txAmount*100)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) ProcessTransactionsReport(ctx context.Context, filename, tableName string) error {
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
		if i == 1 || isEmptyLine(line) {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		processedDate := parseReportDate(line[6])
		txDate := parseReportDate(line[7])
		settleDate := parseReportDate(line[33])
		ofacDate := parseReportDate(line[42])
		corOfacDate := parseReportDate(line[46])

		txAmount, _ := strconv.ParseFloat(line[15], 64)
		settleAmount, _ := strconv.ParseFloat(line[16], 64)
		tabapayFee, _ := strconv.ParseFloat(line[17], 64)
		networkFee, _ := strconv.ParseFloat(line[18], 64)
		interchange, _ := strconv.ParseFloat(line[19], 64)
		convenienceFee, _ := strconv.ParseFloat(line[20], 64)
		beneficiaryAmount, _ := strconv.ParseFloat(line[39], 64)

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO "+tableName+" "+
			"(hash, filename, iso, mid, reference_id, "+
			"transaction_id, corresponding_id, approval_code, processed_date, transaction_date, "+
			"type, source, destination, settlement_network, last_four, "+
			"status, network_rc, transaction_amount, settled_amount, tabapay_fee, "+
			"network_fee, interchange, convenience_fee, first_name, last_name, "+
			"memo, location_name, location_address_1, location_address_2, location_city, "+
			"location_state, location_zip, avs, cvv2, network_id, "+
			"settlement_date, card_brand, card_type, interchange_category, network_fee_codes, "+
			"bin, beneficiary_amount, beneficiary_currency, fx_rate_applied, ofac_date, "+
			"ofac_code, issuer_name, issuer_country, corr_ofac_date, corr_ofac_code, "+
			"corresponding_fn, corresponding_ln) "+
			"VALUES "+
			"($1, $2, $3, $4, $5, "+
			"$6, $7, $8, $9, $10, "+
			"$11, $12, $13, $14, $15, "+
			"$16, $17, $18, $19, $20, "+
			"$21, $22, $23, $24, $25, "+
			"$26, $27, $28, $29, $30, "+
			"$31, $32, $33, $34, $35, "+
			"$36, $37, $38, $39, $40, "+
			"$41, $42, $43, $44, $45, "+
			"$46, $47, $48, $49, $50, "+
			"$51, $52) "+
			"ON CONFLICT DO NOTHING",
			lineHash, filename, line[0], line[1], line[2],
			line[3], line[4], line[5], processedDate, txDate,
			line[8], line[9], line[10], line[11], line[12],
			line[13], line[14], txAmount*100, settleAmount*100, tabapayFee,
			networkFee, interchange*100, convenienceFee*100, line[21], line[22],
			line[23], line[24], line[25], line[26], line[27],
			line[28], line[29], line[30], line[31], line[32],
			settleDate, line[34], line[35], line[36], line[37],
			line[38], beneficiaryAmount*100, line[40], line[41], ofacDate,
			line[43], line[44], line[45], corOfacDate, line[47],
			line[48], line[49])

		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Activity) ProcessMonthlyProcessingFee(ctx context.Context, filename string) error {
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
		if i == 1 || isEmptyLine(line) {
			continue
		}

		// Compute the line hash so we don't insert duplicates
		lineHash, err := computeHash(line, filename)
		if err != nil {
			return err
		}

		quantity, _ := strconv.Atoi(strings.TrimSpace(line[5]))
		unitFee, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line[6], "$")), 64)
		fee, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line[7], "$")), 64)

		_, err = a.b.DB().ExecContext(ctx, "INSERT INTO tabapay_report_monthly_processing_fee "+
			"(hash, filename, iso, iso_name, mid, merchant_name, fee_category, quantity, unit_fee, fee) "+
			"VALUES "+
			"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) "+
			"ON CONFLICT DO NOTHING",
			lineHash, filename, line[0], line[1], line[2], line[3], line[4], quantity, unitFee*100, fee*100)
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

func parseReportDate(input string) time.Time {
	dt, err := time.Parse("01/02/2006", input)
	if err == nil {
		return dt
	}

	dt, err = time.Parse("01/2/2006", input)
	if err == nil {
		return dt
	}

	dt, err = time.Parse("1/02/2006", input)
	if err == nil {
		return dt
	}

	dt, err = time.Parse("1/2/2006", input)
	if err == nil {
		return dt
	}

	dt, err = time.Parse("1/2/2006 15:04", input)
	if err == nil {
		return dt
	}

	return dt
}
