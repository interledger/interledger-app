package pti

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/currency"
	"github.com/interledger/interledger-app/go/backend/providers/pti/external"
	"github.com/interledger/interledger-app/go/log"
	"github.com/interledger/interledger-app/go/mock/mockbos/db"
	"github.com/interledger/interledger-app/go/mock/mockbos/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type Server struct {
	q    *db.Queries
	conn *pgx.Conn
}

func (s *Server) Register(r chi.Router) {
	r.Post("/users", s.CreateUserHandler)
	r.Get("/users/{userID}", s.GetUser)
	r.Post("/users/assessments", s.StartUserAssessment)
	r.Get("/users/{userID}/assessments", s.GetUserAssessment)
	r.Post("/users/{userID}/wallets", s.CreateUserWalletHandler)
	r.Get("/users/{userID}/wallets", s.ListUserWallets)
	r.Get("/users/{userID}/wallets/{walletID}", s.GetUserWalletHandler)
	r.Post("/transactions/transfers", s.CreateTransfer)
	r.Post("/transactions/deposits", s.CreateDeposit)
	r.Post("/transactions/withdrawals", s.CreateWithdrawal)
	r.Get("/transactions/{requestID}", s.GetTransaction)
	r.Post("/transactions/{requestID}/updates", s.UpdateTransactionStatus)
}

func New(conn *pgx.Conn) *Server {
	return &Server{q: db.New(conn), conn: conn}
}

func (s *Server) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var args external.CreateUserArgs
	err = json.Unmarshal(body, &args)
	if err != nil {
		log.Error("failed to unmarshal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		log.Error("failed to start trx", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	id, err := uuid.Parse(args.ID)
	if err != nil {
		log.Error("failed to parse uuid", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	q := s.q.WithTx(tx)
	u, err := q.CreatePTIUser(ctx, db.CreatePTIUserParams{
		ID:            pgtype.UUID{Bytes: [16]byte(id), Valid: true},
		FirstName:     pgtype.Text{String: args.Name.First, Valid: true},
		MiddleName:    pgtype.Text{String: args.Name.Middle, Valid: true},
		LastName:      pgtype.Text{String: args.Name.Last, Valid: true},
		DateOfBirth:   pgtype.Text{String: args.DateOfBirth, Valid: true},
		SourceOfFunds: pgtype.Text{String: args.SourceOfFunds, Valid: true},
		State:         pgtype.Text{String: "ACTIVE", Valid: true},
		Type:          pgtype.Text{String: args.Type, Valid: true},
	})
	if err != nil {
		log.Error("failed to create user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var emailArgs []db.CreatePTIUserEmailParams
	for _, email := range args.Emails {
		emailArgs = append(emailArgs, db.CreatePTIUserEmailParams{
			UserID:    u.ID,
			Address:   email.Address,
			IsDefault: pgtype.Bool{Bool: email.Default, Valid: true},
		})
	}
	results, err := q.CreatePTIUserEmail(ctx, emailArgs)
	if err != nil || results != int64(len(args.Emails)) {
		log.Error("failed to create email", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var phoneArgs []db.CreatePTIUserPhoneParams
	for _, phone := range args.Phones {
		phoneArgs = append(phoneArgs, db.CreatePTIUserPhoneParams{
			UserID:    u.ID,
			Number:    pgtype.Text{String: phone.Number, Valid: true},
			IsDefault: pgtype.Bool{Bool: phone.Default, Valid: true},
		})
	}
	results, err = q.CreatePTIUserPhone(ctx, phoneArgs)
	if err != nil || results != int64(len(args.Phones)) {
		log.Error("failed to create phone", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var addressArgs []db.CreatePTIUserAddressParams
	for _, address := range args.Addresses {
		addressArgs = append(addressArgs, db.CreatePTIUserAddressParams{
			UserID:     u.ID,
			Street:     pgtype.Text{String: address.Street, Valid: true},
			PostalCode: pgtype.Text{String: address.PostalCode, Valid: true},
			StateCode:  pgtype.Text{String: address.StateCode, Valid: true},
			Country:    pgtype.Text{String: address.Country, Valid: true},
			IsDefault:  pgtype.Bool{Bool: address.Default, Valid: true},
		})
	}
	results, err = q.CreatePTIUserAddress(ctx, addressArgs)
	if err != nil || results != int64(len(args.Addresses)) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.CreateUserResponse{
		ID:   args.ID,
		Link: fmt.Sprintf("http://mockbos.mockbos/pti/users/%s", args.ID),
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateUserWalletHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var args external.CreateWalletArgs
	err = json.Unmarshal(body, &args)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		log.Error("Failed to parse uuid", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	_, err = s.q.CreatePTIWallet(ctx, db.CreatePTIWalletParams{
		ID:       args.WalletID,
		UserID:   pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		Currency: pgtype.Text{String: args.Currency, Valid: true},
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.Wallet{
		WalletID: args.WalletID,
		Currency: args.Currency,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetUserWalletHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parsedUserID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	wallet, err := s.q.GetPTIUserWallet(ctx, db.GetPTIUserWalletParams{
		ID:     chi.URLParam(r, "walletID"),
		UserID: pgtype.UUID{Bytes: [16]byte(parsedUserID), Valid: true},
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	cc := currency.ParseCurrency(wallet.Currency.String)
	amt := currency.FromUInt64(wallet.Balance.Int64, cc)
	resp := external.Wallet{
		WalletID: wallet.ID,
		Currency: cc.String(),
		Balance:  amt.Float64(),
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) ListUserWallets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parsedUserID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	wallets, err := s.q.GetPTIUserWallets(ctx, pgtype.UUID{Bytes: [16]byte(parsedUserID), Valid: true})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := []external.Wallet{}
	for _, w := range wallets {
		cc := currency.ParseCurrency(w.Currency.String)
		amt := currency.FromUInt64(w.Balance.Int64, cc)
		resp = append(resp, external.Wallet{
			WalletID: w.ID,
			Currency: w.Currency.String,
			Balance:  amt.Float64(),
		})
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parsedUserID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	u, err := s.q.GetPTIUser(ctx, pgtype.UUID{Bytes: [16]byte(parsedUserID), Valid: true})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	emails, err := s.q.GetPTIUserEmails(ctx, u.ID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	addresses, err := s.q.GetPTIUserAddresses(ctx, u.ID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	phones, err := s.q.GetPTIUserPhones(ctx, u.ID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.User{
		ID:            parsedUserID.String(),
		Type:          "PERSON",
		Status:        u.State.String,
		SourceOfFunds: u.SourceOfFunds.String,
		Name: &external.Name{
			First:  u.FirstName.String,
			Last:   u.LastName.String,
			Middle: u.MiddleName.String,
		},
		DateOfBirth: u.DateOfBirth.String,
	}
	for _, e := range emails {
		resp.Emails = append(resp.Emails, external.Email{
			Address: e.Address,
			Default: e.IsDefault.Bool,
		})
	}
	for _, a := range addresses {
		resp.Addresses = append(resp.Addresses, external.Address{
			Street:     a.Street.String,
			PostalCode: a.PostalCode.String,
			StateCode:  a.StateCode.String,
			Country:    a.Country.String,
			Default:    a.IsDefault.Bool,
		})
	}
	for _, p := range phones {
		resp.Phones = append(resp.Phones, external.Phone{
			Number:  p.Number.String,
			Type:    "MOBILE",
			Default: p.IsDefault.Bool,
		})
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var payload external.TransferArgs
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	sourceWallet, err := s.q.GetPTIWallet(ctx, payload.SourceTransferMethod.PaymentInformation.ID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	destinationWallet, err := s.q.GetPTIWallet(ctx, payload.SourceTransferMethod.PaymentInformation.ID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	q := s.q.WithTx(tx)

	cc := currency.ParseCurrency(sourceWallet.Currency.String)
	amt := currency.FromFloat64(payload.Amount, cc)

	if sourceWallet.Balance.Int64-int64(amt.Value) < 0 {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	err = q.AdjustPTIWalletBalance(ctx, db.AdjustPTIWalletBalanceParams{
		ID:      sourceWallet.ID,
		Balance: pgtype.Int8{Int64: -int64(amt.Value), Valid: true},
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = q.AdjustPTIWalletBalance(ctx, db.AdjustPTIWalletBalanceParams{
		ID:      destinationWallet.ID,
		Balance: pgtype.Int8{Int64: int64(amt.Value), Valid: true},
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	requestID := r.Header.Get("x-pti-request-id")
	_, err = q.CreatePTITransaction(ctx, db.CreatePTITransactionParams{
		RequestID:        requestID,
		UserID:           sourceWallet.UserID,
		ClientID:         "TEST",
		Date:             pgtype.Text{String: time.Now().Format(time.RFC3339), Valid: true},
		Status:           pgtype.Text{String: "SETTLED", Valid: true},
		TransactionType:  pgtype.Text{String: "TRANSFER", Valid: true},
		Amount:           pgtype.Int8{Int64: int64(amt.Value), Valid: true},
		Currency:         pgtype.Text{String: cc.String(), Valid: true},
		PaymentMethod:    pgtype.Text{String: "WALLET", Valid: true},
		TotalAmount:      pgtype.Int8{Int64: int64(amt.Value), Valid: true},
		TotalCurrency:    pgtype.Text{String: cc.String(), Valid: true},
		SubtotalAmount:   pgtype.Int8{Int64: int64(amt.Value), Valid: true},
		SubtotalCurrency: pgtype.Text{String: cc.String(), Valid: true},
		FeeAmount:        pgtype.Int8{Int64: 0, Valid: true},
		FeeCurrency:      pgtype.Text{String: cc.String(), Valid: true},
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.IDResponse{
		ID: requestID,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var payload CreateDepositArgs
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("failed to unmarshal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	destinationWallet, err := s.q.GetPTIWallet(ctx, payload.DestinationMethod.PaymentInformation.ID)
	if err != nil {
		log.Error("destination wallet not found", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if utils.BytesToUUID(destinationWallet.UserID.Bytes) != payload.Initiator.ID {
		log.Error("wallet does not belong to user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		log.Error("failed to start trx", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	q := s.q.WithTx(tx)

	cc := currency.ParseCurrency(destinationWallet.Currency.String)
	amt := currency.FromFloat64(payload.Amount, cc)

	err = q.AdjustPTIWalletBalance(ctx, db.AdjustPTIWalletBalanceParams{
		ID:      destinationWallet.ID,
		Balance: pgtype.Int8{Int64: int64(amt.Value), Valid: true},
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	requestID := r.Header.Get("x-pti-request-id")
	_, err = q.CreatePTITransaction(ctx, db.CreatePTITransactionParams{
		RequestID:        requestID,
		UserID:           destinationWallet.UserID,
		ClientID:         "TEST",
		Date:             pgtype.Text{String: time.Now().Format(time.RFC3339), Valid: true},
		Status:           pgtype.Text{String: "SETTLED", Valid: true},
		TransactionType:  pgtype.Text{String: "DEPOSIT", Valid: true},
		Amount:           pgtype.Int8{Int64: int64(amt.Value), Valid: true},
		Currency:         pgtype.Text{String: cc.String(), Valid: true},
		PaymentMethod:    pgtype.Text{String: "WALLET", Valid: true},
		TotalAmount:      pgtype.Int8{Int64: int64(amt.Value), Valid: true},
		TotalCurrency:    pgtype.Text{String: cc.String(), Valid: true},
		SubtotalAmount:   pgtype.Int8{Int64: int64(amt.Value), Valid: true},
		SubtotalCurrency: pgtype.Text{String: cc.String(), Valid: true},
		FeeAmount:        pgtype.Int8{Int64: 0, Valid: true},
		FeeCurrency:      pgtype.Text{String: cc.String(), Valid: true},
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.IDResponse{
		ID: requestID,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var payload external.InternalCreateWithdrawalArgs
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Error("failed to unmarshal payload", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	sourceWallet, err := s.q.GetPTIWallet(ctx, payload.SourceMethod.PaymentInformation.WalletID)
	if err != nil {
		log.Error("source wallet not found", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if utils.BytesToUUID(sourceWallet.UserID.Bytes) != payload.Initiator.UserID {
		log.Error("wallet does not belong to user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		log.Error("failed to start trx", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	q := s.q.WithTx(tx)

	cc := currency.ParseCurrency(sourceWallet.Currency.String)
	amt := currency.FromFloat64(payload.Amount, cc)

	err = q.AdjustPTIWalletBalance(ctx, db.AdjustPTIWalletBalanceParams{
		ID:      sourceWallet.ID,
		Balance: pgtype.Int8{Int64: -int64(amt.Value), Valid: true},
	})
	if err != nil {
		log.Error("failed to adjust wallet balance", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	requestID := r.Header.Get("x-pti-request-id")
	_, err = q.CreatePTITransaction(ctx, db.CreatePTITransactionParams{
		RequestID:        requestID,
		UserID:           sourceWallet.UserID,
		ClientID:         "TEST",
		Date:             pgtype.Text{String: time.Now().Format(time.RFC3339), Valid: true},
		Status:           pgtype.Text{String: "SETTLED", Valid: true},
		TransactionType:  pgtype.Text{String: "WITHDRAWAL", Valid: true},
		Amount:           pgtype.Int8{Int64: int64(amt.Value), Valid: true},
		Currency:         pgtype.Text{String: cc.String(), Valid: true},
		PaymentMethod:    pgtype.Text{String: "WALLET", Valid: true},
		TotalAmount:      pgtype.Int8{Int64: int64(amt.Value), Valid: true},
		TotalCurrency:    pgtype.Text{String: cc.String(), Valid: true},
		SubtotalAmount:   pgtype.Int8{Int64: int64(amt.Value), Valid: true},
		SubtotalCurrency: pgtype.Text{String: cc.String(), Valid: true},
		FeeAmount:        pgtype.Int8{Int64: 0, Valid: true},
		FeeCurrency:      pgtype.Text{String: cc.String(), Valid: true},
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := external.IDResponse{
		ID: requestID,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	trx, err := s.q.GetPTITransactionByRequestID(ctx, chi.URLParam(r, "requestID"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	cc := currency.ParseCurrency(trx.Currency.String)
	amt := currency.FromUInt64(trx.Amount.Int64, cc)

	resp := external.TransactionStatus{
		ResourceType:    "TRANSACTION_STATUS",
		RequestID:       trx.RequestID,
		ClientID:        "TEST",
		UserID:          utils.BytesToUUID(trx.UserID.Bytes),
		Date:            trx.Date.String,
		Status:          trx.Status.String,
		TransactionType: trx.TransactionType.String,
		PaymentMethod:   trx.PaymentMethod.String,
		PaymentStatusDetail: external.PaymentStatusDetail{
			ProviderResponseCode:     trx.PaymentStatusProviderResponseCode.String,
			ProviderResponseCategory: trx.PaymentStatusProviderResponseCategory.String,
		},
		Amount:   float64(amt.Value),
		Currency: cc.String(),
		Total: external.Total{
			Fee: external.Cost{
				Amount:   0,
				Currency: cc.String(),
			},
			Total: external.Cost{
				Amount:   amt.Float64(),
				Currency: cc.String(),
			},
			Subtotal: external.Cost{
				Amount:   amt.Float64(),
				Currency: cc.String(),
			},
		},
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) StartUserAssessment(w http.ResponseWriter, r *http.Request) {
	resp := external.IDResponse{
		ID: uuid.NewString(),
	}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetUserAssessment(w http.ResponseWriter, r *http.Request) {
	resp := external.Assessment{
		ResourceType: "USER_ASSESSMENT",
		ClientID:     "TEST",
		RequestID:    r.Header.Get("x-pti-request-id"),
		Date:         time.Now().Format(http.TimeFormat),
		Assessment:   "ACCEPTED",
	}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UpdateTransactionStatus(w http.ResponseWriter, r *http.Request) {
	resp := external.CreateTxResponse{
		ID:   uuid.NewString(),
		Link: "yo mamma",
	}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
