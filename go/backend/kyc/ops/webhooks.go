package ops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/persona"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func NewHandlePersonaWebhook(b Backends) http.HandlerFunc {
	pc := persona.New()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("failed to read webhook body", zap.Error(err))
			return
		}

		if !pc.ValidateWebhook(r, data) {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("failed to validate webhook sig")
			return
		}

		var wh persona.Webhook

		err = json.Unmarshal(data, &wh)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("failed to process webhook, unmarhsalling failed", zap.Error(err))
			return
		}

		switch wh.Data.Attributes.Name {
		case "account.created":
			err = accountCreatedWebhook(r.Context(), b, pc, wh.Data.Attributes.Payload)
		case "inquiry.approved":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, kyc.StatusApproved)
		case "inquiry.maked-for-review":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, kyc.StatusInReview)
		case "inquiry.expired":
			err = inquiryExpiredWebhook(r.Context(), b, wh.Data.Attributes.Payload)
		case "inquiry.declined":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, kyc.StatusDenied)
		default:
			log.Info("unknown persona webhook event", zap.String("name", wh.Data.Attributes.Name))
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("failed to process webhook for account creation", zap.Error(err), zap.String("event", wh.Data.Attributes.Name))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func inquiryExpiredWebhook(ctx context.Context, b Backends, js json.RawMessage) error {
	var inq persona.Inquiry
	err := json.Unmarshal(js, &inq)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE kyc_persona_inquiries SET state=$1, updated_at=now() WHERE wallet_id=$2 AND external_id=$3",
		inq.Data.Attributes.Status, inq.Data.Attributes.ReferenceID, inq.Data.ID)

	return err
}

func inquiryWebhook(ctx context.Context, b Backends, js json.RawMessage, status kyc.Status) error {
	var inq persona.Inquiry
	err := json.Unmarshal(js, &inq)
	if err != nil {
		return err
	}

	err = SetKYCStatus(ctx, b, inq.Data.Attributes.ReferenceID, status)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE kyc_persona_inquiries SET state=$1, updated_at=now() WHERE wallet_id=$2 AND external_id=$3",
		inq.Data.Attributes.Status, inq.Data.Attributes.ReferenceID, inq.Data.ID)

	return err
}

func accountCreatedWebhook(ctx context.Context, b Backends, pc persona.Client, js json.RawMessage) error {
	var whAcc persona.Account
	err := json.Unmarshal(js, &whAcc)
	if err != nil {
		return err
	}

	// All information may not be included in the webhook, so we get the full info from the API
	acc, err := pc.GetAccount(ctx, whAcc.Data.ID)
	if err != nil {
		return err
	}

	details := acc.Attributes

	fn := details.NameFirst
	if details.NameMiddle != "" {
		fn += " " + details.NameMiddle
	}

	dob, err := time.Parse("2006-01-02", details.Birthdate)
	if err != nil {
		return err
	}

	_, err = UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:    details.ReferenceID,
		FirstName:   fn,
		LastName:    details.NameLast,
		CountryCode: details.CountryCode,
		Gender:      kyc.GenderUnknown,
		DateOfBirth: dob,

		Address: &kyc.Address{
			Line1:       details.AddressStreet1,
			Line2:       details.AddressStreet2,
			City:        details.AddressCity,
			State:       details.CountryCode + "-" + details.AddressSubdivision, // US-CA for example
			ZipCode:     details.AddressPostalCode,
			CountryCode: details.CountryCode,
		},
	})
	if err != nil {
		return err
	}

	err = SetKYCStatus(ctx, b, details.ReferenceID, kyc.StatusApproved)

	return err
}
