package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/handler"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/logger"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestServer wraps the HTTP server for integration testing
type TestServer struct {
	Router  *chi.Mux
	Store   storage.Storage
	Handler *handler.Handler
}

// NewTestServer creates a test server with in-memory storage
func NewTestServer() *TestServer {
	logger.Info("creating test server")

	store := storage.NewMemoryStorage()
	if err := storage.SeedTestUsers(store); err != nil {
		panic(fmt.Sprintf("Failed to seed test users: %v", err))
	}

	webhookManager := webhook.NewManager("", "test-secret", nil, nil, "")
	h := handler.NewHandler(store, webhookManager)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Setup routes (same as main.go)
	r.Get("/health", h.HealthCheck)
	r.Route("/auth/v1", func(r chi.Router) {
		r.Post("/tokens", h.CreateToken)
		r.Post("/users/managed", h.CreateManagedUser)
		r.Get("/users/managed", h.GetManagedUser)
		r.Put("/users/managed/email", h.UpdateManagedUserEmail)
	})
	r.Route("/id/v1", func(r chi.Router) {
		r.Get("/users/{userID}", h.GetUser)
		r.Post("/users/{userID}/hubs/{gatewayID}", h.StartKYC)
		r.Put("/hubs/{gatewayID}/users/{userID}", h.UpdateKYCState)
	})
	r.Get("/iframe/onboarding", h.KYCIframe)
	r.Post("/iframe/submit", h.KYCIframeSubmit)
	r.Route("/core/v1", func(r chi.Router) {
		r.Post("/wallets", h.CreateWallet)
		r.Get("/wallets/{address}", h.GetWallet)
		r.Get("/wallets/{address}/balance", h.GetWalletBalance)
		r.Post("/transactions", h.CreateTransaction)
		r.Get("/transactions/{txID}", h.GetTransaction)
	})
	r.Route("/rates/v1", func(r chi.Router) {
		r.Get("/rates/current", h.GetCurrentRates)
		r.Get("/liquidity_provider/vaults", h.GetVaults)
	})
	r.Route("/cards/v1", func(r chi.Router) {
		r.Post("/customers/managed", h.CreateManagedCustomer)
		r.Post("/cards", h.CreateCard)
		r.Get("/cards/{cardID}", h.GetCard)
		r.Delete("/cards/{cardID}", h.DeleteCard)
	})

	return &TestServer{
		Router:  r,
		Store:   store,
		Handler: h,
	}
}

// MakeRequest makes an HTTP request to the test server
func (ts *TestServer) MakeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyBytes)
		req := httptest.NewRequest(method, path, bodyReader)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		ts.Router.ServeHTTP(rr, req)
		return rr
	}

	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	ts.Router.ServeHTTP(rr, req)
	return rr
}

// Full Workflow Integration Tests

func TestFullUserJourney(t *testing.T) {
	logger.Info("starting full user journey test")
	ts := NewTestServer()

	// 1. Create a new managed user
	logger.Info("step 1: create managed user")
	createUserReq := models.CreateManagedUserRequest{
		Email: "newuser@example.com",
	}
	rr := ts.MakeRequest("POST", "/auth/v1/users/managed", createUserReq)
	require.Equal(t, http.StatusCreated, rr.Code, "Failed to create user: %s", rr.Body.String())

	var createUserResp models.CreateManagedUserResponse
	err := json.NewDecoder(rr.Body).Decode(&createUserResp)
	require.NoError(t, err)
	user := models.User{
		ID:        createUserResp.ID,
		Email:     createUserResp.Email,
		Activated: createUserResp.Activated,
		Managed:   createUserResp.Managed,
		Role:      createUserResp.Role,
		Features:  createUserResp.Features,
		KYCState:  createUserResp.KYCState,
		RiskLevel: createUserResp.RiskLevel,
	}

	// 2. Start KYC process
	logger.Info("step 2: start kyc")
	kycPath := fmt.Sprintf("/id/v1/users/%s/hubs/gateway-1", user.ID)
	rr = ts.MakeRequest("POST", kycPath, nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var kycResponse models.StartKYCResponse
	err = json.NewDecoder(rr.Body).Decode(&kycResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, kycResponse.IframeURL)
	logger.Info("kyc iframe url generated", zap.String("url", kycResponse.IframeURL))

	// 3. Verify user is in action_required state (not auto-approved)
	logger.Info("step 3: verify kyc is pending approval")
	time.Sleep(100 * time.Millisecond) // Let goroutine complete
	userPath := fmt.Sprintf("/id/v1/users/%s", user.ID)
	rr = ts.MakeRequest("GET", userPath, nil)
	require.Equal(t, http.StatusOK, rr.Code)

	err = json.NewDecoder(rr.Body).Decode(&user)
	require.NoError(t, err)
	assert.Equal(t, "action_required", user.KYCState)
	assert.Equal(t, "low", user.RiskLevel)
	logger.Info("kyc status", zap.String("state", user.KYCState), zap.String("risk", user.RiskLevel))

	// 3b. Submit KYC form to approve user
	logger.Info("step 3b: submit kyc form")
	kycSubmitData := fmt.Sprintf("user_id=%s&first_name=John&last_name=Doe&dob=1990-01-01&address=123+Main+St&city=NY&country=USA&risk_level=low", user.ID)
	req := httptest.NewRequest("POST", "/iframe/submit", bytes.NewBufferString(kycSubmitData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	ts.Router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// 3c. Verify user is now accepted
	logger.Info("step 3c: verify user is approved after form submission")
	rr = ts.MakeRequest("GET", userPath, nil)
	require.Equal(t, http.StatusOK, rr.Code)

	err = json.NewDecoder(rr.Body).Decode(&user)
	require.NoError(t, err)
	assert.Equal(t, "accepted", user.KYCState)
	logger.Info("kyc status after submission", zap.String("state", user.KYCState))

	// 4. Create a wallet
	logger.Info("step 4: create wallet")
	createWalletReq := models.CreateWalletRequest{
		UserID: user.ID,
		Name:   "My Test Wallet",
	}
	rr = ts.MakeRequest("POST", "/core/v1/wallets", createWalletReq)
	require.Equal(t, http.StatusCreated, rr.Code, "Failed to create wallet: %s", rr.Body.String())

	var wallet models.Wallet
	err = json.NewDecoder(rr.Body).Decode(&wallet)
	require.NoError(t, err)
	assert.NotEmpty(t, wallet.Address)
	assert.Equal(t, user.ID, wallet.UserID)
	logger.Info("wallet created", zap.String("address", wallet.Address))

	// 5. Deposit funds
	logger.Info("step 5: deposit funds")
	depositReq := models.CreateTransactionRequest{
		UserID:   user.ID,
		Amount:   500.00,
		Currency: "USD",
	}
	rr = ts.MakeRequest("POST", "/core/v1/transactions", depositReq)
	require.Equal(t, http.StatusCreated, rr.Code, "Failed to create transaction: %s", rr.Body.String())

	var tx models.Transaction
	err = json.NewDecoder(rr.Body).Decode(&tx)
	require.NoError(t, err)
	assert.Equal(t, "500.00", tx.Amount)
	assert.Equal(t, "USD", tx.Currency)
	assert.Equal(t, 1, tx.Status)
	logger.Info("deposited funds", zap.String("amount", tx.Amount), zap.String("currency", tx.Currency), zap.String("transaction_id", tx.ID))

	// 6. Check balance (all currencies)
	logger.Info("step 6: check balance")
	balancePath := fmt.Sprintf("/core/v1/wallets/%s/balance", wallet.Address)
	rr = ts.MakeRequest("GET", balancePath, nil)
	require.Equal(t, http.StatusOK, rr.Code)

	var balances []map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&balances)
	require.NoError(t, err)
	require.NotEmpty(t, balances)
	assert.Equal(t, "USD", balances[1]["vault"].(map[string]interface{})["asset_code"])

	// Find USD balance
	var usdBalance float64
	for _, bal := range balances {
		v := bal["vault"].(map[string]interface{})
		if v["asset_code"] == "USD" {
			valStr, _ := bal["available"].(string)
			fmt.Sscan(valStr, &usdBalance)
			assert.NotEmpty(t, v["uuid"])
			logger.Info("usd balance retrieved", zap.String("balance", valStr), zap.String("vault", fmt.Sprintf("%v", v["uuid"])))
		}
	}
	assert.Equal(t, 500.00, usdBalance)

	logger.Info("full user journey completed successfully")
}

func TestKYCIframe(t *testing.T) {
	ts := NewTestServer()

	rr := ts.MakeRequest("GET", "/iframe/onboarding?token=test-token&user_id=test-user", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rr.Body.String(), "KYC Verification")
	assert.Contains(t, rr.Body.String(), "MockGatehub")
}

func TestTransactionAPICompliance(t *testing.T) {
	logger.Info("testing transaction api compliance")
	ts := NewTestServer()

	// Create user and wallet
	user := &models.User{Email: "txtest@example.com"}
	err := ts.Store.CreateUser(user)
	require.NoError(t, err)

	wallet := &models.Wallet{
		Address: "rTestTxAddr123",
		UserID:  user.ID,
		Name:    "Test Wallet",
	}
	err = ts.Store.CreateWallet(wallet)
	require.NoError(t, err)

	// Create transaction via API
	depositReq := models.CreateTransactionRequest{
		UserID:           user.ID,
		Amount:           123.45,
		Currency:         "EUR",
		VaultUUID:        "a09a0a2c-1a3a-44c5-a1b9-603a6eea9341",
		ReceivingAddress: wallet.Address,
		Type:             1,
		DepositType:      "external",
	}

	rr := ts.MakeRequest("POST", "/core/v1/transactions", depositReq)
	require.Equal(t, http.StatusCreated, rr.Code, "Failed to create transaction: %s", rr.Body.String())

	var tx models.Transaction
	err = json.NewDecoder(rr.Body).Decode(&tx)
	require.NoError(t, err)

	// Verify all fields are correctly formatted
	t.Run("Amount fields are strings", func(t *testing.T) {
		assert.Equal(t, "123.45", tx.Amount)
		assert.Equal(t, "123.45", tx.TotalAmount)
		assert.Equal(t, "0.00", tx.Fee)
	})

	t.Run("Status is integer", func(t *testing.T) {
		assert.Equal(t, 1, tx.Status)
	})

	t.Run("Transaction fields are present", func(t *testing.T) {
		assert.NotEmpty(t, tx.ID)
		assert.Equal(t, "EUR", tx.Currency)
		assert.Equal(t, user.ID, tx.UserID)
		assert.Equal(t, wallet.Address, tx.ReceivingAddress)
		assert.NotZero(t, tx.CreatedAt)
	})

	// Retrieve transaction and verify format
	t.Run("GET transaction returns same format", func(t *testing.T) {
		txPath := fmt.Sprintf("/core/v1/transactions/%s", tx.ID)
		rr := ts.MakeRequest("GET", txPath, nil)
		require.Equal(t, http.StatusOK, rr.Code)

		var retrieved models.Transaction
		err := json.NewDecoder(rr.Body).Decode(&retrieved)
		require.NoError(t, err)

		assert.Equal(t, "123.45", retrieved.Amount)
		assert.Equal(t, 1, retrieved.Status)
		assert.Equal(t, tx.ID, retrieved.ID)
	})

	logger.Info("transaction api compliance verified")
}

func TestMultipleCurrencyDeposits(t *testing.T) {
	logger.Info("testing multiple currency deposits")
	ts := NewTestServer()

	// Create user
	user := &models.User{Email: "multicurrency@example.com"}
	err := ts.Store.CreateUser(user)
	require.NoError(t, err)

	// Approve user
	user.KYCState = "accepted"
	err = ts.Store.UpdateUser(user)
	require.NoError(t, err)

	// Create wallet
	wallet := &models.Wallet{
		Address: "rMultiCurrAddr",
		UserID:  user.ID,
		Name:    "Multi-Currency Wallet",
	}
	err = ts.Store.CreateWallet(wallet)
	require.NoError(t, err)

	// Test deposits in different currencies
	currencies := []struct {
		code   string
		vault  string
		amount float64
	}{
		{"USD", "450d2156-132a-4d3f-88c5-74822547658d", 100.00},
		{"EUR", "a09a0a2c-1a3a-44c5-a1b9-603a6eea9341", 200.50},
		{"GBP", "992b932d-7e9e-44b0-90ea-b82a530b6784", 75.25},
	}

	for _, curr := range currencies {
		t.Run(curr.code, func(t *testing.T) {
			depositReq := models.CreateTransactionRequest{
				UserID:           user.ID,
				Amount:           curr.amount,
				Currency:         curr.code,
				VaultUUID:        curr.vault,
				ReceivingAddress: wallet.Address,
				Type:             1,
				DepositType:      "external",
			}

			rr := ts.MakeRequest("POST", "/core/v1/transactions", depositReq)
			require.Equal(t, http.StatusCreated, rr.Code)

			var tx models.Transaction
			err := json.NewDecoder(rr.Body).Decode(&tx)
			require.NoError(t, err)

			expectedAmount := fmt.Sprintf("%.2f", curr.amount)
			assert.Equal(t, expectedAmount, tx.Amount)
			assert.Equal(t, curr.code, tx.Currency)
			assert.Equal(t, 1, tx.Status)

			logger.Info("currency deposit processed", zap.String("currency", curr.code), zap.String("amount", tx.Amount), zap.Int("status", tx.Status))
		})
	}

	// Verify balances
	balancePath := fmt.Sprintf("/core/v1/wallets/%s/balance", wallet.Address)
	rr := ts.MakeRequest("GET", balancePath, nil)
	require.Equal(t, http.StatusOK, rr.Code)

	logger.Info("multi-currency deposits successful")
}
