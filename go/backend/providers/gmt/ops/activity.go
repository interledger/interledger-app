package ops

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/gmt"
	"gitlab.com/fynbos/backend/providers/gmt/external"
)

type Credentials struct {
	Alias    string
	User     string
	Password string
	MTS      int32
}

type Activity struct {
	b     Backends
	ext   external.Service
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
		ext: external.New(),
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

func (a *Activity) CheckIndividualOFAC(ctx context.Context, walletID string) error {
	id, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return err
	}

	res, err := a.ext.OfacVerificationContext(ctx, &external.OfacVerification{
		Alias:     a.creds.Alias,
		User:      a.creds.User,
		Pass:      a.creds.Password,
		LastName:  id.LastName,
		FirstName: id.FirstName,
	})
	if err != nil {
		return err
	}

	if res.OfacVerificationResult.Error != 0 {
		return fmt.Errorf("error code (%d) Message (%s)", res.OfacVerificationResult.Error, res.OfacVerificationResult.Message)
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

	res, err := a.ext.ComplianceCheckContext(ctx, &external.ComplianceCheck{
		Alias:    a.creds.Alias,
		User:     a.creds.User,
		Pass:     a.creds.Password,
		Sender:   sender,
		Receiver: recv,
		Transfer: &external.WsTransferInfo{
			AgenciaCodigo:          "",
			AgencySpecialDiscounts: "",
			AmountToReceive:        1.0,
			CorrespondentCode:      "", // TODO Get from GMT
			DestinationCurrency:    "USD",
			ExchangeRate:           1,
			Fee:                    0,
			MTSID:                  a.creds.MTS,
			NetAmount:              1.0,
			OfficeCode:             "0", // TODO from GMT
			OriginalCurrency:       "USD",
			OriginalPaymentMethod:  "WALLET", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:           id.Address.City,
			ReceiverState:          id.Address.State,
			SaveSenderReceiver:     true,
			SenderAchAccount:       "",
			SenderAchRouting:       "",
			SenderAchType:          "",
			SenderID:               sender.SenderId,
			ServicioCodigo:         "DCASH", // TODO from type of transaction.. WTF
			SucursalBanco:          "",      // Bank branch or routing number
			ThirdPartyReceipt:      "FYN",   // TODO

		},
	})
	if err != nil {
		return nil, err
	}

	if res.ComplianceCheckResult.Error != 0 {
		return nil, fmt.Errorf("error code (%d) message (%s)", res.ComplianceCheckResult.Error, res.ComplianceCheckResult.Message)
	}

	return &ComplianceResp{
		SenderID:         int64(res.ComplianceCheckResult.SenderID),
		SenderWalletID:   walletID,
		ReceiverID:       int64(res.ComplianceCheckResult.ReceiverID),
		ReceiverWalletID: walletID,
	}, nil
}

func (a *Activity) ACHCompliance(ctx context.Context, args gmt.TransfersArgs) (*ComplianceResp, error) {
	fromLA, err := a.b.LinkedAccounts().Get(ctx, args.FromLinkedAccountID)
	if err != nil {
		return nil, err
	}

	toLA, err := a.b.LinkedAccounts().Get(ctx, args.ToLinkedAccountID)
	if err != nil {
		return nil, err
	}

	toID, err := a.b.KYC().GetIndividualDetails(ctx, toLA.WalletID)
	if err != nil {
		return nil, err
	}

	sender, err := senderFromWallet(ctx, a.b, fromLA.WalletID)
	if err != nil {
		return nil, err
	}

	receiver, err := receiverFromWallet(ctx, a.b, toLA.WalletID)
	if err != nil {
		return nil, err
	}

	res, err := a.ext.ComplianceCheckContext(ctx, &external.ComplianceCheck{
		Alias:    a.creds.Alias,
		User:     a.creds.User,
		Pass:     a.creds.Password,
		Sender:   sender,
		Receiver: receiver,
		Transfer: &external.WsTransferInfo{
			AgenciaCodigo:          "",
			AgencySpecialDiscounts: "",
			AmountToReceive:        args.Amount.Float64(),
			CorrespondentCode:      "", // TODO Get from GMT
			DestinationCurrency:    args.Amount.Currency.String(),
			ExchangeRate:           1,
			Fee:                    0,
			MTSID:                  a.creds.MTS,
			NetAmount:              args.Amount.Float64(),
			OfficeCode:             "0", // TODO from GMT
			OriginalCurrency:       args.Amount.Currency.String(),
			OriginalPaymentMethod:  "ACH", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:           toID.Address.City,
			ReceiverState:          toID.Address.State,
			SaveSenderReceiver:     true,
			SenderAchAccount:       "",
			SenderAchRouting:       "",
			SenderAchType:          "",
			SenderID:               0,       // TODO
			ServicioCodigo:         "DCASH", // TODO from type of transaction.. WTF
			SucursalBanco:          "",      // Bank branch or routing number
			ThirdPartyReceipt:      "FYN",   // TODO
		},
	})
	if err != nil {
		return nil, err
	}

	if res.ComplianceCheckResult.Error != 0 {
		return nil, fmt.Errorf("error code (%d) message (%s)", res.ComplianceCheckResult.Error, res.ComplianceCheckResult.Message)
	}

	return &ComplianceResp{
		SenderID:         int64(res.ComplianceCheckResult.SenderID),
		SenderWalletID:   fromLA.WalletID,
		ReceiverID:       int64(res.ComplianceCheckResult.ReceiverID),
		ReceiverWalletID: fromLA.WalletID,
	}, nil
}

func receiverFromWallet(ctx context.Context, b Backends, walletID string) (*external.WsReceiver, error) {
	recvID, err := b.KYC().GetIndividualDetails(ctx, walletID)
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

	gender := "Male"
	if recvID.Gender == kyc.GenderFemale {
		gender = "Female"
	}

	return &external.WsReceiver{
		ReceiverAchAccount:          "",
		ReceiverAchRouting:          "",
		ReceiverAchType:             "",
		ReceiverAddress:             recvID.Address.FormattedAddress,
		ReceiverAverageMonth:        0,
		ReceiverBirthDate:           external.GMTDate(recvID.DateOfBirth),
		ReceiverCity:                recvID.Address.City,
		ReceiverCompany:             "",
		ReceiverCountry:             recvID.Address.CountryCode,
		ReceiverCountryNationallity: "",
		ReceiverCurrency:            "USD",
		ReceiverDocExpiration:       external.GMTDate{},
		ReceiverEmail:               "",
		ReceiverFileImg:             "",
		ReceiverFileImg2:            "",
		ReceiverGender:              gender,
		ReceiverId:                  int32(rid),
		ReceiverLastName:            recvID.LastName,
		ReceiverMobile:              recvUsers[0].PhoneNumber,
		ReceiverMoneyOrigin:         "",
		ReceiverName:                recvID.FirstName,
		ReceiverState:               recvID.Address.State,
		ReceiverZip:                 recvID.Address.ZipCode,
		SenderID:                    int32(sid),
	}, nil
}

func senderFromWallet(ctx context.Context, b Backends, walletID string) (*external.WsSender, error) {
	senderID, err := b.KYC().GetIndividualDetails(ctx, walletID)
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
		Debit:                       false,
		RepresentativeID:            "",
		SenderAchAccount:            "",
		SenderAchRouting:            "",
		SenderAchType:               "",
		SenderAddress:               senderID.Address.FormattedAddress,
		SenderAddressStreet:         senderID.Address.Apartment,
		SenderBank:                  "",
		SenderBirthDate:             external.GMTDate(senderID.DateOfBirth),
		SenderCardBank:              "",
		SenderCardExpiration:        external.GMTDate{},
		SenderCardName:              "",
		SenderCardNumber:            "",
		SenderCardType:              0,
		SenderCity:                  senderID.Address.City,
		SenderCountryCode:           senderID.Address.CountryCode,
		SenderCountryNationallity:   "",
		SenderCountryResidence:      "",
		SenderCurrencyCode:          "USD",
		SenderEmail:                 senderUsers[0].Email,
		SenderForceNew:              false,
		SenderGender:                gender,
		SenderIP:                    senderID.IPAddress,
		SenderId:                    int32(sid),
		SenderIsBusiness:            false,
		SenderLastName:              senderID.LastName,
		SenderMobile:                senderUsers[0].PhoneNumber,
		SenderMonthAverage:          0,
		SenderName:                  senderID.FirstName,
		SenderResidenceAddress:      senderID.Address.String(),
		SenderResidenceAddressExtra: senderID.Address.Apartment,
		SenderResidenceCity:         senderID.Address.City,
		SenderResidenceCountryCode:  senderID.Address.CountryCode,
		SenderResidenceState:        senderID.Address.State,
		SenderResidenceZip:          senderID.Address.ZipCode,
		SenderSendingReason:         "",
		SenderState:                 senderID.Address.State,
		SenderTrackingNumber:        "",
		SenderZip:                   senderID.Address.ZipCode,
	}, nil
}

func getSenderID(ctx context.Context, b Backends, walletID string) (int64, error) {
	var id sql.NullInt64
	err := b.DB().GetContext(ctx, &id, "SELECT sender_id from gmt_users WHERE wallet_id=$1", walletID)
	if err != nil {
		return 0, err
	}

	return id.Int64, nil
}

func getReceiverID(ctx context.Context, b Backends, walletID string) (int64, error) {
	var id sql.NullInt64
	err := b.DB().GetContext(ctx, &id, "SELECT receiver_id from gmt_users WHERE wallet_id=$1", walletID)
	if err != nil {
		return 0, err
	}

	return id.Int64, nil
}

func (a *Activity) UpdateSendRecvUser(ctx context.Context, cr ComplianceResp) error {
	_, err := a.b.DB().ExecContext(ctx, "INSERT INTO gmt_users (wallet_id, sender_id) VALUES ($1, $2)  ON CONFLICT (wallet_id) DO UPDATE SET sender_id = excluded.sender_id, updated_at=now()",
		cr.SenderWalletID, cr.SenderID)
	if err != nil {
		return err
	}

	_, err = a.b.DB().ExecContext(ctx, "INSERT INTO gmt_users (wallet_id, receiver_id) VALUES ($1, $2)  ON CONFLICT (wallet_id) DO UPDATE SET receiver_id = excluded.receiver_id, updated_at=now()",
		cr.ReceiverWalletID, cr.ReceiverID)
	if err != nil {
		return err
	}

	return nil
}
