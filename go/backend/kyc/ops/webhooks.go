package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/kyc/persona"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/env"
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
			log.Error("failed to validate webhook sig", zap.String("signature", r.Header.Get("Persona-Signature")))
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
		case "inquiry.marked-for-review":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, kyc.StatusInReview)
			if err == nil {
				notifyPersonaReview(r.Context(), wh.Data.Attributes.Payload)
			}
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

func notifyPersonaReview(ctx context.Context, js json.RawMessage) {
	var inq persona.Inquiry
	err := json.Unmarshal(js, &inq)
	if err != nil {
		log.Error("failed to send notify slack of Persona", zap.Error(err))
		return
	}

	slack.SendToChannel(ctx, slack.PersonaChannel, "FynBOT", fmt.Sprintf("New Persona review in [%s] link [https://app.withpersona.com/dashboard/inquiries/%s]",
		env.GetEnv(), inq.Data.ID))
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

	// Lookup the latest enquiry for the users IP Address
	var inqID string
	err = b.DB().GetContext(ctx, &inqID, "SELECT external_id FROM kyc_persona_inquiries WHERE wallet_id=$1 AND state=$2 ORDER BY updated_at DESC",
		details.ReferenceID, "approved")
	if errors.Is(err, sql.ErrNoRows) {
		// Lookup the latest inquiry as the webhook may not have fired to update it's status, We'll double check it
		err = b.DB().GetContext(ctx, &inqID, "SELECT external_id FROM kyc_persona_inquiries WHERE wallet_id=$1 ORDER BY updated_at DESC",
			details.ReferenceID)
	}
	if err != nil {
		return err
	}

	inq, err := pc.LookupInquiry(ctx, inqID)
	if err != nil {
		return err
	}

	if inq.Data.Attributes.Status != "approved" {
		return fmt.Errorf("failed to find approved inquiry, latest enquiry (%s) status (%s)", inqID, inq.Data.Attributes.Status)
	}

	gender := kyc.GenderUnknown
	var ipAddr, birthplace, nationality string
	for _, ii := range inq.Included {
		// These are in order so the last one in the list will always be the most up to date.
		if ii.Type == "inquiry-session" && ii.Attributes.IPAddress != "" {
			ipAddr = ii.Attributes.IPAddress
		}

		if ii.Attributes.Nationality != "" {
			nationality = ii.Attributes.Nationality
		}

		if ii.Attributes.Birthplace != "" {
			birthplace = ii.Attributes.Birthplace
		}

		if strings.EqualFold(ii.Attributes.Gender, "Male") {
			gender = kyc.GenderMale
		}
		if strings.EqualFold(ii.Attributes.Gender, "Female") {
			gender = kyc.GenderFemale
		}
	}

	_, err = UpdateIndividualDetails(ctx, b, kyc.IndividualDetails{
		WalletID:     details.ReferenceID,
		FirstName:    fn,
		LastName:     details.NameLast,
		CountryCode:  details.CountryCode,
		Gender:       gender,
		DateOfBirth:  dob,
		PlaceOfBirth: birthplace,
		Nationality:  nationality,
		IPAddress:    ipAddr,
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
