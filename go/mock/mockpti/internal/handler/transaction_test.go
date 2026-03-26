package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/fynbos/mock/mockpti/internal/jobs"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
)

// createDepositPayload returns a minimal valid deposit request.
func createDepositPayload(userID, piID, walletID string) models.CreateDepositRequest {
	return models.CreateDepositRequest{
		Initiator: models.TransactionInitiator{ID: userID, Type: "PERSON"},
		SourceMethod: models.TransactionMethod{
			Currency:          "USD",
			PaymentMethodType: "ACH",
			PaymentInformation: models.TransactionPaymentInformation{
				Type: "BANK_ACCOUNT",
				ID:   piID,
			},
		},
		DestinationMethod: models.TransactionMethod{
			PaymentMethodType: "WALLET",
			PaymentInformation: models.TransactionPaymentInformation{
				Type: "WALLET",
				ID:   walletID,
			},
		},
		Amount:    100.00,
		USDAmount: 100.00,
		Type:      "DEPOSIT",
	}
}

// createWithdrawalPayload returns a minimal valid withdrawal request.
func createWithdrawalPayload(userID, walletID, piID string) models.CreateWithdrawalRequest {
	return models.CreateWithdrawalRequest{
		Initiator: models.TransactionInitiator{ID: userID, Type: "PERSON"},
		SourceMethod: models.TransactionMethod{
			PaymentMethodType: "WALLET",
			PaymentInformation: models.TransactionPaymentInformation{
				Type: "WALLET",
				ID:   walletID,
			},
		},
		DestinationMethod: models.TransactionMethod{
			Currency:          "USD",
			PaymentMethodType: "ACH",
			PaymentInformation: models.TransactionPaymentInformation{
				Type: "BANK_ACCOUNT",
				ID:   piID,
			},
		},
		Amount:    50.00,
		USDAmount: 50.00,
		Type:      "WITHDRAWAL",
	}
}

// createTransferPayload returns a minimal valid transfer request.
func createTransferPayload(fromUserID, fromWalletID, toUserID, toWalletID string) models.CreateTransferRequest {
	return models.CreateTransferRequest{
		Initiator: models.TransactionInitiator{ID: fromUserID, Type: "PERSON"},
		SourceTransferMethod: models.TransactionMethod{
			PaymentMethodType: "WALLET",
			PaymentInformation: models.TransactionPaymentInformation{
				Type: "WALLET",
				ID:   fromWalletID,
			},
		},
		DestinationTransferMethod: models.TransactionMethod{
			PaymentMethodType: "WALLET",
			PaymentInformation: models.TransactionPaymentInformation{
				Type: "WALLET",
				ID:   toWalletID,
			},
		},
		Amount:   75.00,
		USDValue: 75.00,
		Type:     "TRANSFER",
	}
}

// seedUserAndWallet creates a test user and wallet in the store, returning their IDs.
func seedUserAndWallet(t *testing.T, h *Handler) (userID, walletID string) {
	t.Helper()
	userID = "txuser-" + t.Name()
	walletID = "txwallet-" + t.Name()

	user := &models.User{ID: userID, Type: "PERSON"}
	if err := h.store.SaveUser(context.Background(), user); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	wallet := &models.Wallet{WalletID: walletID, Currency: "USD", UserID: userID}
	if err := h.store.SaveWallet(context.Background(), wallet); err != nil {
		t.Fatalf("failed to seed wallet: %v", err)
	}
	return
}

// ---- Deposit ----

func TestCreateDeposit_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	userID, walletID := seedUserAndWallet(t, h)
	body := createDepositPayload(userID, "pi-1", walletID)
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/deposits", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID == "" {
		t.Error("expected non-empty transaction request id")
	}
}

func TestCreateDeposit_DefaultsToSettled(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createDepositPayload("u1", "pi-1", "w1")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/deposits", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	// Look up the stored transaction and verify its status.
	tx, err := h.store.GetTransaction(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}
	if tx.Status != "SETTLED" {
		t.Errorf("expected status SETTLED, got %s", tx.Status)
	}
	if tx.TransactionType != "DEPOSIT" {
		t.Errorf("expected type DEPOSIT, got %s", tx.TransactionType)
	}
}

func TestCreateDeposit_FailureScenario(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createDepositPayload("u1", "pi-1", "w1")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/deposits", bytes.NewReader(payload))
	ptiHeaders(req)
	req.Header.Set("x-pti-scenario-id", "REFUSE_DEPOSIT")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	tx, err := h.store.GetTransaction(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}
	if tx.Status != "REFUSED" {
		t.Errorf("expected status REFUSED, got %s", tx.Status)
	}
}

func TestCreateDeposit_ErrorScenario(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createDepositPayload("u1", "pi-1", "w1")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/deposits", bytes.NewReader(payload))
	ptiHeaders(req)
	req.Header.Set("x-pti-scenario-id", "TRIGGER_ERROR")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	tx, err := h.store.GetTransaction(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}
	if tx.Status != "ERROR" {
		t.Errorf("expected status ERROR, got %s", tx.Status)
	}
}

func TestCreateDeposit_CancelScenario(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createDepositPayload("u1", "pi-1", "w1")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/deposits", bytes.NewReader(payload))
	ptiHeaders(req)
	req.Header.Set("x-pti-scenario-id", "CANCEL_DEPOSIT")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	tx, err := h.store.GetTransaction(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}
	if tx.Status != "CANCELED" {
		t.Errorf("expected status CANCELED, got %s", tx.Status)
	}
}

func TestCreateDeposit_UsesRequestIDHeader(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createDepositPayload("u1", "pi-1", "w1")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/deposits", bytes.NewReader(payload))
	ptiHeaders(req)
	req.Header.Set("x-pti-request-id", "req-fixed-id")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID != "req-fixed-id" {
		t.Errorf("expected id req-fixed-id, got %s", resp.ID)
	}
}

func TestCreateDeposit_InvalidBody(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	req := httptest.NewRequest(http.MethodPost, "/transactions/deposits", bytes.NewReader([]byte("not-json")))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// ---- Withdrawal ----

func TestCreateWithdrawal_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	userID, walletID := seedUserAndWallet(t, h)
	body := createWithdrawalPayload(userID, walletID, "pi-2")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/withdrawals", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID == "" {
		t.Error("expected non-empty transaction request id")
	}
}

func TestCreateWithdrawal_DefaultsToSettled(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createWithdrawalPayload("u1", "w1", "pi-2")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/withdrawals", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	tx, err := h.store.GetTransaction(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}
	if tx.Status != "SETTLED" {
		t.Errorf("expected status SETTLED, got %s", tx.Status)
	}
	if tx.TransactionType != "WITHDRAWAL" {
		t.Errorf("expected type WITHDRAWAL, got %s", tx.TransactionType)
	}
}

func TestCreateWithdrawal_FailureScenario(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createWithdrawalPayload("u1", "w1", "pi-2")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/withdrawals", bytes.NewReader(payload))
	ptiHeaders(req)
	req.Header.Set("x-pti-scenario-id", "FAIL_WITHDRAWAL")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	tx, err := h.store.GetTransaction(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}
	if tx.Status != "REFUSED" {
		t.Errorf("expected status REFUSED, got %s", tx.Status)
	}
}

func TestCreateWithdrawal_InvalidBody(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	req := httptest.NewRequest(http.MethodPost, "/transactions/withdrawals", bytes.NewReader([]byte("{bad")))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// ---- Transfer ----

func TestCreateTransfer_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createTransferPayload("user-a", "wallet-a", "user-b", "wallet-b")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/transfers", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID == "" {
		t.Error("expected non-empty transaction request id")
	}
}

func TestCreateTransfer_DefaultsToSettled(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createTransferPayload("user-a", "wallet-a", "user-b", "wallet-b")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/transfers", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	tx, err := h.store.GetTransaction(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}
	if tx.Status != "SETTLED" {
		t.Errorf("expected status SETTLED, got %s", tx.Status)
	}
	if tx.TransactionType != "TRANSFER" {
		t.Errorf("expected type TRANSFER, got %s", tx.TransactionType)
	}
}

func TestCreateTransfer_FailureScenario(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := createTransferPayload("user-a", "wallet-a", "user-b", "wallet-b")
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/transfers", bytes.NewReader(payload))
	ptiHeaders(req)
	req.Header.Set("x-pti-scenario-id", "REFUSE_TRANSFER")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	tx, err := h.store.GetTransaction(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("failed to get transaction: %v", err)
	}
	if tx.Status != "REFUSED" {
		t.Errorf("expected status REFUSED, got %s", tx.Status)
	}
}

func TestCreateTransfer_InvalidBody(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	req := httptest.NewRequest(http.MethodPost, "/transactions/transfers", bytes.NewReader([]byte("{bad")))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// ---- GetTransaction ----

func TestGetTransaction_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	// Seed a transaction directly.
	tx := &models.Transaction{
		RequestID:       "req-get-test",
		Status:          "SETTLED",
		TransactionType: "DEPOSIT",
		Amount:          200.00,
		Currency:        "USD",
	}
	if err := h.store.SaveTransaction(context.Background(), tx); err != nil {
		t.Fatalf("failed to seed transaction: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/transactions/req-get-test", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.Transaction
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.RequestID != "req-get-test" {
		t.Errorf("expected requestId req-get-test, got %s", resp.RequestID)
	}
	if resp.Status != "SETTLED" {
		t.Errorf("expected status SETTLED, got %s", resp.Status)
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	req := httptest.NewRequest(http.MethodGet, "/transactions/nonexistent", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ---- UpdateTransaction ----

func TestUpdateTransaction_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	// Seed a transaction.
	tx := &models.Transaction{
		RequestID:       "req-update-test",
		Status:          "SETTLED",
		TransactionType: "DEPOSIT",
	}
	if err := h.store.SaveTransaction(context.Background(), tx); err != nil {
		t.Fatalf("failed to seed transaction: %v", err)
	}

	body := models.UpdateTransactionRequest{
		TransactionID: "req-update-test",
		Feedback:      "SETTLED",
		ProviderName:  "test-provider",
		Payload:       `{"status":"SETTLED"}`,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/req-update-test/updates", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID == "" {
		t.Error("expected non-empty update id")
	}
}

func TestUpdateTransaction_NotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	body := models.UpdateTransactionRequest{Feedback: "SETTLED"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/transactions/nonexistent/updates", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateTransaction_InvalidBody(t *testing.T) {
	h := newTestHandler()
	router := newTestRouterFull(h)

	// Seed transaction first.
	tx := &models.Transaction{RequestID: "req-invalid-body", Status: "SETTLED"}
	_ = h.store.SaveTransaction(context.Background(), tx)

	req := httptest.NewRequest(http.MethodPost, "/transactions/req-invalid-body/updates", bytes.NewReader([]byte("{bad")))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// ---- transactionStatus helper ----

func TestTransactionStatus_DefaultSettled(t *testing.T) {
	if got := transactionStatus(""); got != "SETTLED" {
		t.Errorf("expected SETTLED for empty scenario, got %s", got)
	}
	if got := transactionStatus("some-random-id"); got != "SETTLED" {
		t.Errorf("expected SETTLED for random scenario, got %s", got)
	}
}

func TestTransactionStatus_Refused(t *testing.T) {
	for _, scenario := range []string{"REFUSE_DEPOSIT", "FAIL_DEPOSIT", "refuse_transfer", "fail-tx"} {
		if got := transactionStatus(scenario); got != "REFUSED" {
			t.Errorf("expected REFUSED for scenario %q, got %s", scenario, got)
		}
	}
}

func TestTransactionStatus_Error(t *testing.T) {
	for _, scenario := range []string{"TRIGGER_ERROR", "error_case"} {
		if got := transactionStatus(scenario); got != "ERROR" {
			t.Errorf("expected ERROR for scenario %q, got %s", scenario, got)
		}
	}
}

func TestTransactionStatus_Canceled(t *testing.T) {
	for _, scenario := range []string{"CANCEL_DEPOSIT", "cancel_flow"} {
		if got := transactionStatus(scenario); got != "CANCELED" {
			t.Errorf("expected CANCELED for scenario %q, got %s", scenario, got)
		}
	}
}

func TestCreateDeposit_EnqueuesWebhookJob(t *testing.T) {
	store := storage.NewMemoryStorage()
	h := NewHandler(store, newTestHandler().config)
	q := jobs.NewQueue(store)
	h.SetQueue(q)
	router := newTestRouter(h)

	userID, walletID := seedUserAndWallet(t, h)
	piID := "pi-1"
	payload := createDepositPayload(userID, piID, walletID)
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/transactions/deposits", bytes.NewReader(body))
	ptiHeaders(req)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	readyJobs, err := store.ListReadyJobs(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListReadyJobs failed: %v", err)
	}
	if len(readyJobs) == 0 {
		t.Fatal("expected queued webhook job")
	}
	if readyJobs[0].JobType != jobs.JobTypeTransactionStatusWebhook {
		t.Fatalf("expected job type %s, got %s", jobs.JobTypeTransactionStatusWebhook, readyJobs[0].JobType)
	}
}

// ---- ErrTransactionNotFound sentinel ----

func TestStorageErrTransactionNotFound(t *testing.T) {
	store := storage.NewMemoryStorage()
	_, err := store.GetTransaction(context.Background(), "nope")
	if !errors.Is(err, storage.ErrTransactionNotFound) {
		t.Errorf("expected ErrTransactionNotFound, got %v", err)
	}
}

// ---- Reset handler ----

func TestReset_Success(t *testing.T) {
	h := newTestHandler()

	// Seed some data that will be cleared.
	_ = h.store.SaveUser(context.Background(), &models.User{ID: "u-reset", Type: "PERSON"})
	_ = h.store.SaveTransaction(context.Background(), &models.Transaction{RequestID: "req-reset", Status: "SETTLED"})

	r := newTestRouterFull(h)

	req := httptest.NewRequest(http.MethodPost, "/test/reset", nil)
	rr := httptest.NewRecorder()
	r.Post("/test/reset", h.Reset)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify the store was cleared.
	if _, err := h.store.GetTransaction(context.Background(), "req-reset"); !errors.Is(err, storage.ErrTransactionNotFound) {
		t.Error("expected transaction to be cleared after reset")
	}
}
