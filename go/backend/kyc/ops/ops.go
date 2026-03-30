package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"

	"github.com/bxcodec/faker/v3"

	"gitlab.com/fynbos/backend/kyc"
)

func mergeIdentities(old dbIndividualDetails, new kyc.IndividualDetails) (dbIndividualDetails, bool, error) {
	var merged dbIndividualDetails
	noop := true
	merged.WalletID = new.WalletID
	merged.IPAddress = new.IPAddress

	merged.FirstName = old.FirstName
	if new.FirstName != "" && new.FirstName != old.FirstName {
		noop = false
		merged.FirstName = new.FirstName
	}

	merged.LastName = old.LastName
	if new.LastName != "" && new.LastName != old.LastName {
		noop = false
		merged.LastName = new.LastName
	}

	merged.CountryCode = old.CountryCode
	if new.CountryCode != "" && new.CountryCode != old.CountryCode {
		noop = false
		merged.CountryCode = new.CountryCode
	}

	merged.Gender = old.Gender
	if new.Gender != kyc.GenderUnknown && new.Gender != old.Gender {
		noop = false
		merged.Gender = new.Gender
	}

	merged.DateOfBirth = old.DateOfBirth
	if !new.DateOfBirth.IsZero() && !new.DateOfBirth.Equal(old.DateOfBirth.Time) {
		noop = false
		merged.DateOfBirth = sql.NullTime{
			Time:  new.DateOfBirth,
			Valid: true,
		}
	}

	merged.Nationality = old.Nationality
	if new.Nationality != "" {
		merged.Nationality = sql.NullString{
			String: new.Nationality,
			Valid:  true,
		}
	}

	merged.PlaceOfBirth = old.PlaceOfBirth
	if new.PlaceOfBirth != "" {
		merged.PlaceOfBirth = sql.NullString{
			String: new.PlaceOfBirth,
			Valid:  true,
		}
	}

	merged.Address = old.Address
	if new.Address != nil {
		noop = false
		addressJson, err := json.Marshal(new.Address)
		if err != nil {
			return merged, noop, err
		}
		merged.Address = sql.NullString{
			String: string(addressJson),
			Valid:  true,
		}
	}

	merged.Revision = old.Revision + 1

	return merged, noop, nil
}

type dbIndividualDetails struct {
	kyc.IndividualDetails
	Revision     int            `db:"revision"`
	DateOfBirth  sql.NullTime   `db:"date_of_birth"`
	Address      sql.NullString `db:"address"`
	PlaceOfBirth sql.NullString `db:"place_of_birth"`
	Nationality  sql.NullString `db:"nationality"`
}

func getIndividualDetails(ctx context.Context, b Backends, walletID string) (*dbIndividualDetails, error) {
	var id dbIndividualDetails
	err := b.DB().GetContext(ctx, &id,
		"SELECT wallet_id, revision, country_code, first_name, last_name, gender, date_of_birth, address, ip_address, place_of_birth, nationality FROM individual_kyc_details WHERE wallet_id=$1 ORDER BY revision DESC LIMIT 1",
		walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, kyc.ErrNoKYCInfo
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	return &id, nil
}

func GetIndividualDetails(ctx context.Context, b Backends, walletID string) (*kyc.IndividualDetails, error) {
	db, err := getIndividualDetails(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	return convertDBDetails(*db)
}

func UpdateIndividualDetails(ctx context.Context, b Backends, ident kyc.IndividualDetails) (*kyc.IndividualDetails, error) {
	err := b.Validator().Struct(ident)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	old, err := getIndividualDetails(ctx, b, ident.WalletID)
	if err != nil && !errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, err
	}

	// If this is the first insert we can compare with blank identity
	if old == nil {
		old = &dbIndividualDetails{}
	}
	merged, noop, err := mergeIdentities(*old, ident)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	if noop {
		//  Do nothing, just lookup the existing and off we go
		return convertDBDetails(merged)
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO individual_kyc_details (revision, wallet_id, country_code, first_name, last_name, gender, date_of_birth, address, ip_address, place_of_birth, nationality)"+
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		merged.Revision, merged.WalletID, merged.CountryCode, merged.FirstName, merged.LastName, merged.Gender, merged.DateOfBirth, merged.Address, merged.IPAddress, merged.PlaceOfBirth, merged.Nationality)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	return convertDBDetails(merged)
}

func convertDBDetails(details dbIndividualDetails) (*kyc.IndividualDetails, error) {
	resp := &kyc.IndividualDetails{
		WalletID:     details.WalletID,
		FirstName:    details.FirstName,
		LastName:     details.LastName,
		CountryCode:  details.CountryCode,
		PlaceOfBirth: details.PlaceOfBirth.String,
		Nationality:  details.PlaceOfBirth.String,
		Gender:       details.Gender,
		DateOfBirth:  details.DateOfBirth.Time,
		IPAddress:    details.IPAddress,
	}

	if !details.Address.Valid {
		return resp, nil
	}

	var address kyc.Address
	err := json.Unmarshal([]byte(details.Address.String), &address)
	if err != nil {
		return nil, fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	resp.Address = &address

	return resp, nil
}

func SetKYCStatus(ctx context.Context, b Backends, args kyc.StatusUpdateArgs) error {
	walletID := args.WalletID
	status := args.Status

	wo := client.StartWorkflowOptions{
		ID:                       "kyc_set_status_" + walletID + "_" + status.String(),
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: 2 * time.Minute,
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		return fmt.Errorf("%w %s", kyc.ErrInternal, err)
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// return workflow if it's running
	var await client.WorkflowRun
	var executeErr error
	if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		await = b.Temporal().GetWorkflow(ctx, wo.ID, "")
	} else {
		await, executeErr = b.Temporal().ExecuteWorkflow(ctx, wo, SetKYCStatusWorkflow, SetKYCStatusWorkflowArgs{
			WalletID:  walletID,
			Status:    status,
			Reason:    args.Reason,
			EventType: args.EventType,
		})
	}
	if executeErr != nil {
		return fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	return await.Get(ctx, nil)
}

func GetKYCStatus(ctx context.Context, b Backends, walletID string) (kyc.Status, error) {
	var s kyc.Status
	err := b.DB().GetContext(ctx, &s, "select status from wallet_kyc_status where wallet_id = $1", walletID)
	if err != nil {
		// Return unknown if it doesn't exist yet
		if errors.Is(err, sql.ErrNoRows) {
			return s, nil
		}
		return s, err
	}

	return s, nil
}

func GetKYCStatusMetadata(ctx context.Context, b Backends, walletID string) (*kyc.StatusMetadata, error) {
	var row struct {
		Status                 kyc.Status     `db:"status"`
		StatusReason           sql.NullString `db:"status_reason"`
		LastWebhookEventType   sql.NullString `db:"last_webhook_event_type"`
		ResubmissionCount      int32          `db:"resubmission_count"`
		DocumentExpirationDate sql.NullTime   `db:"document_expiration_date"`
	}

	err := b.DB().GetContext(ctx, &row, "select status, status_reason, last_webhook_event_type, resubmission_count, document_expiration_date from wallet_kyc_status where wallet_id = $1", walletID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &kyc.StatusMetadata{Status: kyc.StatusUnknown}, nil
		}

		return nil, err
	}

	meta := &kyc.StatusMetadata{
		Status:            row.Status,
		ResubmissionCount: row.ResubmissionCount,
	}

	if row.StatusReason.Valid {
		meta.Reason = row.StatusReason.String
	}

	if row.LastWebhookEventType.Valid {
		meta.LastWebhookEvent = row.LastWebhookEventType.String
	}

	if row.DocumentExpirationDate.Valid {
		exp := row.DocumentExpirationDate.Time
		meta.ExpirationDate = &exp
	}

	return meta, nil
}

func CanResubmitKYC(ctx context.Context, b Backends, walletID string) (bool, error) {
	status, err := GetKYCStatus(ctx, b, walletID)
	if err != nil {
		return false, err
	}

	switch status {
	case kyc.StatusUnknown, kyc.StatusDocumentsRequired, kyc.StatusDenied:
		return true, nil
	default:
		return false, nil
	}
}

func GenerateKycData(ctx context.Context, b Backends, walletID string) error {
	_, err := UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:     walletID,
		FirstName:    faker.FirstName(),
		LastName:     faker.LastName(),
		CountryCode:  "ZA",
		PlaceOfBirth: "Cape Town",
		Nationality:  "ZA",
		Gender:       kyc.GenderMale,
		DateOfBirth:  time.Date(1991, time.August, 21, 0, 0, 0, 0, time.UTC),
		Address: &kyc.Address{
			Line1:       "31 Royal Ave",
			Line2:       "",
			Building:    "",
			Apartment:   "",
			City:        "Cape Town",
			ZipCode:     "9301",
			CountryCode: "ZA",
		},
		IPAddress: "8.8.8.8",
	})
	if err != nil {
		return err
	}

	err = SetKYCStatus(ctx, b, kyc.StatusUpdateArgs{WalletID: walletID, Status: kyc.StatusLevel1})

	return err
}

func IsKYCApproved(ctx context.Context, b Backends, walletID string) (bool, error) {
	status, err := GetKYCStatus(ctx, b, walletID)
	if err != nil {
		return false, err
	}
	return status == kyc.StatusApproved || status == kyc.StatusLevel1 || status == kyc.StatusLevel2, nil
}
