package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"gitlab.com/fynbos/backend/country"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/backend/transactions"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers"
	"gitlab.com/fynbos/backend/providers/gmt/external"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/mx"
	"go.temporal.io/sdk/temporal"
)

const (
	gmtEventsChannel = "gmt_events"
	gmt3DSChannel    = "gmt_3ds"
)

const (
	card2AchAuthenticate3ds = "CARD_2_ACH_AUTHENTICATE_3DS"
)

type Activity struct {
	b   Backends
	ext external.Client
	mts int32
}

func NewActivity(b Backends) *Activity {
	mtsStr := getEnvDefault("GMT_MTS_ID", "1")
	mts, err := strconv.Atoi(mtsStr)
	if err != nil {
		return nil
	}
	a := &Activity{
		b: b,
		ext: external.NewClient(
			otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, external.Redact),
			),
		),
		mts: int32(mts),
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

// IndividualCompliance does a compliance check by doing a $1 payment where the user is both the sender and receiver.
func (a *Activity) IndividualCompliance(ctx context.Context, args providers.TransfersArgs, walletID string) (*ComplianceResp, error) {
	id, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sender, err := senderFromWallet(ctx, a.b, args, walletID)
	if err != nil {
		return nil, err
	}

	recv, err := receiverFromWallet(ctx, a.b, walletID)
	if err != nil {
		return nil, err
	}

	res, err := a.ext.ComplianceCheck(ctx, external.ComplianceCheck{
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
			MTSID:                  a.mts,
			NetAmount:              1.0,
			OriginalCurrency:       "USD",
			OriginalPaymentMethod:  "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:           id.Address.City,
			ReceiverState:          id.Address.State,
			SaveSenderReceiver:     true,
			SenderID:               sender.SenderId,
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

func (a *Activity) Card2ACHCompliance(ctx context.Context, args providers.TransfersArgs) (*ComplianceResp, error) {
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

	sender, err := senderFromWallet(ctx, a.b, args, fromLA.WalletID)
	if err != nil {
		return nil, err
	}
	sender.SenderTrackingNumber = args.FromTransactionID

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
		Sender:   sender,
		Receiver: receiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       args.Amount.Float64(),
			CorrespondentCode:     "GACH",
			DestinationCurrency:   args.Amount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.mts,
			NetAmount:             args.Amount.Float64(),
			OriginalCurrency:      args.Amount.Currency.String(),
			OriginalPaymentMethod: "DEBIT", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:          toID.Address.City,
			ReceiverState:         toID.Address.State,
			SenderID:              sender.SenderId,
			SucursalBanco:         toAcc.RoutingNumber, // Bank branch or routing number
			BankAccount:           toAcc.AccountNumber,
			BankCode:              toAcc.InstitutionCode,
			ServicioCodigo:        "BD",
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

func (a *Activity) ACH2CardCompliance(ctx context.Context, args providers.TransfersArgs) (*ComplianceResp, error) {
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

	fromAcc, err := a.b.MX().GetAccount(ctx, args.FromWalletID, fromLA.ProviderID)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sender, err := senderFromWallet(ctx, a.b, args, fromLA.WalletID)
	if err != nil {
		return nil, err
	}
	sender.SenderTrackingNumber = args.FromTransactionID
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

	res, err := a.ext.ComplianceCheck(ctx, external.ComplianceCheck{
		Sender:   sender,
		Receiver: receiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       args.Amount.Float64(),
			CorrespondentCode:     "USCD",
			DestinationCurrency:   args.Amount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.mts,
			NetAmount:             args.Amount.Float64(),
			OriginalCurrency:      args.Amount.Currency.String(),
			OriginalPaymentMethod: "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:          toID.Address.City,
			ReceiverState:         toID.Address.State,
			SenderID:              sender.SenderId,
			TipoCuentaCodigo:      "CDN",
			ServicioCodigo:        "BD",
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

func (a *Activity) ACHCompliance(ctx context.Context, args providers.TransfersArgs) (*ComplianceResp, error) {
	fromLA, err := a.b.LinkedAccounts().Get(ctx, args.FromLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("failed to load from linked account: "+err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to load from linked account")
	}

	toLA, err := a.b.LinkedAccounts().Get(ctx, args.ToLinkedAccountID)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("failed to load to linked account: "+err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to load to linked account")
	}

	toID, err := a.b.KYC().GetIndividualDetails(ctx, toLA.WalletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError("failed to load to kyc info: "+err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to load to kyc info")
	}

	sender, err := senderFromWallet(ctx, a.b, args, fromLA.WalletID)
	if err != nil {
		return nil, err
	}

	fromAcc, err := a.b.MX().GetAccount(ctx, args.FromWalletID, fromLA.ProviderID)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("failed to load from mx account: "+err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to load from mx account")
	}

	sender.SenderTrackingNumber = args.FromTransactionID
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
		return nil, temporal.NewNonRetryableApplicationError("failed to load to mx account: "+err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to load to mx account")
	}

	res, err := a.ext.ComplianceCheck(ctx, external.ComplianceCheck{
		Sender:   sender,
		Receiver: receiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       args.Amount.Float64(),
			CorrespondentCode:     "GACH",
			DestinationCurrency:   args.Amount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.mts,
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
		return nil, temporal.NewNonRetryableApplicationError("failed to load recv wallet KYC: "+err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to load recv wallet KYC")
	}

	recvUsers, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to get receiver user from wallet ID")
	}

	sid, err := getSenderID(ctx, b, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to get receiver sender ID")
	}

	rid, err := getReceiverID(ctx, b, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to get receiver ID")
	}

	gender := "Male"
	if recvID.Gender == kyc.GenderFemale {
		gender = "Female"
	}

	address := recvID.Address.FormattedAddress
	if address == "" {
		address = recvID.Address.String()
	}

	return &external.WsReceiver{
		ReceiverAddress:   address,
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

func senderFromWallet(ctx context.Context, b Backends, args providers.TransfersArgs, walletID string) (*external.WsSender, error) {
	senderID, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError("failed to load to wallet KYC: "+err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to load to wallet KYC")
	}

	senderUsers, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w : %s, walletID (%s)", err, "failed to listusers for wallet", walletID)
	}

	sid, err := getSenderID(ctx, b, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to get sender ID for wallet")
	}

	ipAddress := senderID.IPAddress
	if args.IPAddress != "" {
		ipAddress = args.IPAddress
	}

	gender := "Male"
	if senderID.Gender == kyc.GenderFemale {
		gender = "Female"
	}

	exceeds, err := b.Limits().ExceedsGMTLimits(ctx, walletID, args.Amount)
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to get whether sender exceeds limits")
	}

	address := senderID.Address.FormattedAddress
	if address == "" {
		address = senderID.Address.String()
	}

	sender := &external.WsSender{
		SenderAddress:               address,
		SenderAddressStreet:         senderID.Address.Apartment,
		SenderBirthDate:             external.GMTDate(senderID.DateOfBirth),
		SenderCity:                  senderID.Address.City,
		SenderCountryCode:           senderID.Address.CountryCode,
		SenderCurrencyCode:          "USD",
		SenderEmail:                 senderUsers[0].Email,
		SenderForceNew:              false,
		SenderGender:                gender,
		SenderIP:                    ipAddress,
		SenderId:                    int32(sid),
		SenderIsBusiness:            false,
		SenderLastName:              senderID.LastName,
		SenderMobile:                senderUsers[0].PhoneNumber,
		SenderName:                  senderID.FirstName,
		SenderResidenceAddress:      address,
		SenderResidenceAddressExtra: senderID.Address.Apartment,
		SenderResidenceCity:         senderID.Address.City,
		SenderResidenceCountryCode:  senderID.Address.CountryCode,
		SenderResidenceState:        senderID.Address.State,
		SenderResidenceZip:          senderID.Address.ZipCode,
		SenderState:                 senderID.Address.State,
		SenderZip:                   senderID.Address.ZipCode,
	}

	if (!exceeds && !args.ForceEDD) || args.ForceNoEDD {
		return sender, nil
	}

	log.Info("Setting user EDD on GMT sender", zap.String("wallet_id", args.FromWalletID), zap.Bool("force_edd", args.ForceEDD), zap.Bool("exceeds", exceeds))

	idNums, err := b.KYC().GetPersonaIDNumbers(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w : %s", err, "failed to load sender Persona ID numbers")
	}

	sender.SenderIdNumber2 = idNums.SocialSecurity
	sender.SenderIdNumber = idNums.IdentificationNumber
	if !idNums.ExpirationDate.IsZero() {
		sender.SenderDocExpiration = external.GMTDate(idNums.ExpirationDate)
	}

	sender.SenderIdIssuer = idNums.IssuingCountry
	state, ok := country.States[country.US][idNums.IssuingState]
	if ok {
		sender.SenderIdIssuer = state
	}

	switch idNums.IdentificationClass {
	case "dl":
		sender.SenderIdType = "DRIVERS LICENSE"
	case "pp", "ppc":
		sender.SenderIdType = "PASSPORT"
	case "rp":
		sender.SenderIdType = "RESIDENT ALIEN CARD"
	case "id":
		if idNums.IssuingState != "" && idNums.IssuingState != "US" {
			sender.SenderIdType = "STATE-ISSUED ID"
		} else {
			sender.SenderIdType = "NATIONAL ID CARD"
		}
	default:
		log.Error("Unknown Persona ID number type", zap.String("persona_id_type", idNums.IdentificationClass))
	}

	sender.SenderOccupation = "Contracting Engineer"

	return sender, nil
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

func (a *Activity) InsertACH(ctx context.Context, args providers.TransfersArgs) (*TransactionResp, error) {
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

	sender, err := senderFromWallet(ctx, a.b, args, fromLA.WalletID)
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

	sender.SenderTrackingNumber = args.FromTransactionID
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
		Sender:   sender,
		Receiver: receiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       args.Amount.Float64(),
			CorrespondentCode:     "GACH",
			DestinationCurrency:   args.Amount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.mts,
			NetAmount:             args.Amount.Float64(),
			OriginalCurrency:      args.Amount.Currency.String(),
			OriginalPaymentMethod: "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:          toID.Address.City,
			ReceiverState:         toID.Address.State,
			SaveSenderReceiver:    true,
			SenderID:              sender.SenderId,
			ServicioCodigo:        "BD",                // DCASH | BD | MW | PIX
			SucursalBanco:         toAcc.RoutingNumber, // Bank branch or routing number
			BankAccount:           toAcc.AccountNumber,
			BankCode:              toAcc.InstitutionCode,
			OfficeCode:            "0",
		},
	})
	if err != nil {
		return nil, err
	}

	if res.Error != 0 {
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s) Status (%s) ID (%s)", res.Error, res.Message, res.Status, res.Password), "external", nil)
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

func (a *Activity) InsertCard2ACH(ctx context.Context, args providers.TransfersArgs) (*TransactionResp, error) {
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

	sender, err := senderFromWallet(ctx, a.b, args, fromLA.WalletID)
	if err != nil {
		return nil, err
	}
	sender.SenderTrackingNumber = args.FromTransactionID

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
		Sender:   sender,
		Receiver: receiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       args.Amount.Float64(),
			CorrespondentCode:     "GACH",
			DestinationCurrency:   args.Amount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.mts,
			NetAmount:             args.Amount.Float64(),
			OriginalCurrency:      args.Amount.Currency.String(),
			OriginalPaymentMethod: "DEBIT", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:          toID.Address.City,
			ReceiverState:         toID.Address.State,
			SaveSenderReceiver:    true,
			SenderID:              sender.SenderId,
			ServicioCodigo:        "BD",                // DCASH | BD | MW | PIX
			SucursalBanco:         toAcc.RoutingNumber, // Bank branch or routing number
			BankAccount:           toAcc.AccountNumber,
			BankCode:              toAcc.InstitutionCode,
			OfficeCode:            "0",
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

func (a *Activity) InsertACH2Card(ctx context.Context, args providers.TransfersArgs) (*ComplianceResp, error) {
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

	fromAcc, err := a.b.MX().GetAccount(ctx, args.FromWalletID, fromLA.ProviderID)
	if errors.Is(err, mx.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	sender, err := senderFromWallet(ctx, a.b, args, fromLA.WalletID)
	if err != nil {
		return nil, err
	}
	sender.SenderTrackingNumber = args.FromTransactionID
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

	res, err := a.ext.InsertTransaction(ctx, external.InsertTransaction{
		Sender:   sender,
		Receiver: receiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       args.Amount.Float64(),
			CorrespondentCode:     "USCD",
			DestinationCurrency:   args.Amount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.mts,
			NetAmount:             args.Amount.Float64(),
			OriginalCurrency:      args.Amount.Currency.String(),
			OriginalPaymentMethod: "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:          toID.Address.City,
			ReceiverState:         toID.Address.State,
			SenderID:              sender.SenderId,
			TipoCuentaCodigo:      "CDN",
			ServicioCodigo:        "BD",
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

func (a *Activity) SaveReceipt(ctx context.Context, tr TransactionResp) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO gmt_receipts (external_id, receipt, licence, right_to_return, error_msg, contact) VALUES ($1, $2, $3, $4, $5, $6)",
		tr.ID, tr.ReceiptRef, tr.Licence, tr.RTR, tr.ErrorMsg, tr.Contact)
	return err
}

func (a *Activity) VerifyTransaction(ctx context.Context, id string) error {
	res, err := a.ext.SetVerified(ctx, external.SetVerified{
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

func (a *Activity) ConfirmPaidNotification(ctx context.Context, externalID string) error {
	resp, err := a.ext.ConfirmPayment(ctx, external.ConfirmPayment{
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

func (a *Activity) UpdateCardTransactionStatus(ctx context.Context, externalID string, txStatus transactions.State) error {
	// GMT status codes : 0=paid, 1=cancelled, 2=transmitted, 3=undo payment
	status := "0"
	if txStatus != transactions.StateCompleted {
		status = "3"
	}

	resp, err := a.ext.UpdateTransactionStatus(ctx, external.UpdateTransactionStatus{
		Reference:  externalID,
		Statuscode: status,
		Date:       external.GMTDate(time.Now()),
	})

	if err != nil {
		return err
	}

	if resp.Error != 0 {
		return temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s)", resp.Error, resp.Message), "external", nil)
	}

	return nil
}
