package ops

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type webhook struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type incomingPaymentData struct {
	Payment struct {
		ID               string `json:"id"`
		PaymentPointerID string `json:"paymentPointerId"`
		CreatedAt        string `json:"createdAt"`
		ExpiresAt        string `json:"expiresAt"`
		ReceivedAmount   amount `json:"receivedAmount"`
		Completed        bool   `json:"completed"`
	} `json:"incomingPayment"`
}

type amount struct {
	Value      string `json:"value"`
	AssetCode  string `json:"assetCode"`
	AssetScale int    `json:"assetScale"`
}

func EventWebhook(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("failed to read rafiki webhook body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var hook webhook
		err = json.Unmarshal(raw, &hook)
		if err != nil {
			log.Error("failed to unmarshal rafiki webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch hook.Type {
		case "incoming_payment.created", "incoming_payment.completed", "incoming_payment.expired":
			var payment incomingPaymentData
			err = json.Unmarshal(hook.Data, &payment)
			if err != nil {
				log.Error("failed to unmarshal rafiki incoming payment", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			completed := payment.Payment.Completed || hook.Type == "incoming_payment.completed" || hook.Type == "incoming_payment.expired"

			var amt uint64
			amt, err = strconv.ParseUint(payment.Payment.ReceivedAmount.Value, 10, 64)
			if err != nil {
				log.Error("failed to convert rafiki incoming payment amount", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			_, err = b.DB().ExecContext(r.Context(), `INSERT INTO rafiki_incoming_payments 
  (id, payment_pointer_id, completed, received_amount, received_amount_asset) 
	VALUES 
  ($1, $2, $3, $4, $5) ON CONFLICT (id) 
    DO UPDATE SET 
                completed = EXCLUDED.completed, 
                received_amount = EXCLUDED.received_amount,
                updated_at = now()`, payment.Payment.ID, payment.Payment.PaymentPointerID, completed, amt, payment.Payment.ReceivedAmount.AssetCode)
			if err != nil {
				log.Error("failed to upsert rafiki incoming payment", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			log.Info("rafiki unsupported webhook type", zap.String("type", hook.Type))
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
