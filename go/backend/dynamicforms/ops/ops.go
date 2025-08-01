package ops

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/dynamicforms"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/env"
)

func SubmitForm(ctx context.Context, b Backends, args *dynamicforms.SubmitArgs) (*dynamicforms.Submission, error) {
	var form dynamicforms.Submission

	jsonData, err := json.Marshal(args.Data)
	if err != nil {
		return nil, fmt.Errorf("%w failed to marshal form data: %w", dynamicforms.ErrInternal, err)
	}

	var walletId sql.NullString
	if args.WalletID != "" {
		walletId = sql.NullString{String: args.WalletID, Valid: true}
	}

	err = b.DB().GetContext(ctx, &form, "INSERT INTO dynamic_forms(form_id, data, wallet_id) VALUES($1, $2, $3) RETURNING id, form_id, data, wallet_id", args.FormID, jsonData, walletId)
	if err != nil {
		return nil, fmt.Errorf("%w failed to create form: %w", dynamicforms.ErrInternal, err)
	}

	if env.IsProd() {
		slack.SendToChannel(ctx, slack.ChannelNotifyForms, "wallet-info-bot", fmt.Sprintf(":incoming_envelope: New Form Submission: %s", args.FormID))
	}

	return &form, nil
}

func ListSubmissionCount(ctx context.Context, b Backends, _ db.Pagination) ([]dynamicforms.SubmissionCount, error) {
	var forms []dynamicforms.SubmissionCount

	err := b.DB().SelectContext(ctx, &forms, "SELECT form_id, COUNT(*) FROM dynamic_forms GROUP BY form_id")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", dynamicforms.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", dynamicforms.ErrInternal, err)
	}

	return forms, nil
}

func ListSubmissions(ctx context.Context, b Backends, formID string) ([]dynamicforms.Submission, error) {
	var submissions []dynamicforms.Submission

	err := b.DB().SelectContext(ctx, &submissions, "SELECT id, form_id, wallet_id, data, created_at  FROM dynamic_forms WHERE form_id = $1", formID)
	if errors.Is(err, sql.ErrNoRows) {
		return submissions, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", dynamicforms.ErrInternal, err)
	}

	return submissions, nil
}

func GetSubmission(ctx context.Context, b Backends, id string) (*dynamicforms.Submission, error) {
	var submission dynamicforms.Submission

	err := b.DB().GetContext(ctx, &submission, "SELECT id, form_id, wallet_id, data, created_at FROM dynamic_forms WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", dynamicforms.ErrNotFound, err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", dynamicforms.ErrInternal, err)
	}

	return &submission, nil
}

func ExportSubmissions(ctx context.Context, b Backends, formID string, writer io.Writer) error {
	var formResps []string

	err := b.DB().SelectContext(ctx, &formResps, "SELECT data FROM dynamic_forms WHERE form_id = $1", formID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w %s", dynamicforms.ErrNotFound, err)
	}
	if err != nil {
		return fmt.Errorf("%w %s", dynamicforms.ErrInternal, err)
	}

	var jsonStrings []string
	for _, jsonString := range formResps {
		formResp, err := strconv.Unquote(jsonString)
		if err != nil {
			return fmt.Errorf("%w %s %s", dynamicforms.ErrInternal, err, formResp)
		}

		jsonStrings = append(jsonStrings, formResp)
	}

	jsonStringArray := "[" + strings.Join(jsonStrings, ",") + "]"

	err = jsonToCSV(strings.NewReader(jsonStringArray), writer)
	if err != nil {
		return fmt.Errorf("%w %s", dynamicforms.ErrInternal, err)
	}

	return nil
}

// https://github.com/yukithm/json2csv
// modified to make column names more readable
func jsonToCSV(r io.Reader, w io.Writer) error {
	dec := json.NewDecoder(r)

	var data interface{}

	if err := dec.Decode(&data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	var rows []map[string]string

	switch value := data.(type) {
	case []interface{}:
		for i := range value {
			rows = append(rows, topLevelObject(value[i]))
		}
	default:
		rows = append(rows, topLevelObject(value))
	}

	columns := make(map[string]string)
	for i := range rows {
		for col := range rows[i] {
			columns[col] = col
		}
	}
	var colRecord []string
	for c := range columns {
		colRecord = append(colRecord, c)
	}
	sort.Strings(colRecord)

	cw := csv.NewWriter(w)

	if err := cw.Write(colRecord); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for i := range rows {
		record := make([]string, 0, len(columns))
		for _, col := range colRecord {
			record = append(record, rows[i][col])
		}
		if err := cw.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}

	}

	cw.Flush()

	return nil
}

func topLevelObject(object interface{}) map[string]string {
	values := make(map[string]string)

	switch value := object.(type) {
	case string:
		values["text"] = value
	case map[string]interface{}:
		flattenObject("", values, value)
	case float64:
		values["number"] = strconv.FormatFloat(value, 'f', -1, 64)
	case []interface{}:
		addValue("", values, value)
	}

	return values
}

func flattenObject(path string, values map[string]string, obj map[string]interface{}) {
	for k, v := range obj {
		p := k
		if path != "" {
			p = path + "." + k
		}
		addValue(p, values, v)
	}
}

func addValue(path string, values map[string]string, v interface{}) {
	switch value := v.(type) {
	case string:
		path = slugToSentence(path)
		values[path] = value
	case map[string]interface{}:
		flattenObject(path, values, value)
	case float64:
		path = slugToSentence(path)
		values[path] = strconv.FormatFloat(value, 'f', -1, 64)
	case []interface{}:
		for i := range value {
			p := strconv.Itoa(i)
			if path != "" {
				p = path + "." + p
			}
			addValue(p, values, value[i])
		}
	}
}

func slugToSentence(slug string) string {
	sentence := strings.ReplaceAll(slug, "-", " ")
	runes := []rune(sentence)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
