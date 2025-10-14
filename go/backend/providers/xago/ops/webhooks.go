package ops

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/fynbos/backend/transactions"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/xago"
	"gitlab.com/fynbos/log"
	"gitlab.com/fynbos/pacioli"
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
	Code                   int     `json:"code"`
	Status                 string  `json:"status"`
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
		defer r.Body.Close()

		var hook Webhook
		err = json.Unmarshal(raw, &hook)
		if err != nil {
			log.Error("failed to unmarshal xago webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// check Xago transaction exists
		xagoTransaction, err := b.External().GetDeposit(r.Context(), hook.TransactionID)
		if err != nil {
			log.Error("failed to get xago transaction for webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return

		}
		if xagoTransaction == nil {
			log.Error("failed to get xago transaction for xago webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if xagoTransaction.Amount != hook.Amount {
			log.Error("failed verify amount on xago webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info("xago webhook received", zap.Any("webhook", hook), zap.Any("xagoTransaction", xagoTransaction))
		if hook.Code != 104 {
			log.Error("unsupported xago webhook received", zap.String("webhook", string(raw)))
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		subAcc, err := LookupByAccountID(r.Context(), b, hook.AccountID)
		if err != nil {
			log.Error("failed to find sub account for xago webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		lal, err := b.LinkedAccounts().ListByWalletId(r.Context(), subAcc.WalletID)
		if err != nil {
			log.Error("failed to list linked accounts for xago webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cc := currency.ParseCurrency(hook.Currency)
		if cc != currency.ZAR && cc != currency.USD {
			log.Error("invalid currency for xago webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		amt := currency.FromFloat64(hook.Amount, cc)
		var acc linkedaccounts.LinkedAccount
		for _, la := range lal {
			if la.Provider == xago.ProviderName && la.Type == xago.AccTypeBalance && la.SendCurrency == cc {
				acc = la
				break
			}
		}
		if acc.ID == "" {
			log.Error("failed to find balance linked account for xago webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// This is an idempotent call so we can safely do it repeatedly
		trxID, err := b.Transactions().CreateTransaction(r.Context(), transactions.CreateTransactionArgs{
			ID:                      hook.TransactionID,
			WalletID:                acc.WalletID,
			ForeignID:               hook.TransactionID,
			ForeignType:             transactions.TransactionTypeDeposit,
			Provider:                transactions.ProviderXago,
			State:                   transactions.StateCompleted,
			Note:                    "Deposit received",
			Source:                  "Bank Deposit",
			Title:                   "Deposit",
			Destination:             acc.WalletID,
			Amount:                  amt,
			LinkedAccountTitle:      acc.Title(),
			DestinationIdentity:     acc.WalletID,
			DestinationIdentityType: "WalletID",
		})
		if err != nil {
			log.Error("failed to create transaction for xago webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		opsAcc := xago.USDOpsAccount
		ledger := xago.LedgerIDUSD
		if cc == currency.ZAR {
			opsAcc = xago.ZAROpsAccount
			ledger = xago.LedgerIDZAR
		}

		// Also an idempotent call
		tr, err := b.Pacioli().CreateTransfers(r.Context(), []pacioli.CreateTransferArgs{
			{
				ID:              hook.TransactionID,
				Amount:          amt.Value,
				DebitAccountID:  opsAcc,
				CreditAccountID: acc.ID,
				Pending:         false,
				Code:            1,
				Ledger:          ledger,
			},
		})
		if err != nil {
			log.Error("failed to find create pacioli transactions for xago webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = b.Transactions().AddTransfers(r.Context(), trxID, []transactions.TransferArgs{
			{
				LinkedAccountID: acc.ID,
				ForeignID:       hook.TransactionID,
				Type:            transactions.TransferTypeCreditBalance,
				Amount:          amt,
				State:           transactions.StateCompleted,
			},
		})
		if err != nil {
			log.Error("failed to find create transfer for transaction for xago webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(tr) == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}
		if tr[0].Code != 0 {
			log.Error("failed to find create pacioli transactions for xago webhook", zap.String("code", tr[0].Code.String()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}

}
