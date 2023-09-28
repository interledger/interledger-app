package ops

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gmt/external"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
)

func (a *Activity) PaymentNeedsCompliance(ctx context.Context, paymentID string) (bool, error) {
	p, err := a.b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return false, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return false, err
	}

	return p.Type == payments.TypePeer2Peer, nil

}

func (a *Activity) CheckPaymentSenderOFAC(ctx context.Context, paymentID string) error {
	p, err := a.b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return err
	}

	err = a.CheckWalletOFAC(ctx, p.Sender.WalletID)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) CheckPaymentReceiverOFAC(ctx context.Context, paymentID string) error {
	p, err := a.b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return err
	}

	err = a.CheckWalletOFAC(ctx, p.Receiver.WalletID)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) PaymentCompliance(ctx context.Context, paymentID string) (*ComplianceResp, error) {
	p, err := a.b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	receiverAcc, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	receiverKYC, err := a.b.KYC().GetIndividualDetails(ctx, receiverAcc.WalletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	gmtSender, err := senderFromPayment(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	gmtReceiver, err := receiverFromWallet(ctx, a.b, receiverAcc.WalletID)
	if err != nil {
		return nil, err
	}

	res, err := a.ext.ComplianceCheck(ctx, external.ComplianceCheck{
		Sender:   gmtSender,
		Receiver: gmtReceiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       p.ReceiverAmount.Float64(),
			CorrespondentCode:     "USCD",
			DestinationCurrency:   p.ReceiverAmount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.mts,
			OfficeCode:            "0",
			NetAmount:             p.SenderAmount.Float64(),
			OriginalCurrency:      p.SenderAmount.Currency.String(),
			OriginalPaymentMethod: "DEBIT", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:          receiverKYC.Address.City,
			ReceiverState:         receiverKYC.Address.State,
			SenderID:              gmtSender.SenderId,
			ServicioCodigo:        "BD",
		},
	})
	if err != nil {
		return nil, err
	}

	// Check status for actual failures
	if res.Error == 999 || // Required fields missing
		res.Error == 1000 || // Invalid User
		res.Error == 1002 || // Cannot find or insert sender
		res.Error == 1003 || // Cannot find or insert receiver
		res.Error == 2000 || // Amount over limit
		res.Error == 2001 || // Compliance fields missing
		res.Error == 2002 || // Fields missing for validation
		res.Error == 2003 { // Sender or receiver blocked
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s)", res.Error, res.Message), "external", nil)
	}

	return &ComplianceResp{
		SenderID:         int64(res.SenderID),
		SenderWalletID:   p.Sender.WalletID,
		ReceiverID:       int64(res.ReceiverID),
		ReceiverWalletID: p.Receiver.WalletID,
	}, nil
}

func senderFromPayment(ctx context.Context, b Backends, paymentID string) (*external.WsSender, error) {
	p, err := b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	senderID, err := b.KYC().GetIndividualDetails(ctx, p.Sender.WalletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	senderUsers, err := b.Users().ListUsers(ctx, p.Sender.WalletID)
	if err != nil {
		return nil, err
	}

	sid, err := getSenderID(ctx, b, p.Sender.WalletID)
	if err != nil {
		return nil, err
	}

	ipAddress := senderID.IPAddress
	if p.IPAddress != "" {
		ipAddress = p.IPAddress
	}

	gender := "Male"
	if senderID.Gender == kyc.GenderFemale {
		gender = "Female"
	}

	exceeds, err := b.Limits().ExceedsGMTLimits(ctx, p.Sender.WalletID, p.SenderAmount)
	if err != nil {
		return nil, err
	}

	var address string
	if senderID.Address != nil {
		address = senderID.Address.String()
	}

	trackingNumber := p.SendTransactionID
	if trackingNumber == "" {
		trackingNumber = p.ID
	}
	sender := &external.WsSender{
		SenderAddress:               address,
		SenderAddressStreet:         senderID.Address.Apartment,
		SenderBirthDate:             external.GMTDate(senderID.DateOfBirth),
		SenderCity:                  senderID.Address.City,
		SenderCountryCode:           senderID.Address.CountryCode,
		SenderCurrencyCode:          p.SenderAmount.Currency.String(),
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
		SenderCountryNationallity:   senderID.Nationality,
		SenderPOB:                   senderID.PlaceOfBirth,
		SenderTrackingNumber:        trackingNumber,
	}

	if !exceeds {
		return sender, nil
	}

	idNums, err := b.KYC().GetPersonaIDNumbers(ctx, p.Sender.WalletID)
	if err != nil {
		return nil, err
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

	return sender, nil
}

func (a *Activity) PaymentGMTTransaction(ctx context.Context, paymentID string) (*TransactionResp, error) {
	p, err := a.b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	receiverAcc, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	receiverKYC, err := a.b.KYC().GetIndividualDetails(ctx, receiverAcc.WalletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	gmtSender, err := senderFromPayment(ctx, a.b, paymentID)
	if err != nil {
		return nil, err
	}

	gmtReceiver, err := receiverFromWallet(ctx, a.b, receiverAcc.WalletID)
	if err != nil {
		return nil, err
	}

	// We are paying out to a card, so use the last 4 digits as the bank account. The Mask for the linked account is the Last 4 digits in Tabapay
	var bankAccNum string
	if receiverAcc.Provider == tabapay.ProviderName {
		bankAccNum = strings.TrimSpace(strings.ReplaceAll(receiverAcc.Mask, "*", ""))
	}

	res, err := a.ext.InsertTransaction(ctx, external.InsertTransaction{
		Sender:   gmtSender,
		Receiver: gmtReceiver,
		Transfer: &external.WsTransferInfo{
			AmountToReceive:       p.ReceiverAmount.Float64(),
			CorrespondentCode:     "USCD",
			BankAccount:           bankAccNum,
			DestinationCurrency:   p.ReceiverAmount.Currency.String(),
			ExchangeRate:          1,
			Fee:                   0,
			MTSID:                 a.mts,
			OfficeCode:            "0",
			NetAmount:             p.SenderAmount.Float64(),
			OriginalCurrency:      p.SenderAmount.Currency.String(),
			OriginalPaymentMethod: "DEBIT", // ACH | CHECK | WALLET | CASH | DEBIT | WIRE
			ReceiverCity:          receiverKYC.Address.City,
			ReceiverState:         receiverKYC.Address.State,
			SenderID:              gmtSender.SenderId,
			ServicioCodigo:        "BD",
		},
	})
	if err != nil {
		return nil, err
	}

	// Check status for actual failures
	if res.Error == 999 || // Required fields missing
		res.Error == 1000 || // Invalid User
		res.Error == 1002 || // Cannot find or insert sender
		res.Error == 1003 || // Cannot find or insert receiver
		res.Error == 2000 || // Amount over limit
		res.Error == 2001 || // Compliance fields missing
		res.Error == 2002 || // Fields missing for validation
		res.Error == 2003 { // Sender or receiver blocked
		return nil, temporal.NewNonRetryableApplicationError(fmt.Sprintf("error code (%d) Message (%s)", res.Error, res.Message), "external", nil)
	}

	if res.Error != 0 {
		return nil, fmt.Errorf("error code (%d) Message (%s)", res.Error, res.Message)
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
