package xago

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"gitlab.com/fynbos/backend/providers/xago/external"
	"gitlab.com/fynbos/backend/providers/xago/ops"
	"gitlab.com/fynbos/log"
	"gitlab.com/fynbos/mockbos/db"
	"gitlab.com/fynbos/mockbos/utils"
	"go.uber.org/zap"
)

type Server struct {
	db         *db.Queries
	webhookUrl string
}

type TransactionType string

const (
	Withdrawal TransactionType = "withdrawal"
	Deposit    TransactionType = "deposit"
)

func New(db *db.Queries) *Server {
	webhookUrl := os.Getenv("XAGO_WEBHOOK_URL")
	if webhookUrl == "" {
		log.Warn("xago: webhook url not set")
	}
	return &Server{db: db, webhookUrl: webhookUrl}
}

func (s *Server) CreateSubAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req external.SubAccountReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		depRef := generateDepositReference()

		sa, err := s.db.CreateXagoSubAccount(r.Context(), db.CreateXagoSubAccountParams{
			DepositTag:   depRef,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Email:        req.Email,
			MobileNumber: req.MobileNumber,
			DateOfBirth: pgtype.Date{
				Time:             time.Date(1985, time.July, 23, 0, 0, 0, 0, time.UTC),
				InfinityModifier: pgtype.Finite,
				Valid:            true,
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Mock response
		resp := external.SubAccount{
			AccountID:      utils.BytesToUUID(sa.Bytes),
			DepositAddress: "rF2132313cCDsdcCaasdaeed",
			DepositTag:     91273812,
			DepositDetails: map[string][]external.DepositDetails{
				"ZAR": {
					{
						BankName:       "Capitec Business",
						AccountName:    "Xago Technologies PTY LTD",
						AccountNumber:  "1050835450",
						BankAddress:    "142 West Street, Sandown, 2196",
						AccountAddress: "The Matrix, Bridgeway, Century City, 7441, South Africa",
						BranchCode:     "450105",
					},
					{
						BankName:       "Bidvest Bank",
						AccountName:    "Xago Technologies PTY LTD",
						AccountNumber:  "13874093401",
						BankAddress:    "142 West Street, Sandown, 2196",
						AccountAddress: "The Matrix, Bridgeway, Century City, 7441, South Africa",
						BranchCode:     "462005",
					},
				},
			},
			Beneficiaries: []external.Beneficiaries{
				{
					BeneficiaryID:    uuid.New().String(),
					BeneficiaryType:  "rollup",
					DepositReference: depRef,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			// At this point, since the header and possibly part of the body are already written,
			// you cannot send a new status code or additional headers.
			// Log the error for server-side diagnostics.
			log.Warn("Failed to encode JSON response: %v", zap.Error(err))
			// Consider how you want to handle this knowing you can't change the response sent to the client.
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) AddBeneficiary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req external.CreateBeneficiaryReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		beneficiaryID, err := s.db.CreateXagoBeneficiary(r.Context(), db.CreateXagoBeneficiaryParams{
			Name:          pgtype.Text{String: req.Name, Valid: true},
			Scope:         pgtype.Text{String: req.Scope, Valid: true},
			CurrencyCode:  pgtype.Text{String: req.CurrencyCode, Valid: true},
			AccountNumber: pgtype.Text{String: req.AccountNumber, Valid: true},
			BranchCode:    pgtype.Text{String: req.BranchCode, Valid: true},
			BankName:      pgtype.Text{String: req.BankName, Valid: true},
			BankCountry:   pgtype.Text{String: req.BankCountry, Valid: true},
			AccountName:   pgtype.Text{String: req.AccountName, Valid: true},
			Reference:     pgtype.Text{String: req.Reference, Valid: true},
			Iban:          pgtype.Text{String: req.Iban, Valid: true},
			Bic:           pgtype.Text{String: req.Bic, Valid: true},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get beneficiary and transform to AccountBeneficiaries
		beneficiary, err := s.db.GetXagoBeneficiary(r.Context(), beneficiaryID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b := external.AccountBeneficiaries{
			ID:                 utils.BytesToUUID(beneficiary.ID.Bytes),
			BranchCode:         beneficiary.BranchCode.String,
			Reference:          beneficiary.Reference.String,
			BeneficiaryAddress: beneficiary.BeneficiaryPhysicalAddress.String,
			BankName:           beneficiary.BankName.String,
			AccountNumber:      beneficiary.AccountNumber.String,
			Status:             "active",
			CurrencyCode:       beneficiary.CurrencyCode.String,
			Name:               beneficiary.Name.String,
			Wallet:             nil,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(b); err != nil {
			// At this point, since the header and possibly part of the body are already written,
			// you cannot send a new status code or additional headers.
			// Log the error for server-side diagnostics.
			log.Warn("Failed to encode JSON response: %v", zap.Error(err))
			// Consider how you want to handle this knowing you can't change the response sent to the client.
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) CreateTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var trx external.CreateTransferReq
		if err := json.NewDecoder(r.Body).Decode(&trx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validations
		if trx.Amount < 1.0 {
			http.Error(w, "amount must be greater than 1.0", http.StatusBadRequest)
			return
		}
		// Check if beneficiary exists
		var beneficiaryId pgtype.UUID
		err := beneficiaryId.Scan(trx.BeneficiaryID)
		if err != nil {
			http.Error(w, "beneficiaryId is not a UUID", http.StatusBadRequest)
			return
		}
		b, err := s.db.GetXagoBeneficiary(r.Context(), beneficiaryId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "beneficiary does not exist", http.StatusBadRequest)
			} else {
				http.Error(w, "error getting beneficiary", http.StatusBadRequest)
			}
			return
		}

		var trxID pgtype.UUID
		err = trxID.Scan(uuid.New().String())
		if err != nil {
			http.Error(w, "could not generate transaction id", http.StatusBadRequest)
			return
		}

		t, err := s.db.CreateXagoTransaction(r.Context(), db.CreateXagoTransactionParams{
			ID:             trxID,
			CurrencyCode:   trx.CurrencyCode,
			Amount:         trx.Amount,
			OriginAmount:   trx.Amount,
			Status:         "pending",
			BeneficiaryID:  utils.BytesToUUID(b.ID.Bytes),
			IdempotencyKey: trx.IdempotencyKey,
			Type:           string(Withdrawal),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Mock response
		var resp = fmt.Sprintf("\"%s\"", utils.BytesToUUID(t.ID.Bytes))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(resp)); err != nil {
			// At this point, since the header and possibly part of the body are already written,
			// you cannot send a new status code or additional headers.
			// Log the error for server-side diagnostics.
			log.Warn("Failed to encode JSON response: %v", zap.Error(err))
			// Consider how you want to handle this knowing you can't change the response sent to the client.
		}
	}
}

func (s *Server) GetTransaction() http.HandlerFunc {
	type transaction struct {
		Id             string    `json:"id"`
		CurrencyCode   string    `json:"currencyCode"`
		Commission     int       `json:"commission"`
		Total          float64   `json:"total"`
		Amount         float64   `json:"amount"`
		BatchType      string    `json:"batchType"`
		Status         string    `json:"status"`
		AddFeeToAmount bool      `json:"addFeeToAmount"`
		UpdatedAt      time.Time `json:"updatedAt"`
		CreatedAt      time.Time `json:"createdAt"`
		Reference      string    `json:"reference"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var trxId pgtype.UUID
		err := trxId.Scan(r.URL.Query().Get("transactionId"))
		if err != nil {
			http.Error(w, "id is not a UUID", http.StatusBadRequest)
			return
		}

		trx, err := s.db.GetXagoTransaction(r.Context(), trxId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "transaction does not exist", http.StatusBadRequest)
			} else {
				http.Error(w, "error getting transaction", http.StatusBadRequest)
			}
			return
		}
		var resp = transaction{
			Id:             utils.BytesToUUID(trx.ID.Bytes),
			CurrencyCode:   trx.CurrencyCode,
			Commission:     0,
			Total:          trx.Amount,
			Amount:         trx.Amount,
			BatchType:      "transfer",
			Status:         trx.Status,
			AddFeeToAmount: false,
			UpdatedAt:      trx.UpdatedAt.Time,
			CreatedAt:      trx.CreatedAt.Time,
			Reference:      "",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			// At this point, since the header and possibly part of the body are already written,
			// you cannot send a new status code or additional headers.
			// Log the error for server-side diagnostics.
			log.Warn("Failed to encode JSON response: %v", zap.Error(err))
			// Consider how you want to handle this knowing you can't change the response sent to the client.
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) ListDepositTransactions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Defaults
		page := 1
		limit := 10

		if r.URL.Query().Has("page") {
			pageString := r.URL.Query().Get("page")
			i, err := strconv.Atoi(pageString)
			if err != nil {
				http.Error(w, "invalid page", http.StatusBadRequest)
				return
			}
			page = i
		}

		if r.URL.Query().Has("limit") {
			limitString := r.URL.Query().Get("limit")
			i, err := strconv.Atoi(limitString)
			if err != nil {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = i
		}

		if page < 1 {
			page = 1
		}
		if limit <= 0 || limit >= 50 {
			limit = 10 // default limit
		}
		offset := (page - 1) * limit

		d, err := s.db.ListXagoDeposits(r.Context(), db.ListXagoDepositsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "error getting deposits", http.StatusBadRequest)
			return
		}

		deps := make([]external.Deposit, 0)

		for _, xagoTransaction := range d {
			deps = append(deps, external.Deposit{
				TransactionID:          utils.BytesToUUID(xagoTransaction.ID.Bytes),
				AccountID:              xagoTransaction.BeneficiaryID,
				OriginAmount:           xagoTransaction.Amount,
				Amount:                 xagoTransaction.Amount,
				Status:                 xagoTransaction.Status,
				IsRequested:            false,
				IsDuplicate:            false,
				DuplicateTransactionID: "", // What does this represent?
				CreatedAt:              xagoTransaction.CreatedAt.Time,
				SettledAt:              "",
			})
		}

		resp := external.ListDepositsResponse{
			// Mock pagination
			Pagination: external.Pagination{
				NumberOfPages: 1,
				Limit:         10,
				PageNumber:    1,
			},
			Deposits: deps,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			// At this point, since the header and possibly part of the body are already written,
			// you cannot send a new status code or additional headers.
			// Log the error for server-side diagnostics.
			log.Warn("Failed to encode JSON response: %v", zap.Error(err))
			// Consider how you want to handle this knowing you can't change the response sent to the client.
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) CreateDeposit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Warn("here1")

		var req external.TestDepositReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if sub account exists
		log.Warn("here2")
		sa, err := s.db.GetXagoSubAccountByDepositReference(r.Context(), req.DepositReference)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "account does not exist", http.StatusBadRequest)
			} else {
				http.Error(w, "error getting account", http.StatusBadRequest)
			}
			return
		}

		log.Warn("here3")

		var trxID pgtype.UUID
		err = trxID.Scan(req.BankTransactionID)
		if err != nil {
			http.Error(w, "bankTransactionId is not a UUID", http.StatusBadRequest)
			return
		}

		log.Warn("here4")

		t, err := s.db.CreateXagoTransaction(r.Context(), db.CreateXagoTransactionParams{
			ID:             trxID,
			CurrencyCode:   "ZAR",
			Amount:         req.Amount,
			OriginAmount:   req.Amount,
			Status:         "Success",
			BeneficiaryID:  utils.BytesToUUID(sa.ID.Bytes),
			IdempotencyKey: "",
			Type:           string(Deposit),
		})

		if err != nil {
			http.Error(w, "error creating deposit", http.StatusBadRequest)
			return
		}

		log.Warn("here5")

		if s.webhookUrl != "" {
			sendWebhook(r.Context(), s.webhookUrl, t)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func sendWebhook(ctx context.Context, webhookUrl string, trx db.XagoTransaction) {
	webhookReq := ops.Webhook{
		AccountID:              trx.BeneficiaryID,
		Amount:                 trx.Amount,
		BeneficiaryID:          trx.BeneficiaryID,
		CreatedAt:              trx.CreatedAt.Time.String(),
		Currency:               trx.CurrencyCode,
		DuplicateTransactionID: "",
		IsDuplicate:            false,
		IsRequested:            false,
		IsRequestMatched:       false,
		OriginAmount:           trx.OriginAmount,
		ParentExtension:        "",
		SettledAt:              "",
		Code:                   200,
		Status:                 trx.Status,
		TransactionID:          utils.BytesToUUID(trx.ID.Bytes),
		TransactionReference:   "",
		RequestData: struct {
			Amount          float64 `json:"amount"`
			Currency        string  `json:"currencyCode"`
			CustomRequestID string  `json:"customRequestId"`
		}(struct {
			Amount          float64
			Currency        string
			CustomRequestID string
		}{
			Amount:          0,
			Currency:        "",
			CustomRequestID: "",
		}),
	}
	marshalled, err := json.Marshal(webhookReq)
	if err != nil {
		log.Error("xago: error marshalling webhook request")
	}
	req, err := http.NewRequestWithContext(ctx, "POST", webhookUrl, bytes.NewReader(marshalled))
	if err != nil {
		log.Error("xago: could not build request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	// send the request
	res, err := client.Do(req)
	if err != nil {
		log.Error("xago: could not to send request")
	}

	defer res.Body.Close()
}

func (s *Server) CreateLogin() http.HandlerFunc {
	type tokenResp struct {
		Token string `json:"tokenValue"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := tokenResp{Token: "randomtoken"}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			// At this point, since the header and possibly part of the body are already written,
			// you cannot send a new status code or additional headers.
			// Log the error for server-side diagnostics.
			log.Warn("Failed to encode JSON response: %v", zap.Error(err))
			// Consider how you want to handle this knowing you can't change the response sent to the client.
		}
		w.WriteHeader(http.StatusOK)
	}
}

func generateDepositReference() string {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	code := make([]byte, 6)
	for i := range code {
		code[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return fmt.Sprintf("test%s", string(code))
}
