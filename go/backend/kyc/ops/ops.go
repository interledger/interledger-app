package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/email"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/workflows"
	"go.temporal.io/api/enums/v1"
	temporal "go.temporal.io/sdk/client"
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

func SetKYCStatus(ctx context.Context, b Backends, walletID string, status kyc.Status) error {
	old, err := GetKYCStatus(ctx, b, walletID)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx,
		"INSERT INTO wallet_kyc_status (wallet_id, status) VALUES ($1, $2) ON CONFLICT (wallet_id) DO UPDATE SET status = excluded.status;",
		walletID, status)
	if err != nil {
		return fmt.Errorf("%w %s", kyc.ErrInternal, err)
	}

	err = b.Notify().NotifyWallet(ctx, walletID, notify.NotificationTypeKyc)
	if err != nil {
		log.Error("notify error", zap.Error(err), zap.String("type", "kyc"))
	}

	// only send out email if moving into the denied state.
	kycData, err := GetIndividualDetails(ctx, b, walletID)
	if err != nil {
		kycData = &kyc.IndividualDetails{}
	}

	greeting := fmt.Sprintf("Hello %s", kycData.FirstName)
	greeting = strings.TrimSpace(greeting) + ","
	if old != kyc.StatusDenied && status == kyc.StatusDenied {
		err = b.Email().SendMailTemplate(ctx, walletID, email.ApplicationDenied, map[string]interface{}{
			"subject": email.ApplicationDenied.Subject(),
			"data": []map[string]interface{}{
				{"paragraph": greeting},
				{"heading": "Your wallet verification was denied"},
				{"paragraph": "We are unable to verify your identity at this time. Please contact support using the details below."},
			},
		}, nil)

		// we don't send out approved email if user moves from kyc level 1 to kyc level 2
	} else if old != kyc.StatusLevel1 && old != kyc.StatusLevel2 && (status == kyc.StatusLevel1 || status == kyc.StatusLevel2) {
		var w *wallets.Wallet
		w, err = b.Wallets().Get(ctx, walletID)
		if err != nil {
			w = &wallets.Wallet{}
		}

		var address wallets.Address
		if len(w.Addresses) > 1 {
			address = w.Addresses[0]
		}

		err = b.Email().SendMailTemplate(ctx, walletID, email.ApplicationApproved, map[string]interface{}{
			"subject": email.ApplicationApproved.Subject(),
			"data": []map[string]interface{}{
				{"paragraph": greeting},
				{"heading": "Your wallet is active"},
				{"code": address.String()},
				{"paragraph": "You can now use your wallet to send and receive payments."},
			},
		}, nil)
	} else if old != kyc.StatusPending && status == kyc.StatusPending {
		err = b.Email().SendMailTemplate(ctx, walletID, email.ApplicationPending, map[string]interface{}{
			"subject": email.ApplicationPending.Subject(),
			"data": []map[string]interface{}{
				{"paragraph": greeting},
				{"heading": "Pending review"},
				{"paragraph": "Your account creation is currently under review. We will notify you once it's complete or if any further information is needed."},
			},
		}, nil)
	}

	if err != nil {
		log.Error("Failed to send kyc application denied email.", zap.Error(err), zap.String("walletID", walletID), zap.String("previous status", old.String()))
	}

	return nil
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

func StartKYC(ctx context.Context, b Backends, walletID string) error {

	workflowOptions := temporal.StartWorkflowOptions{
		ID:                       "start_kyc_" + walletID,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: time.Hour * 24 * 8, // Workflow has 8 days to complete
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	_, err := b.Temporal().ExecuteWorkflow(ctx, workflowOptions, workflows.StartKYC, workflows.StartKYCArgs{
		WalletID: walletID,
	})

	return err
}
