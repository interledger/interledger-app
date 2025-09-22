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

	"gitlab.com/fynbos/backend/country"
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

		timestamp, err := time.Parse(time.RFC3339, wh.Data.Attributes.CreatedAt)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error("persona webhook: failed to process event, unmarhsalling created-at failed.", zap.Error(err), zap.String("event", wh.Data.Attributes.Name), zap.String("eventID", wh.Data.ID))
			return
		}

		switch wh.Data.Attributes.Name {
		case "account.created":
			err = accountCreatedWebhook(r.Context(), b, wh.Data.Attributes.Payload)
		case "account.tag-added":
			err = accountTagAddedWebhook(r.Context(), b, pc, wh.Data.Attributes.Payload)
		case "inquiry.created":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, persona.InquiryCreated, timestamp)
		case "inquiry.started":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, persona.InquiryPending, timestamp)
		case "inquiry.expired":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, persona.InquiryExpired, timestamp)
		case "inquiry.approved":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, persona.InquiryApproved, timestamp)
		case "inquiry.failed":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, persona.InquiryFailed, timestamp)
		case "inquiry.marked-for-review":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, persona.InquiryNeedsReview, timestamp)
			if err == nil {
				notifyPersonaReview(r.Context(), wh.Data.Attributes.Payload)
			}
		case "inquiry.declined":
			err = inquiryWebhook(r.Context(), b, wh.Data.Attributes.Payload, persona.InquiryDeclined, timestamp)
		default:
			log.Warn("persona webhook: unknown event.", zap.String("name", wh.Data.Attributes.Name), zap.String("eventID", wh.Data.ID))
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("persona webhook: failed to process event.", zap.Error(err), zap.String("event", wh.Data.Attributes.Name), zap.String("eventID", wh.Data.ID))
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

	slack.SendToChannel(ctx, slack.ChannelPersona, "wallet-info-bot", fmt.Sprintf("New Persona review in [%s] link [https://app.withpersona.com/dashboard/inquiries/%s]",
		env.GetEnv(), inq.Data.ID))
}

func inquiryWebhook(ctx context.Context, b Backends, js json.RawMessage, state persona.InquiryStatus, timestamp time.Time) error {
	var inq persona.Inquiry
	err := json.Unmarshal(js, &inq)
	if err != nil {
		return err
	}
	if inq.Data.Attributes.ReferenceID == "" {
		log.Error("persona inquiry does not have a referenceID.", zap.String("inquiryID", inq.Data.ID))
		return nil
	}

	var res sql.Result
	if state == persona.InquiryCreated {
		res, err = b.DB().ExecContext(ctx, "INSERT INTO kyc_persona_inquiries(external_id, wallet_id, state) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;",
			inq.Data.ID, inq.Data.Attributes.ReferenceID, state)
	} else {
		res, err = b.DB().ExecContext(ctx, "UPDATE kyc_persona_inquiries SET state=$1, updated_at=$4 WHERE wallet_id=$2 AND external_id=$3 AND (updated_at < $4 OR updated_at IS NULL);",
			state, inq.Data.Attributes.ReferenceID, inq.Data.ID, timestamp)
	}
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows < 1 {
		log.Info("not upating persona inquiry state", zap.Time("timestamp", timestamp), zap.String("inquiryID", inq.Data.ID), zap.String("webhook inquiry state", string(state)))
	}

	return err
}

func accountCreatedWebhook(ctx context.Context, b Backends, js json.RawMessage) error {
	var whAcc persona.Account
	err := json.Unmarshal(js, &whAcc)
	if err != nil {
		return err
	}

	if whAcc.Data.Attributes.ReferenceID == "" {
		log.Error("persona account does not have a referenceID.", zap.String("accountID", whAcc.Data.ID))
		return nil
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO kyc_persona_accounts (external_id, wallet_id) VALUES ($1,$2) ON CONFLICT DO NOTHING;", whAcc.Data.ID, whAcc.Data.Attributes.ReferenceID)
	if err != nil {
		return err
	}
	return err
}

func accountTagAddedWebhook(ctx context.Context, b Backends, pc persona.Client, js json.RawMessage) error {
	var whAcc persona.Account
	err := json.Unmarshal(js, &whAcc)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO kyc_persona_accounts (external_id, wallet_id) VALUES ($1,$2) ON CONFLICT DO NOTHING;", whAcc.Data.ID, whAcc.Data.Attributes.ReferenceID)
	if err != nil {
		log.Error(
			"persona webhook: failed to map persona account to wallet.",
			zap.Error(err),
			zap.String("external accountID", whAcc.Data.ID),
			zap.String("walletID", whAcc.Data.Attributes.ReferenceID),
		)
	}

	// All information may not be included in the webhook, so we get the full info from the API
	externalAcc, err := pc.GetAccount(ctx, whAcc.Data.ID)
	if err != nil {
		return err
	}

	var isDirty bool
	var kycState kyc.Status
	var stateTags int
	for _, tag := range externalAcc.Attributes.Tags {
		switch persona.AccountTag(tag) {
		case persona.AccountTagDirty:
			isDirty = true
		case persona.AccountTagPending:
			kycState = kyc.StatusPending
			stateTags++
		case persona.AccountTagRejected:
			kycState = kyc.StatusDenied
			stateTags++
		case persona.AccountTagVerified:
			// deprecated
			log.Error("persona webhook: received deprecated account tag=(STATUS:VERIFIED).", zap.String("external accountID", whAcc.Data.ID), zap.String("walletID", whAcc.Data.Attributes.ReferenceID))
		case persona.AccountTagReview:
			kycState = kyc.StatusInReview
			stateTags++
		case persona.AccountTagLevel1:
			kycState = kyc.StatusLevel1
			stateTags++
		case persona.AccountTagLevel2:
			kycState = kyc.StatusLevel2
			stateTags++
		}
	}

	if isDirty { // then sync account state
		details := externalAcc.Attributes

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
			log.Error("persona webhook: no inquiry found in database.", zap.String("external accountID", externalAcc.ID), zap.String("walletID", externalAcc.Attributes.ReferenceID))
			return nil
		}

		inq, err := pc.LookupInquiry(ctx, inqID)
		if err != nil {
			log.Error("persona webhook: inquiry not found on Persona.", zap.String("inquiryID", inqID), zap.String("external accountID", externalAcc.ID), zap.String("walletID", externalAcc.Attributes.ReferenceID))
			return err
		}

		inquiryData := inq.Included
		if inq.Data.Attributes.Status != "approved" {
			log.Warn("persona webhook: failed to find approved inquiry.", zap.String("latest inquiryID", inqID), zap.String("status", inq.Data.Attributes.Status))
			inquiryData = []persona.InquiryIncluded{}
		}

		gender := kyc.GenderUnknown
		var ipAddr, birthplace, nationality string
		for _, ii := range inquiryData {
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

		addressSubdivision := strings.ToUpper(strings.TrimSpace(details.AddressSubdivision))
		state, err := country.GetStateCode(country.Country(details.CountryCode), addressSubdivision)
		if err != nil {
			log.Error("persona webhook: unknown state.", zap.Error(err), zap.String("walletID", whAcc.Data.Attributes.ReferenceID), zap.String("externalID", whAcc.Data.ID))
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
				State:       state,
				ZipCode:     strings.TrimSuffix(details.AddressPostalCode, "-0000"), // temp due to persona bug
				CountryCode: details.CountryCode,
			},
		})
		if err != nil {
			return err
		}

		_, err = pc.RemoveTag(ctx, externalAcc.ID, string(persona.AccountTagDirty))
		if err != nil && !errors.Is(err, persona.ErrNotFound) { // ignore if dirty tag is not there
			return err
		}
	}

	// Too many state tags.
	if stateTags > 1 {
		slack.SendToChannel(ctx, slack.ChannelPersona, "wallet-info-bot", fmt.Sprintf("Persona account in [%s] is in unknown state. link [https://app.withpersona.com/dashboard/accounts/%s]",
			env.GetEnv(), externalAcc.ID))
	} else if kycState != 0 {
		err = SetKYCStatus(ctx, b, externalAcc.Attributes.ReferenceID, kycState)
		if err != nil {
			return err
		}
	}

	return nil
}
