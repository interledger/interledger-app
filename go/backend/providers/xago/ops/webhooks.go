package ops

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type Webhook struct {
	AccountID              string  `json:"accountId"`
	Amount                 float64 `json:"amount"`
	BeneficiaryID          string  `json:"beneficiaryId"`
	CreatedAt              string  `json:"createdAt"`
	Currency               string  `json:"currencyCode"`
	DuplicateTransactionID string  `json:"duplicateTransactionId"`
	IsDuplicate            bool    `json:"isDuplicate"`
	IsRequested            bool    `json:"isRequested"`
	IsRequestMatched       bool    `json:"isRequestMatched"`
	OriginAmount           float64 `json:"originAmount"`
	ParentExtension        string  `json:"parentExtension"`
	SettledAt              string  `json:"settledAt"`
	StatusCode             string  `json:"statusCode"`
	StatusMessage          string  `json:"statusMessage"`
	TransactionID          string  `json:"transactionId"`
	TransactionReference   string  `json:"transactionReference"`
	RequestData            struct {
		Amount          float64 `json:"amount"`
		Currency        string  `json:"currencyCode"`
		CustomRequestID string  `json:"customRequestId"`
	} `json:"requestData"`
}

func EventWebhook(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("failed to read xago webhook body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var hook Webhook
		err = json.Unmarshal(raw, &hook)
		if err != nil {
			log.Error("failed to unmarshal xago webhook", zap.Error(err))
		}

		log.Info("xago webhook unsupported")
		w.WriteHeader(http.StatusNotImplemented)
	}
}
