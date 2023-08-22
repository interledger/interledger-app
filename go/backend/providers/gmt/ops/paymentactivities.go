package ops

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/identities"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gmt/external"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
)

func (a *Activity) CheckPaymentSenderOFAC(ctx context.Context, paymentID string) error {
	p, err := a.b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return err
	}

	senderWallet, err := lookupWallet(ctx, a.b, p.Sender)
	if err != nil {
		return err
	}

	err = a.CheckWalletOFAC(ctx, senderWallet.ID)
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

	receiverWallet, err := lookupWallet(ctx, a.b, p.Receiver)
	if err != nil {
		return err
	}

	err = a.CheckWalletOFAC(ctx, receiverWallet.ID)
	if err != nil {
		return err
	}

	return nil
}

func lookupWallet(ctx context.Context, b Backends, identity payments.Identity) (*wallets.Wallet, error) {
	var resp *wallets.Wallet
	var err error
	switch identity.Type {
	case payments.IdentityTypeWalletID:
		resp, err = b.Wallets().Get(ctx, identity.Identifier)
	case payments.IdentityTypeWalletURL:
		resp, err = b.Wallets().GetFromAddress(ctx, identity.Identifier)
	case payments.IdentityTypeTwitter:
		var id *identities.Identity
		id, err = b.Identities().GetByIdentifier(ctx, identity.Identifier)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(string(id.Platform), string(identities.PlatformTwitter)) {
			return nil, fmt.Errorf("identifier (%s) type mismatch expected (%s) got (%s)", identity.Identifier, identities.PlatformTwitter, identity.Type)
		}
		resp, err = b.Wallets().Get(ctx, id.WalletID)
	default:
		return nil, fmt.Errorf("unknown identity type %s", identity.Type)
	}
	return resp, err
}

func (a *Activity) PaymentCompliance(ctx context.Context, paymentID string) (*ComplianceResp, error) {
	p, err := a.b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	senderWallet, err := lookupWallet(ctx, a.b, p.Sender)
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

	gmtSender, err := senderFromPayment(ctx, a.b, senderWallet.ID)
	if err != nil {
		return nil, err
	}
	gmtSender.SenderTrackingNumber = p.SendTransactionID

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
		SenderWalletID:   senderWallet.ID,
		ReceiverID:       int64(res.ReceiverID),
		ReceiverWalletID: receiverAcc.WalletID,
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

	senderWallet, err := lookupWallet(ctx, b, p.Sender)
	if err != nil {
		return nil, err
	}

	senderID, err := b.KYC().GetIndividualDetails(ctx, senderWallet.ID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	senderUsers, err := b.Users().ListUsers(ctx, senderWallet.ID)
	if err != nil {
		return nil, err
	}

	sid, err := getSenderID(ctx, b, senderWallet.ID)
	if err != nil {
		return nil, err
	}

	ipAddress := senderID.IPAddress
	// TODO: Add IP address to the payment
	/*if args.IPAddress != "" {
		ipAddress = args.IPAddress
	}*/

	gender := "Male"
	if senderID.Gender == kyc.GenderFemale {
		gender = "Female"
	}

	exceeds, err := b.Limits().ExceedsGMTLimits(ctx, senderWallet.ID, p.SenderAmount)
	if err != nil {
		return nil, err
	}

	var address string
	if senderID.Address != nil {
		address = senderID.Address.String()
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
	}

	if !exceeds {
		return sender, nil
	}

	idNums, err := b.KYC().GetPersonaIDNumbers(ctx, senderWallet.ID)
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
