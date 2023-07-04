package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
			log.Error("failed to process webhook, unmarhsalling created-at failed.", zap.Error(err))
			return
		}

		switch wh.Data.Attributes.Name {
		case "account.created":
			err = accountCreatedWebhook(r.Context(), b, wh.Data.Attributes.Payload)
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

func inquiryWebhook(ctx context.Context, b Backends, js json.RawMessage, state persona.InquiryStatus, timestamp time.Time) error {
	var inq persona.Inquiry
	err := json.Unmarshal(js, &inq)
	if err != nil {
		return err
	}

	res, err := b.DB().ExecContext(ctx, "UPDATE kyc_persona_inquiries SET state=$1, updated_at=$4 WHERE wallet_id=$2 AND external_id=$3 AND (updated_at > $4 OR updated_at IS NULL);",
		state, inq.Data.Attributes.ReferenceID, inq.Data.ID, timestamp)

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

	_, err = b.DB().ExecContext(ctx, "INSERT INTO kyc_persona_accounts (external_id, wallet_id) VALUES ($1,$2);", whAcc.Data.ID, whAcc.Data.Attributes.ReferenceID)
	if err != nil {
		return err
	}
	return err
}
