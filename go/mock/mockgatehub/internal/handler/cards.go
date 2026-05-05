package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/logger"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// generateMaskedPan generates a realistic masked PAN with numeric-only digits
func generateMaskedPan() string {
	return fmt.Sprintf("%s******%s", randomDigits(6), randomDigits(4))
}

// randomDigits returns a string of n cryptographically random decimal digits
func randomDigits(n int) string {
	digits := make([]byte, n)
	for i := range digits {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		digits[i] = '0' + byte(num.Int64())
	}
	return string(digits)
}

// CreateCustomer creates a card customer (generic endpoint)
func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	h.CreateManagedCustomer(w, r)
}

// CreateManagedCustomer creates a card customer with account and initial card
func (h *Handler) CreateManagedCustomer(w http.ResponseWriter, r *http.Request) {
	logger.Info("create managed customer called")

	userID := r.Header.Get("x-gatehub-managed-user-uuid")
	if userID == "" {
		h.sendError(w, http.StatusBadRequest, "missing x-gatehub-managed-user-uuid header")
		return
	}

	// Validate user exists and KYC is accepted
	user, err := h.store.GetUser(userID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "user not found")
		return
	}
	if user.KYCState != consts.KYCStateAccepted {
		h.sendError(w, http.StatusForbidden, "user KYC state must be accepted")
		return
	}

	var req models.CreateCustomerAndCardArgs
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate nameOnCard
	if req.NameOnCard == "" {
		h.sendError(w, http.StatusBadRequest, "nameOnCard is required")
		return
	}
	if len(req.NameOnCard) > 26 {
		h.sendError(w, http.StatusBadRequest, "nameOnCard must be 26 characters or less")
		return
	}

	// Validate currency (real GateHub uses product-prefixed codes like "PW_EUR")
	currency := strings.TrimPrefix(req.Account.Currency, "PW_")
	if currency == "" {
		currency = "EUR"
	}
	if currency != "EUR" {
		h.sendError(w, http.StatusBadRequest, "only EUR currency is supported for cards")
		return
	}

	productCode := req.Account.ProductCode
	if productCode == "" {
		productCode = "PWSR_DEBP_2404"
	}

	// Create customer
	customerID := utils.GenerateUUID()
	customer := &models.Customer{
		ID:        &customerID,
		SourceID:  userID,
		Type:      "Citizen",
		Code:      fmt.Sprintf("CUST-%s", customerID[:8]),
		KYCStatus: consts.KYCStateAccepted,
		CreatedAt: time.Now(),
	}

	if err := h.store.CreateCustomer(customer); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create customer")
		return
	}

	// Create delivery address if provided
	if req.Delivery != nil {
		addr := &models.CustomerDeliveryAddress{
			ID:               utils.GenerateUUID(),
			SourceID:         utils.GenerateUUID(),
			CustomerID:       customerID,
			CustomerSourceID: userID,
			Type:             req.Delivery.Type,
			Line1:            req.Delivery.Line1,
			Line2:            req.Delivery.Line2,
			Line3:            req.Delivery.Line3,
			City:             req.Delivery.City,
			PostOffice:       req.Delivery.PostOffice,
			ZipCode:          req.Delivery.ZipCode,
			CountryCode:      req.Delivery.CountryCode,
			Status:           "ACTIVE",
		}
		if err := h.store.CreateCustomerAddress(customerID, addr); err != nil {
			logger.Warn("failed to create delivery address", zap.Error(err))
		}
	}

	// Create account
	accountID := utils.GenerateUUID()
	account := &models.Account{
		ID:               &accountID,
		SourceID:         accountID,
		CustomerID:       &customerID,
		CustomerSourceID: userID,
		ProductCode:      productCode,
		Currency:         currency,
		AccountNumber:    fmt.Sprintf("GB29NWBK%s", utils.GenerateUUID()[:12]),
		Type:             "DEBIT",
		Status:           "ACTIVE",
		CreatedAt:        time.Now(),
	}

	if err := h.store.CreateAccount(account); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	// Create card
	cardID := utils.GenerateUUID()
	card := &models.Card{
		ID:               cardID,
		SourceID:         cardID,
		AccountID:        accountID,
		AccountSourceID:  accountID,
		CustomerID:       customerID,
		CustomerSourceID: userID,
		NameOnCard:       req.NameOnCard,
		ProductCode:      productCode,
		PanToken:         fmt.Sprintf("pan_%s", cardID),
		MaskedPan:        generateMaskedPan(),
		Status:           consts.CardStatusActive,
		RelationType:     "PRIMARY",
		IsFirstTimeLock:  false,
		PlasticCreated:   false,
		CreatedAt:        time.Now(),
	}

	if err := h.store.CreateCard(card); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create card")
		return
	}

	// Seed default card limits
	defaultLimits := []models.CardLimit{
		{Type: "dailyOverall", Limit: 1000.00, Currency: "EUR", IsDisabled: false},
		{Type: "perTransaction", Limit: 500.00, Currency: "EUR", IsDisabled: false},
		{Type: "monthlyOverall", Limit: 5000.00, Currency: "EUR", IsDisabled: false},
		{Type: "dailyAtm", Limit: 300.00, Currency: "EUR", IsDisabled: false},
		{Type: "dailyEcomm", Limit: 800.00, Currency: "EUR", IsDisabled: false},
	}
	if err := h.store.SetCardLimits(cardID, defaultLimits); err != nil {
		logger.Warn("failed to seed card limits", zap.Error(err))
	}

	// Build response matching the documented CustomerResponse shape
	account.Cards = []models.Card{*card}
	customer.Accounts = []models.Account{*account}

	response := models.CustomerResponse{
		WalletAddress: req.WalletAddress,
		Customer:      *customer,
	}

	h.sendJSON(w, http.StatusCreated, response)

	// Fire cards.card.created webhook so the backend can store the customer/account IDs
	go h.webhookManager.SendAsync(consts.WebhookEventCardCreated, userID, map[string]interface{}{
		"cardId":           cardID,
		"cardSourceId":     cardID,
		"nameOnCard":       req.NameOnCard,
		"productCode":      productCode,
		"maskedPan":        card.MaskedPan,
		"accountId":        accountID,
		"accountSourceId":  accountID,
		"lockLevel":        nil,
		"customerId":       customerID,
		"customerSourceId": userID,
	}, 0)
}

// ListCards retrieves cards for a customer
func (h *Handler) ListCards(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerID")
	logger.Info("list cards called", zap.String("customer_id", customerID))

	cards, err := h.store.GetCardsByCustomer(customerID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get cards")
		return
	}

	// Filter out SoftDelete cards
	var activeCards []models.Card
	for _, c := range cards {
		if c.Status != consts.CardStatusSoftDelete {
			activeCards = append(activeCards, *c)
		}
	}
	if activeCards == nil {
		activeCards = []models.Card{}
	}

	response := models.ListCardsResponse{
		Data: activeCards,
		Pagination: models.Pagination{
			PageNumber: 1,
			PageSize:   100,
			TotalPages: 1,
		},
	}

	h.sendJSON(w, http.StatusOK, response)
}

// CreateCard creates a new card (stub)
func (h *Handler) CreateCard(w http.ResponseWriter, r *http.Request) {
	logger.Info("create card called")
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"id":     utils.GenerateUUID(),
		"status": consts.CardStatusActive,
	})
}

// GetCard retrieves card details
func (h *Handler) GetCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	logger.Info("get card called", zap.String("card_id", cardID))

	card, err := h.store.GetCard(cardID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "card not found")
		return
	}

	h.sendJSON(w, http.StatusOK, card)
}

// DeleteCard soft-deletes a card
func (h *Handler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	logger.Info("delete card called", zap.String("card_id", cardID))

	card, err := h.store.GetCard(cardID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "card not found")
		return
	}

	if card.Status == consts.CardStatusSoftDelete {
		h.sendError(w, http.StatusBadRequest, "card is already deleted")
		return
	}

	card.Status = consts.CardStatusSoftDelete
	card.StatusReasonCode = nil
	card.LockLevel = nil

	if err := h.store.UpdateCard(card); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to delete card")
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// LockCard locks a card temporarily
func (h *Handler) LockCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	reasonCode := r.URL.Query().Get("reasonCode")
	logger.Info("lock card called", zap.String("card_id", cardID), zap.String("reason_code", reasonCode))

	if reasonCode == "" {
		reasonCode = "ClientRequestedLock"
	}

	card, err := h.store.GetCard(cardID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "card not found")
		return
	}

	// Validate card can be locked
	if card.Status == consts.CardStatusSoftDelete {
		h.sendError(w, http.StatusBadRequest, "cannot lock a deleted card")
		return
	}
	if card.Status == consts.CardStatusBlocked {
		h.sendError(w, http.StatusBadRequest, "cannot lock a permanently blocked card")
		return
	}
	if card.Status == consts.CardStatusTemporaryBlocked {
		h.sendError(w, http.StatusBadRequest, "card is already locked")
		return
	}

	card.Status = consts.CardStatusTemporaryBlocked
	card.LockLevel = &reasonCode
	if !card.IsFirstTimeLock {
		card.IsFirstTimeLock = true
	}

	if err := h.store.UpdateCard(card); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to lock card")
		return
	}

	h.sendJSON(w, http.StatusOK, card)
}

// UnlockCard unlocks a temporarily blocked card
func (h *Handler) UnlockCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	logger.Info("unlock card called", zap.String("card_id", cardID))

	card, err := h.store.GetCard(cardID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "card not found")
		return
	}

	if card.Status != consts.CardStatusTemporaryBlocked {
		h.sendError(w, http.StatusBadRequest, "card is not locked")
		return
	}

	card.Status = consts.CardStatusActive
	card.LockLevel = nil

	if err := h.store.UpdateCard(card); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to unlock card")
		return
	}

	h.sendJSON(w, http.StatusOK, card)
}

// BlockCard permanently blocks a card
func (h *Handler) BlockCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	reasonCode := r.URL.Query().Get("reasonCode")
	logger.Info("block card called", zap.String("card_id", cardID))

	card, err := h.store.GetCard(cardID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "card not found")
		return
	}

	if card.Status == consts.CardStatusSoftDelete {
		h.sendError(w, http.StatusBadRequest, "cannot block a deleted card")
		return
	}
	if card.Status == consts.CardStatusBlocked {
		h.sendError(w, http.StatusBadRequest, "card is already blocked")
		return
	}

	card.Status = consts.CardStatusBlocked
	card.StatusReasonCode = &reasonCode
	card.LockLevel = nil

	if err := h.store.UpdateCard(card); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to block card")
		return
	}

	h.sendJSON(w, http.StatusOK, card)
}

// GetPendingConfirmations retrieves pending 3DS confirmations
func (h *Handler) GetPendingConfirmations(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("x-gatehub-managed-user-uuid")
	logger.Info("get pending confirmations called", zap.String("user_id", userID))

	challenges, err := h.store.GetPendingThreeDSChallenges(userID)
	if err != nil {
		h.sendJSON(w, http.StatusOK, map[string][]models.PendingThreeDSConfirmation{
			"pendingConfirmations": nil,
		})
		return
	}

	var pending []models.PendingThreeDSConfirmation
	for _, c := range challenges {
		pending = append(pending, models.PendingThreeDSConfirmation{
			TransactionID:    c.TransactionID,
			MerchantName:     c.MerchantName,
			PurchaseAmount:   c.PurchaseAmount,
			PurchaseCurrency: c.PurchaseCurrency,
			PurchaseDate:     c.PurchaseDate,
			Timeout:          c.Timeout.Format(time.RFC3339),
		})
	}

	if pending == nil {
		pending = []models.PendingThreeDSConfirmation{}
	}

	h.sendJSON(w, http.StatusOK, map[string][]models.PendingThreeDSConfirmation{
		"pendingConfirmations": pending,
	})
}

// CreateCustomerAddress creates a delivery address for a card customer
func (h *Handler) CreateCustomerAddress(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerID")
	logger.Info("creating customer address", zap.String("customer_id", customerID))

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	address := &models.CustomerDeliveryAddress{
		ID:         utils.GenerateUUID(),
		CustomerID: customerID,
		Type:       fmt.Sprintf("%v", req["type"]),
		Line1:      fmt.Sprintf("%v", req["line1"]),
		City:       fmt.Sprintf("%v", req["city"]),
		ZipCode:    fmt.Sprintf("%v", req["zipCode"]),
		Status:     "ACTIVE",
	}
	if cc, ok := req["countryCode"].(string); ok {
		address.CountryCode = cc
	}

	if err := h.store.CreateCustomerAddress(customerID, address); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create address")
		return
	}

	h.sendJSON(w, http.StatusCreated, []models.CustomerDeliveryAddress{*address})
}

// GetCustomerAddresses retrieves delivery addresses for a card customer
func (h *Handler) GetCustomerAddresses(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerID")
	logger.Info("getting customer addresses", zap.String("customer_id", customerID))

	addresses, err := h.store.GetCustomerAddresses(customerID)
	if err != nil {
		h.sendJSON(w, http.StatusOK, []models.CustomerDeliveryAddress{})
		return
	}

	h.sendJSON(w, http.StatusOK, addresses)
}

// OrderAdditionalCard orders an additional card for an account
func (h *Handler) OrderAdditionalCard(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")
	logger.Info("ordering additional card", zap.String("account_id", accountID))

	card := map[string]interface{}{
		"id":         utils.GenerateUUID(),
		"accountID":  accountID,
		"status":     "PENDING",
		"cardNumber": "****1234",
		"createdAt":  time.Now().UTC().Format(time.RFC3339),
	}

	h.sendJSON(w, http.StatusCreated, card)
}

// GetCardLimits retrieves spending limits for a card
func (h *Handler) GetCardLimits(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	logger.Info("get card limits called", zap.String("card_id", cardID))

	limits, err := h.store.GetCardLimits(cardID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get limits")
		return
	}

	// Seed default limits if empty
	if len(limits) == 0 {
		limits = []models.CardLimit{
			{Type: "dailyOverall", Limit: 1000.00, Currency: "EUR", IsDisabled: false},
			{Type: "perTransaction", Limit: 500.00, Currency: "EUR", IsDisabled: false},
			{Type: "monthlyOverall", Limit: 5000.00, Currency: "EUR", IsDisabled: false},
			{Type: "dailyAtm", Limit: 300.00, Currency: "EUR", IsDisabled: false},
			{Type: "dailyEcomm", Limit: 800.00, Currency: "EUR", IsDisabled: false},
		}
		_ = h.store.SetCardLimits(cardID, limits)
	}

	h.sendJSON(w, http.StatusOK, limits)
}

// UpdateCardLimits updates spending limits for a card
func (h *Handler) UpdateCardLimits(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	logger.Info("update card limits called", zap.String("card_id", cardID))

	var limits []models.CardLimit
	if err := json.NewDecoder(r.Body).Decode(&limits); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.SetCardLimits(cardID, limits); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to update limits")
		return
	}

	h.sendJSON(w, http.StatusOK, limits)
}

// GetCardToken generates a token for retrieving card data.
//
// In the real GateHub API the response includes an absolute URL in
// `links[0].href` that the browser fetches directly with the token as a
// `Bearer` Authorization header. The mock therefore must:
//   - generate a real JWT carrying the requested cardId and the caller's
//     RSA public key (so the data endpoint can encrypt the response with
//     a key only the browser can decrypt), and
//   - return an absolute URL pointing back at this mockgatehub instance.
func (h *Handler) GetCardToken(w http.ResponseWriter, r *http.Request) {
	logger.Info("get card token called")

	var req models.GetCardTokenArgs
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CardID == "" {
		h.sendError(w, http.StatusBadRequest, "cardId is required")
		return
	}

	// publicKey is required: the data endpoint encrypts its response with
	// this key, so a token issued without it would be unusable by callers.
	// Reject up front to keep POST /token/card-data and GET /token/card-data/data
	// behaviourally coherent.
	if req.PublicKey == nil || strings.TrimSpace(*req.PublicKey) == "" {
		h.sendError(w, http.StatusBadRequest, "publicKey is required")
		return
	}
	publicKey := strings.TrimSpace(*req.PublicKey)

	now := time.Now()
	claims := CardDataClaims{
		CardID:    req.CardID,
		PublicKey: publicKey,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(cardDataTokenTTL).Unix(),
	}

	jwt, err := generateCardDataJWT(h.config.CardDataTokenSecret, claims)
	if err != nil {
		logger.Error("failed to sign card data token", zap.Error(err))
		h.sendError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	href := h.config.PublicBaseURL + cardDataPath

	response := models.CardTokenResponse{
		Token: jwt,
		Links: []models.CardTokenLink{
			{
				Href:   href,
				Rel:    "data",
				Method: "GET",
			},
		},
	}

	h.sendJSON(w, http.StatusOK, response)
}

// GetCardData returns mock sensitive card data, encrypted with the RSA public
// key the caller supplied when creating the card-data token. It is invoked
// directly by the browser using the token as a `Bearer` Authorization header
// and therefore is excluded from HMAC authentication.
func (h *Handler) GetCardData(w http.ResponseWriter, r *http.Request) {
	logger.Info("get card data called")

	authz := r.Header.Get("Authorization")
	if !strings.HasPrefix(authz, "Bearer ") {
		h.sendError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}
	token := strings.TrimPrefix(authz, "Bearer ")

	claims, err := parseCardDataJWT(h.config.CardDataTokenSecret, token)
	if err != nil {
		logger.Warn("invalid card data token", zap.Error(err))
		h.sendError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	if claims.PublicKey == "" {
		h.sendError(w, http.StatusBadRequest, "token has no publicKey claim")
		return
	}

	// Build mock sensitive payload. Field casing matches protea's
	// CardProcessorSensitiveDataResponse type (Pan / ExpiryDate / Cvc2).
	pan := generateUnmaskedPAN(claims.CardID)
	payload := map[string]string{
		"Pan":        pan,
		"ExpiryDate": "12/2030",
		"Cvc2":       randomDigits(3),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to encode card data")
		return
	}

	cypher, err := encryptWithBase64SPKI(claims.PublicKey, payloadBytes)
	if err != nil {
		logger.Warn("card data encryption failed", zap.Error(err))
		h.sendError(w, http.StatusBadRequest, "failed to encrypt card data with provided publicKey")
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]string{"cypher": cypher})
}

// generateUnmaskedPAN returns a 16-digit PAN derived deterministically from
// the cardID by hashing it with SHA-256 and mapping the resulting bytes to
// decimal digits. This keeps repeated calls for the same card stable (so the
// mock behaves like a real provider that always returns the same PAN for a
// given card) without requiring any stored PAN. An empty cardID falls back to
// random digits so the function is still safe for edge cases.
func generateUnmaskedPAN(cardID string) string {
	if cardID == "" {
		return randomDigits(16)
	}
	sum := sha256.Sum256([]byte(cardID))
	digits := make([]byte, 16)
	for i := 0; i < 16; i++ {
		digits[i] = '0' + (sum[i] % 10)
	}
	return string(digits)
}

// CreateCardTransaction creates a card transaction
func (h *Handler) CreateCardTransaction(w http.ResponseWriter, r *http.Request) {
	logger.Info("create card transaction called")

	var req models.CreateCardTransactionArgs
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	txID := fmt.Sprintf("tx-%s", utils.GenerateUUID()[:8])
	now := time.Now().UTC().Format(time.RFC3339)

	tx := &models.CardTransaction{
		TransactionID:         txID,
		GHResponseCode:        "00",
		GHResponseDescription: "Approved",
		TransactionAmount:     &req.Amount,
		TransactionCurrency:   &req.Currency,
		BillingAmount:         &req.Amount,
		BillingCurrency:       &req.Currency,
		Type:                  req.Type,
		CardScheme:            2,
		TerminalID:            "TERM001",
		CreatedAt:             now,
		TxStatus:              strPtr("COMPLETED"),
		Operation:             0,
		MerchantName:          req.MerchantName,
		MerchantCity:          req.MerchantCity,
		MerchantCountry:       req.MerchantCountry,
		TransactionDateTime:   &now,
		ProcessDateTime:       &now,
	}

	if err := h.store.CreateCardTransaction(tx); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create transaction")
		return
	}

	// Index by card
	if req.CardID != "" {
		_ = h.store.AddCardTransactionIndex(req.CardID, txID)
	}

	h.sendJSON(w, http.StatusCreated, tx)
}

// GetCardTransaction retrieves a single card transaction by ID
func (h *Handler) GetCardTransaction(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txID")
	logger.Info("get card transaction called", zap.String("tx_id", txID))

	tx, err := h.store.GetCardTransaction(txID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "transaction not found")
		return
	}

	h.sendJSON(w, http.StatusOK, tx)
}

// ListCardTransactions retrieves card transactions (stub)
func (h *Handler) ListCardTransactions(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	logger.Info("list card transactions called", zap.String("card_id", cardID))

	txIDs, err := h.store.GetCardTransactionIDs(cardID)
	if err != nil {
		h.sendJSON(w, http.StatusOK, models.CardTransactionsResponse{
			Data: []models.CardTransaction{},
			Pagination: models.CardTransactionsPagination{
				PageNumber:   1,
				PageSize:     20,
				TotalPages:   1,
				TotalRecords: 0,
			},
		})
		return
	}

	var txs []models.CardTransaction
	for _, txID := range txIDs {
		tx, err := h.store.GetCardTransaction(txID)
		if err == nil {
			txs = append(txs, *tx)
		}
	}
	if txs == nil {
		txs = []models.CardTransaction{}
	}

	h.sendJSON(w, http.StatusOK, models.CardTransactionsResponse{
		Data: txs,
		Pagination: models.CardTransactionsPagination{
			PageNumber:   1,
			PageSize:     20,
			TotalPages:   1,
			TotalRecords: uint(len(txs)),
		},
	})
}

// CreateThreeDSChallenge creates a 3DS challenge for testing
func (h *Handler) CreateThreeDSChallenge(w http.ResponseWriter, r *http.Request) {
	logger.Info("create 3ds challenge called")

	var req struct {
		CardID           string `json:"cardId"`
		UserID           string `json:"userId"`
		MerchantName     string `json:"merchantName"`
		PurchaseAmount   string `json:"purchaseAmount"`
		PurchaseCurrency string `json:"purchaseCurrency"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	txID := fmt.Sprintf("3ds-%s", utils.GenerateUUID()[:8])

	challenge := &models.ThreeDSChallenge{
		TransactionID:    txID,
		CardID:           req.CardID,
		UserID:           req.UserID,
		MerchantName:     req.MerchantName,
		PurchaseAmount:   req.PurchaseAmount,
		PurchaseCurrency: req.PurchaseCurrency,
		PurchaseDate:     time.Now().UTC().Format(time.RFC3339),
		Timeout:          time.Now().Add(5 * time.Minute),
		Status:           "pending",
		CreatedAt:        time.Now(),
	}

	if err := h.store.CreateThreeDSChallenge(challenge); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to create 3DS challenge")
		return
	}

	h.sendJSON(w, http.StatusCreated, challenge)
}

// ConfirmThreeDS confirms or denies a 3DS challenge
func (h *Handler) ConfirmThreeDS(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txID")
	logger.Info("confirm 3ds called", zap.String("tx_id", txID))

	var req models.ThreeDSPaymentConfirmationArgs
	if err := h.decodeJSON(r, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	challenge, err := h.store.GetThreeDSChallenge(txID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "3DS challenge not found")
		return
	}

	if req.Confirmed {
		challenge.Status = "approved"
	} else {
		challenge.Status = "declined"
	}

	if err := h.store.UpdateThreeDSChallenge(challenge); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to update 3DS challenge")
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"transactionId": txID,
		"status":        challenge.Status,
	})
}

// GetCardApplicationProducts retrieves available card products
func (h *Handler) GetCardApplicationProducts(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appID")
	logger.Info("get card application products called", zap.String("app_id", appID))

	products := []map[string]interface{}{
		{
			"uuid":               utils.GenerateUUID(),
			"accountProductCode": "PROD_EUR_ACCOUNT",
			"code":               "PWSR_DEBP_2404",
			"name":               "Virtual Debit Card",
			"cost":               "0.00",
			"cardProductLimits": []map[string]interface{}{
				{
					"type":       "dailyOverall",
					"currency":   "EUR",
					"limit":      "1000.00",
					"isDisabled": false,
				},
			},
		},
		{
			"uuid":               utils.GenerateUUID(),
			"accountProductCode": "PROD_EUR_ACCOUNT",
			"code":               "PROD_PLASTIC_CARD",
			"name":               "Physical Debit Card",
			"cost":               "5.00",
			"cardProductLimits": []map[string]interface{}{
				{
					"type":       "dailyOverall",
					"currency":   "EUR",
					"limit":      "1000.00",
					"isDisabled": false,
				},
			},
		},
	}

	h.sendJSON(w, http.StatusOK, products)
}

// OrderPlasticCard orders a physical card
func (h *Handler) OrderPlasticCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	logger.Info("order plastic card called", zap.String("card_id", cardID))

	h.sendJSON(w, http.StatusCreated, map[string]interface{}{
		"orderId":   utils.GenerateUUID(),
		"cardId":    cardID,
		"status":    "PENDING",
		"type":      "PLASTIC",
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	})
}

// strPtr returns a pointer to a string
func strPtr(s string) *string {
	return &s
}
