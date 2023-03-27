package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gmt"
	"gitlab.com/fynbos/backend/providers/gmt/external"
	"gitlab.com/fynbos/backend/providers/mx"
	"go.temporal.io/sdk/temporal"
)

const gmtEventsChannel = "gmt_events"

type Credentials struct {
	Alias    string
	User     string
	Password string
	MTS      int32
}

type Activity struct {
	b     Backends
	ext   external.Client
	creds Credentials
}

func NewActivity(b Backends) *Activity {
	mtsStr := getEnvDefault("GMT_MTS_ID", "1")
	mts, err := strconv.Atoi(mtsStr)
	if err != nil {
		return nil
	}
	a := &Activity{
		b:   b,
		ext: external.NewClient(),
		creds: Credentials{
			Alias:    getEnvDefault("GMT_ALIAS", "FYN001"),
			User:     getEnvDefault("GMT_USER", "Fynbos_api"),
			Password: getEnvDefault("GMT_PASSWORD", "VUJ6bnkxN2dQVXkwMjZaOA=="),
			MTS:      int32(mts),
		},
	}

	return a
}

func getEnvDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	return val
}

func (a *Activity) CheckWalletOFAC(ctx context.Context, walletID string) error {
	id, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return err
	}

	res, err := a.ext.OfacVerification(ctx, external.OfacVerification{
		Alias:     a.creds.Alias,
		User:      a.creds.User,
		Pass:      a.creds.Password,
		LastName:  id.LastName,
		FirstName: id.FirstName,
	})
	if err != nil {
		return err
	}

	if res.Error != 0 {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s)", res.Error, res.Message), "external", nil)
	}

	return nil
}

type ComplianceResp struct {
	SenderID         int64
	SenderWalletID   string
	ReceiverID       int64
	ReceiverWalletID string
}

func (a *Activity) IndividualCompliance(ctx context.Context, walletID string) (*ComplianceResp, error) {
	id, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sender, err := senderFromWallet(ctx, a.b, walletID)
	if err != nil {
		return nil, err
	}

	recv, err := receiverFromWallet(ctx, a.b, walletID)
	if err != nil {
		return nil, err
	}

	las, err := a.b.LinkedAccounts().ListByWalletId(ctx, walletID)
	if err != nil {
		return nil, err
	}

	var la linkedaccounts.LinkedAccount
	var found bool
	for _, l := range las {
		if l.Provider == mx.ProviderName && l.Type == mx.TypeBankAccount {
			la = l
			found = true
			break
		}
	}

	if !found {
		return nil, temporal.NewNonRetryableApplicationError("no linked account found for wallet", "NotFound", linkedaccounts.ErrNotFound)
	}

	acc, err := a.b.MX().GetAccount(ctx, walletID, la.ProviderID)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sender.SenderAchAccount = acc.AccountNumber
	sender.SenderAchRouting = acc.RoutingNumber
	if acc.Type == mx.TypeSavings {
		sender.SenderAchType = "SV"
	} else if acc.Type == mx.TypeChecking {
		sender.SenderAchType = "CHK"
	}

	res, err := a.ext.ComplianceCheck(ctx, external.ComplianceCheck{
		Alias:    a.creds.Alias,
		User:     a.creds.User,
		Pass:     a.creds.Password,
		Sender:   sender,
		Receiver: recv,
		Transfer: &external.WsTransferInfo{
			AgenciaCodigo:          "",
			AgencySpecialDiscounts: "",
			AmountToReceive:        1.0,
			CorrespondentCode:      "GACH",
			DestinationCurrency:    "USD",
			ExchangeRate:           1,
			Fee:                    0,
			MTSID:                  a.creds.MTS,
			NetAmount:              1.0,
			OriginalCurrency:       "USD",
			OriginalPaymentMethod:  "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:           id.Address.City,
			ReceiverState:          id.Address.State,
			SaveSenderReceiver:     true,
			SenderID:               sender.SenderId,
			SucursalBanco:          acc.RoutingNumber, // Bank branch or routing number
			BankAccount:            acc.AccountNumber,
			BankCode:               acc.InstitutionCode,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.Error != 0 {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s)", res.Error, res.Message), "external", nil)
	}

	return &ComplianceResp{
		SenderID:         int64(res.SenderID),
		SenderWalletID:   walletID,
		ReceiverID:       int64(res.ReceiverID),
		ReceiverWalletID: walletID,
	}, nil
}

func (a *Activity) ACHCompliance(ctx context.Context, args gmt.TransfersArgs) (*ComplianceResp, error) {
	fromLA, err := a.b.LinkedAccounts().Get(ctx, args.FromLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	toLA, err := a.b.LinkedAccounts().Get(ctx, args.ToLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	toID, err := a.b.KYC().GetIndividualDetails(ctx, toLA.WalletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sender, err := senderFromWallet(ctx, a.b, fromLA.WalletID)
	if err != nil {
		return nil, err
	}

	fromAcc, err := a.b.MX().GetAccount(ctx, args.FromWalletID, fromLA.ProviderID)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sender.SenderAchAccount = fromAcc.AccountNumber
	sender.SenderAchRouting = fromAcc.RoutingNumber
	if fromAcc.Type == mx.TypeSavings {
		sender.SenderAchType = "SV"
	} else if fromAcc.Type == mx.TypeChecking {
		sender.SenderAchType = "CHK"
	}

	receiver, err := receiverFromWallet(ctx, a.b, toLA.WalletID)
	if err != nil {
		return nil, err
	}

	toAcc, err := a.b.MX().GetAccount(ctx, args.ToWalletID, toLA.ProviderID)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	res, err := a.ext.ComplianceCheck(ctx, external.ComplianceCheck{
		Alias:    a.creds.Alias,
		User:     a.creds.User,
		Pass:     a.creds.Password,
		Sender:   sender,
		Receiver: receiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       args.Amount.Float64(),
			CorrespondentCode:     "GACH",
			DestinationCurrency:   args.Amount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.creds.MTS,
			NetAmount:             args.Amount.Float64(),
			OriginalCurrency:      args.Amount.Currency.String(),
			OriginalPaymentMethod: "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:          toID.Address.City,
			ReceiverState:         toID.Address.State,
			SenderID:              sender.SenderId,
			SucursalBanco:         toAcc.RoutingNumber, // Bank branch or routing number
			BankAccount:           toAcc.AccountNumber,
			BankCode:              toAcc.InstitutionCode,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.Error != 0 {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s)", res.Error, res.Message), "external", nil)
	}

	return &ComplianceResp{
		SenderID:         int64(res.SenderID),
		SenderWalletID:   args.FromWalletID,
		ReceiverID:       int64(res.ReceiverID),
		ReceiverWalletID: args.ToWalletID,
	}, nil
}

func receiverFromWallet(ctx context.Context, b Backends, walletID string) (*external.WsReceiver, error) {
	recvID, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	recvUsers, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return nil, err
	}

	sid, err := getSenderID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	rid, err := getReceiverID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	gender := "Male"
	if recvID.Gender == kyc.GenderFemale {
		gender = "Female"
	}

	return &external.WsReceiver{
		ReceiverAddress:   recvID.Address.FormattedAddress,
		ReceiverBirthDate: external.GMTDate(recvID.DateOfBirth),
		ReceiverCity:      recvID.Address.City,
		ReceiverCountry:   recvID.Address.CountryCode,
		ReceiverCurrency:  "USD",
		ReceiverEmail:     recvUsers[0].Email,
		ReceiverGender:    gender,
		ReceiverId:        int32(rid),
		ReceiverLastName:  recvID.LastName,
		ReceiverMobile:    recvUsers[0].PhoneNumber,
		ReceiverName:      recvID.FirstName,
		ReceiverState:     recvID.Address.State,
		ReceiverZip:       recvID.Address.ZipCode,
		SenderID:          int32(sid),
	}, nil
}

func senderFromWallet(ctx context.Context, b Backends, walletID string) (*external.WsSender, error) {
	senderID, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	senderUsers, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return nil, err
	}

	sid, err := getSenderID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	gender := "Male"
	if senderID.Gender == kyc.GenderFemale {
		gender = "Female"
	}
	return &external.WsSender{
		SenderAddress:               senderID.Address.FormattedAddress,
		SenderAddressStreet:         senderID.Address.Apartment,
		SenderBirthDate:             external.GMTDate(senderID.DateOfBirth),
		SenderCity:                  senderID.Address.City,
		SenderCountryCode:           senderID.Address.CountryCode,
		SenderCurrencyCode:          "USD",
		SenderEmail:                 senderUsers[0].Email,
		SenderForceNew:              false,
		SenderGender:                gender,
		SenderIP:                    senderID.IPAddress,
		SenderId:                    int32(sid),
		SenderIsBusiness:            false,
		SenderLastName:              senderID.LastName,
		SenderMobile:                senderUsers[0].PhoneNumber,
		SenderName:                  senderID.FirstName,
		SenderResidenceAddress:      senderID.Address.String(),
		SenderResidenceAddressExtra: senderID.Address.Apartment,
		SenderResidenceCity:         senderID.Address.City,
		SenderResidenceCountryCode:  senderID.Address.CountryCode,
		SenderResidenceState:        senderID.Address.State,
		SenderResidenceZip:          senderID.Address.ZipCode,
		SenderState:                 senderID.Address.State,
		SenderZip:                   senderID.Address.ZipCode,
	}, nil
}

func getSenderID(ctx context.Context, b Backends, walletID string) (int64, error) {
	var id sql.NullInt64
	err := b.DB().GetContext(ctx, &id, "SELECT sender_id from gmt_users WHERE wallet_id=$1", walletID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return id.Int64, nil
}

func getReceiverID(ctx context.Context, b Backends, walletID string) (int64, error) {
	var id sql.NullInt64
	err := b.DB().GetContext(ctx, &id, "SELECT receiver_id from gmt_users WHERE wallet_id=$1", walletID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return id.Int64, nil
}

func (a *Activity) UpdateSendRecvUser(ctx context.Context, cr ComplianceResp) error {
	if cr.SenderID != 0 {
		_, err := a.b.DB().ExecContext(ctx, "INSERT INTO gmt_users (wallet_id, sender_id) VALUES ($1, $2)  ON CONFLICT (wallet_id) DO UPDATE SET sender_id = excluded.sender_id, updated_at=now()",
			cr.SenderWalletID, cr.SenderID)
		if err != nil {
			return err
		}
	}

	if cr.ReceiverID != 0 {
		_, err := a.b.DB().ExecContext(ctx, "INSERT INTO gmt_users (wallet_id, receiver_id) VALUES ($1, $2)  ON CONFLICT (wallet_id) DO UPDATE SET receiver_id = excluded.receiver_id, updated_at=now()",
			cr.ReceiverWalletID, cr.ReceiverID)
		if err != nil {
			return err
		}
	}

	return nil
}

type TransactionResp struct {
	ID         string
	ReceiptRef string // GMT interal reference
	Status     string
	//Receipt information
	Licence  string
	RTR      string // Right to refund
	ErrorMsg string
	Contact  string
}

func (a *Activity) InsertACH(ctx context.Context, args gmt.TransfersArgs) (*TransactionResp, error) {
	fromLA, err := a.b.LinkedAccounts().Get(ctx, args.FromLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	toLA, err := a.b.LinkedAccounts().Get(ctx, args.ToLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	toID, err := a.b.KYC().GetIndividualDetails(ctx, toLA.WalletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sender, err := senderFromWallet(ctx, a.b, fromLA.WalletID)
	if err != nil {
		return nil, err
	}

	fromAcc, err := a.b.MX().GetAccount(ctx, args.FromWalletID, fromLA.ProviderID)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sender.SenderAchAccount = fromAcc.AccountNumber
	sender.SenderAchRouting = fromAcc.RoutingNumber
	if fromAcc.Type == mx.TypeSavings {
		sender.SenderAchType = "SV"
	} else if fromAcc.Type == mx.TypeChecking {
		sender.SenderAchType = "CHK"
	}

	receiver, err := receiverFromWallet(ctx, a.b, toLA.WalletID)
	if err != nil {
		return nil, err
	}

	toAcc, err := a.b.MX().GetAccount(ctx, args.ToWalletID, toLA.ProviderID)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	res, err := a.ext.InsertTransaction(ctx, external.InsertTransaction{
		Alias:    a.creds.Alias,
		User:     a.creds.User,
		Pass:     a.creds.Password,
		Sender:   sender,
		Receiver: receiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       args.Amount.Float64(),
			CorrespondentCode:     "GACH",
			DestinationCurrency:   args.Amount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.creds.MTS,
			NetAmount:             args.Amount.Float64(),
			OriginalCurrency:      args.Amount.Currency.String(),
			OriginalPaymentMethod: "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:          toID.Address.City,
			ReceiverState:         toID.Address.State,
			SaveSenderReceiver:    true,
			SenderID:              sender.SenderId,
			SucursalBanco:         toAcc.RoutingNumber, // Bank branch or routing number
			BankAccount:           toAcc.AccountNumber,
			BankCode:              toAcc.InstitutionCode,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.Error != 0 {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s)", res.Error, res.Message), "external", nil)
	}

	return &TransactionResp{
		ID:         res.Password,
		ReceiptRef: res.Receipt,
		Status:     res.Status,
		Licence:    res.Receipt_License,
		RTR:        res.Receipt_RTR_EN,
		ErrorMsg:   res.Receipt_Error_EN,
		Contact:    res.Status,
	}, nil
}

func (a *Activity) SaveReceipt(ctx context.Context, tr TransactionResp) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO gmt_receipts (external_id, receipt, licence, right_to_return, error_msg, contact) VALUES ($1, $2, $3, $4, $5, $6)",
		tr.ID, tr.ReceiptRef, tr.Licence, tr.RTR, tr.ErrorMsg, tr.Contact)
	return err
}

func (a *Activity) VerifyTransaction(ctx context.Context, id string) error {
	res, err := a.ext.SetVerified(ctx, external.SetVerified{
		Alias:   a.creds.Alias,
		User:    a.creds.User,
		Pass:    a.creds.Password,
		Receipt: id,
		Passed:  true,
	})
	if err != nil {
		return err
	}

	if !res.Valid {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s)", res.Error, res.Message), "external", nil)
	}

	return nil
}

type CreateWorkflowRefArgs struct {
	ExternalID    string
	WorkflowID    string
	WorkflowRunID string
	ActivityName  string
}

func (a *Activity) CreateWorkflowRef(ctx context.Context, args CreateWorkflowRefArgs) (string, error) {
	id := uuid.NewString()
	insert := db.NewInsert("gmt_workflow_refs").
		Value("id", id).
		Value("external_id", args.ExternalID).
		Value("workflow_id", args.WorkflowID).
		Value("workflow_run_id", args.WorkflowRunID).
		Value("activity_name", args.ActivityName).
		Value("completed", false)

	statement, values, err := insert.GetStatement()
	if err != nil {
		return "", err
	}

	_, err = a.b.DB().ExecContext(
		ctx,
		statement,
		values...,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (a *Activity) CompleteWorkflowRef(ctx context.Context, refID string) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE gmt_workflow_refs SET completed=true, updated_at=now() WHERE id=$1", refID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) ConfirmNotification(ctx context.Context, externalID string) error {
	resp, err := a.ext.ConfirmCollection(ctx, external.ConfirmCollection{
		Alias:   a.creds.Alias,
		User:    a.creds.User,
		Pass:    a.creds.Password,
		Receipt: externalID,
	})

	if err != nil {
		return err
	}

	if resp.Error != 0 {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s)", resp.Error, resp.Message), "external", nil)
	}

	return nil
}
