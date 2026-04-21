package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/webhook"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupCardsHandler creates a handler with a KYC-accepted user1.
func setupCardsHandler(t *testing.T) (*Handler, *storage.MemoryStorage) {
	t.Helper()
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "test-secret", nil, store, "default-org")
	h := NewHandler(store, wm)

	// Accept KYC for user1 so card creation can proceed
	u, _ := store.GetUser(consts.TestUser1ID)
	u.KYCState = consts.KYCStateAccepted
	_ = store.UpdateUser(u)

	return h, store
}

func cardChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func createManagedCustomerReq(nameOnCard string) *http.Request {
	body := models.CreateCustomerAndCardArgs{
		WalletAddress: "rTestWallet",
		Account: models.CardAccount{
			ProductCode: "PWSR_DEBP_2404",
			Currency:    "EUR",
			Card:        models.NewCardArgs{ProductCode: "PWSR_DEBP_2404"},
		},
		NameOnCard: nameOnCard,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/customers", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-gatehub-managed-user-uuid", consts.TestUser1ID)
	return req
}

// --- CreateManagedCustomer ---

func TestCreateManagedCustomer_Success(t *testing.T) {
	h, _ := setupCardsHandler(t)

	w := httptest.NewRecorder()
	h.CreateManagedCustomer(w, createManagedCustomerReq("John Doe"))

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp models.CustomerResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotNil(t, resp.Customer.ID)
	assert.Equal(t, "rTestWallet", resp.WalletAddress)
	assert.Len(t, resp.Customer.Accounts, 1)
	assert.Len(t, resp.Customer.Accounts[0].Cards, 1)
	assert.Equal(t, consts.CardStatusActive, resp.Customer.Accounts[0].Cards[0].Status)
}

func TestCreateManagedCustomer_MissingHeader(t *testing.T) {
	h, _ := setupCardsHandler(t)

	body := models.CreateCustomerAndCardArgs{NameOnCard: "Test"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/customers", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	// no x-gatehub-managed-user-uuid header

	w := httptest.NewRecorder()
	h.CreateManagedCustomer(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateManagedCustomer_KYCNotAccepted(t *testing.T) {
	store := storage.NewMemoryStorage()
	storage.SeedTestUsers(store)
	wm := webhook.NewManager("", "test-secret", nil, store, "default-org")
	h := NewHandler(store, wm)
	// User1 KYC is still action_required by default

	w := httptest.NewRecorder()
	h.CreateManagedCustomer(w, createManagedCustomerReq("Jane Doe"))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateManagedCustomer_MissingNameOnCard(t *testing.T) {
	h, _ := setupCardsHandler(t)

	w := httptest.NewRecorder()
	h.CreateManagedCustomer(w, createManagedCustomerReq(""))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateManagedCustomer_NameOnCardTooLong(t *testing.T) {
	h, _ := setupCardsHandler(t)

	w := httptest.NewRecorder()
	h.CreateManagedCustomer(w, createManagedCustomerReq("ABCDEFGHIJKLMNOPQRSTUVWXYZ+"))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateManagedCustomer_NonEURCurrency(t *testing.T) {
	h, _ := setupCardsHandler(t)

	body := models.CreateCustomerAndCardArgs{
		WalletAddress: "rWallet",
		Account: models.CardAccount{
			Currency: "USD",
			Card:     models.NewCardArgs{ProductCode: "PWSR_DEBP_2404"},
		},
		NameOnCard: "John",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/customers", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-gatehub-managed-user-uuid", consts.TestUser1ID)

	w := httptest.NewRecorder()
	h.CreateManagedCustomer(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// helper: create a customer+card in the store and return IDs
func seedCard(t *testing.T, store *storage.MemoryStorage) (customerID, accountID, cardID string) {
	t.Helper()
	cid := "cust-test-id"
	aid := "acct-test-id"
	caid := "card-test-id"

	customer := &models.Customer{
		ID:        &cid,
		SourceID:  consts.TestUser1ID,
		KYCStatus: consts.KYCStateAccepted,
	}
	require.NoError(t, store.CreateCustomer(customer))

	account := &models.Account{
		ID:               &aid,
		SourceID:         aid,
		CustomerID:       &cid,
		CustomerSourceID: consts.TestUser1ID,
		Currency:         "EUR",
		Status:           "ACTIVE",
	}
	require.NoError(t, store.CreateAccount(account))

	card := &models.Card{
		ID:               caid,
		SourceID:         caid,
		AccountID:        aid,
		CustomerID:       cid,
		CustomerSourceID: consts.TestUser1ID,
		NameOnCard:       "Test User",
		Status:           consts.CardStatusActive,
		MaskedPan:        "123456******7890",
	}
	require.NoError(t, store.CreateCard(card))
	return cid, aid, caid
}

// --- ListCards ---

func TestListCards_Success(t *testing.T) {
	h, store := setupCardsHandler(t)
	custID, _, _ := seedCard(t, store)

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/customers/"+custID+"/cards", nil)
	req = cardChiParams(req, map[string]string{"customerID": custID})

	w := httptest.NewRecorder()
	h.ListCards(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ListCardsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
}

func TestListCards_FiltersSoftDelete(t *testing.T) {
	h, store := setupCardsHandler(t)
	custID, _, cardID := seedCard(t, store)

	card, _ := store.GetCard(cardID)
	card.Status = consts.CardStatusSoftDelete
	require.NoError(t, store.UpdateCard(card))

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/customers/"+custID+"/cards", nil)
	req = cardChiParams(req, map[string]string{"customerID": custID})

	w := httptest.NewRecorder()
	h.ListCards(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ListCardsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
}

// --- GetCard ---

func TestGetCard_Success(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/cards/"+cardID, nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.GetCard(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetCard_NotFound(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/cards/no-such", nil)
	req = cardChiParams(req, map[string]string{"cardID": "no-such"})

	w := httptest.NewRecorder()
	h.GetCard(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- DeleteCard ---

func TestDeleteCard_Success(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	req := httptest.NewRequest(http.MethodDelete, "/cards/v1/cards/"+cardID, nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.DeleteCard(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	card, _ := store.GetCard(cardID)
	assert.Equal(t, consts.CardStatusSoftDelete, card.Status)
}

func TestDeleteCard_NotFound(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/cards/v1/cards/nope", nil)
	req = cardChiParams(req, map[string]string{"cardID": "nope"})

	w := httptest.NewRecorder()
	h.DeleteCard(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteCard_AlreadyDeleted(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	card, _ := store.GetCard(cardID)
	card.Status = consts.CardStatusSoftDelete
	_ = store.UpdateCard(card)

	req := httptest.NewRequest(http.MethodDelete, "/cards/v1/cards/"+cardID, nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.DeleteCard(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- LockCard ---

func TestLockCard_Success(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/"+cardID+"/lock", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.LockCard(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	card, _ := store.GetCard(cardID)
	assert.Equal(t, consts.CardStatusTemporaryBlocked, card.Status)
	assert.True(t, card.IsFirstTimeLock)
}

func TestLockCard_AlreadyLocked(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	card, _ := store.GetCard(cardID)
	card.Status = consts.CardStatusTemporaryBlocked
	_ = store.UpdateCard(card)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/"+cardID+"/lock", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.LockCard(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLockCard_DeletedCard(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	card, _ := store.GetCard(cardID)
	card.Status = consts.CardStatusSoftDelete
	_ = store.UpdateCard(card)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/"+cardID+"/lock", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.LockCard(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLockCard_BlockedCard(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	card, _ := store.GetCard(cardID)
	card.Status = consts.CardStatusBlocked
	_ = store.UpdateCard(card)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/"+cardID+"/lock", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.LockCard(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- UnlockCard ---

func TestUnlockCard_Success(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	card, _ := store.GetCard(cardID)
	card.Status = consts.CardStatusTemporaryBlocked
	_ = store.UpdateCard(card)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/"+cardID+"/unlock", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.UnlockCard(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	card, _ = store.GetCard(cardID)
	assert.Equal(t, consts.CardStatusActive, card.Status)
}

func TestUnlockCard_NotLocked(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/"+cardID+"/unlock", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.UnlockCard(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- BlockCard ---

func TestBlockCard_Success(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/"+cardID+"/block?reasonCode=Stolen", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.BlockCard(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	card, _ := store.GetCard(cardID)
	assert.Equal(t, consts.CardStatusBlocked, card.Status)
}

func TestBlockCard_DeletedCard(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	card, _ := store.GetCard(cardID)
	card.Status = consts.CardStatusSoftDelete
	_ = store.UpdateCard(card)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/"+cardID+"/block", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.BlockCard(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBlockCard_AlreadyBlocked(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	card, _ := store.GetCard(cardID)
	card.Status = consts.CardStatusBlocked
	_ = store.UpdateCard(card)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/"+cardID+"/block", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.BlockCard(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- GetPendingConfirmations ---

func TestGetPendingConfirmations_Empty(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/confirmations", nil)
	req.Header.Set("x-gatehub-managed-user-uuid", consts.TestUser1ID)

	w := httptest.NewRecorder()
	h.GetPendingConfirmations(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		PendingConfirmations []models.PendingThreeDSConfirmation `json:"pendingConfirmations"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.PendingConfirmations)
}

// --- CreateCustomerAddress ---

func TestCreateCustomerAddress_Success(t *testing.T) {
	h, store := setupCardsHandler(t)
	custID, _, _ := seedCard(t, store)

	body := map[string]interface{}{
		"type":        "RESIDENTIAL",
		"line1":       "123 Main St",
		"city":        "London",
		"zipCode":     "EC1A 1BB",
		"countryCode": "GB",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/customers/"+custID+"/addresses", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = cardChiParams(req, map[string]string{"customerID": custID})

	w := httptest.NewRecorder()
	h.CreateCustomerAddress(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp []models.CustomerDeliveryAddress
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
	assert.Equal(t, "123 Main St", resp[0].Line1)
}

// --- GetCustomerAddresses ---

func TestGetCustomerAddresses_Empty(t *testing.T) {
	h, store := setupCardsHandler(t)
	custID, _, _ := seedCard(t, store)

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/customers/"+custID+"/addresses", nil)
	req = cardChiParams(req, map[string]string{"customerID": custID})

	w := httptest.NewRecorder()
	h.GetCustomerAddresses(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- GetCardLimits ---

func TestGetCardLimits_SeedsDefaults(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/cards/"+cardID+"/limits", nil)
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.GetCardLimits(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var limits []models.CardLimit
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &limits))
	assert.Len(t, limits, 5)
}

// --- UpdateCardLimits ---

func TestUpdateCardLimits_Success(t *testing.T) {
	h, store := setupCardsHandler(t)
	_, _, cardID := seedCard(t, store)

	limits := []models.CardLimit{{Type: "dailyOverall", Limit: 2000, Currency: "EUR"}}
	b, _ := json.Marshal(limits)
	req := httptest.NewRequest(http.MethodPut, "/cards/v1/cards/"+cardID+"/limits", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = cardChiParams(req, map[string]string{"cardID": cardID})

	w := httptest.NewRecorder()
	h.UpdateCardLimits(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- GetCardToken ---

func TestGetCardToken_Success(t *testing.T) {
	h, _ := setupCardsHandler(t)

	pubKeyB64 := generateTestPublicKeyB64(t)
	body := map[string]interface{}{"cardId": "card-123", "publicKey": pubKeyB64}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/token/card-data", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.GetCardToken(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.CardTokenResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	// Token must be a parseable HS256 JWT carrying the cardId and publicKey.
	require.NotEmpty(t, resp.Token)
	claims, err := parseCardDataJWT(resp.Token)
	require.NoError(t, err)
	assert.Equal(t, "card-123", claims.CardID)
	assert.Equal(t, pubKeyB64, claims.PublicKey)

	// Link must be an absolute URL pointing at the public base URL so the
	// browser can fetch it directly.
	require.Len(t, resp.Links, 1)
	assert.True(t,
		strings.HasPrefix(resp.Links[0].Href, "https://") || strings.HasPrefix(resp.Links[0].Href, "http://"),
		"expected absolute URL, got %q", resp.Links[0].Href,
	)
	assert.Contains(t, resp.Links[0].Href, "/cards/v1/token/card-data/data")
	assert.Equal(t, "GET", resp.Links[0].Method)
	assert.Equal(t, "data", resp.Links[0].Rel)
}

func TestGetCardToken_MissingCardID(t *testing.T) {
	h, _ := setupCardsHandler(t)

	body := map[string]interface{}{"publicKey": "ignored"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/token/card-data", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.GetCardToken(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- GetCardData ---

func TestGetCardData_RoundTripDecryption(t *testing.T) {
	h, _ := setupCardsHandler(t)

	priv := generateTestRSAKey(t)
	pubKeyB64 := publicKeyToBase64SPKI(t, &priv.PublicKey)

	// 1. Get a card-data token using the public key.
	tokenBody := map[string]interface{}{"cardId": "card-123", "publicKey": pubKeyB64}
	tb, _ := json.Marshal(tokenBody)
	tReq := httptest.NewRequest(http.MethodPost, "/cards/v1/token/card-data", bytes.NewReader(tb))
	tReq.Header.Set("Content-Type", "application/json")
	tw := httptest.NewRecorder()
	h.GetCardToken(tw, tReq)
	require.Equal(t, http.StatusOK, tw.Code)

	var tokenResp models.CardTokenResponse
	require.NoError(t, json.Unmarshal(tw.Body.Bytes(), &tokenResp))

	// 2. Call the data endpoint with that token as a Bearer header.
	dReq := httptest.NewRequest(http.MethodGet, cardDataPath, nil)
	dReq.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	dw := httptest.NewRecorder()
	h.GetCardData(dw, dReq)
	require.Equal(t, http.StatusOK, dw.Code, "body=%s", dw.Body.String())

	var dataResp struct {
		Cypher string `json:"cypher"`
	}
	require.NoError(t, json.Unmarshal(dw.Body.Bytes(), &dataResp))
	require.NotEmpty(t, dataResp.Cypher)

	// 3. Decrypt the cypher with the matching private key and verify shape.
	cipher, err := base64.StdEncoding.DecodeString(dataResp.Cypher)
	require.NoError(t, err)
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, priv, cipher)
	require.NoError(t, err)

	var sensitive map[string]string
	require.NoError(t, json.Unmarshal(plain, &sensitive))
	assert.NotEmpty(t, sensitive["Pan"])
	assert.NotEmpty(t, sensitive["ExpiryDate"])
	assert.NotEmpty(t, sensitive["Cvc2"])
}

func TestGetCardData_MissingBearer(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodGet, cardDataPath, nil)
	w := httptest.NewRecorder()
	h.GetCardData(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCardData_InvalidToken(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodGet, cardDataPath, nil)
	req.Header.Set("Authorization", "Bearer not-a-real-jwt")
	w := httptest.NewRecorder()
	h.GetCardData(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCardData_TokenWithoutPublicKey(t *testing.T) {
	h, _ := setupCardsHandler(t)

	// Sign a JWT directly with no publicKey claim.
	jwt, err := generateCardDataJWT(CardDataClaims{
		CardID:    "card-x",
		IssuedAt:  1,
		ExpiresAt: 9999999999,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, cardDataPath, nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	h.GetCardData(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- CreateCardTransaction ---

func TestCreateCardTransaction_Success(t *testing.T) {
	h, _ := setupCardsHandler(t)

	body := map[string]interface{}{
		"cardId":   "card-123",
		"amount":   "25.50",
		"currency": "EUR",
		"type":     0,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/transactions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateCardTransaction(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var tx models.CardTransaction
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tx))
	assert.NotEmpty(t, tx.TransactionID)
	assert.Equal(t, "00", tx.GHResponseCode)
}

// --- GetCardTransaction ---

func TestGetCardTransaction_NotFound(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/transactions/nope", nil)
	req = cardChiParams(req, map[string]string{"txID": "nope"})

	w := httptest.NewRecorder()
	h.GetCardTransaction(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- ListCardTransactions ---

func TestListCardTransactions_Empty(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/cards/card-123/transactions", nil)
	req = cardChiParams(req, map[string]string{"cardID": "card-123"})

	w := httptest.NewRecorder()
	h.ListCardTransactions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.CardTransactionsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.Data)
}

// --- CreateThreeDSChallenge ---

func TestCreateThreeDSChallenge_Success(t *testing.T) {
	h, _ := setupCardsHandler(t)

	body := map[string]interface{}{
		"cardId":           "card-123",
		"userId":           consts.TestUser1ID,
		"merchantName":     "ACME",
		"purchaseAmount":   "10.00",
		"purchaseCurrency": "EUR",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/3ds", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateThreeDSChallenge(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var challenge models.ThreeDSChallenge
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &challenge))
	assert.Equal(t, "pending", challenge.Status)
}

// --- ConfirmThreeDS ---

func TestConfirmThreeDS_Approve(t *testing.T) {
	h, store := setupCardsHandler(t)

	challenge := &models.ThreeDSChallenge{
		TransactionID: "3ds-test",
		CardID:        "card-123",
		UserID:        consts.TestUser1ID,
		Status:        "pending",
	}
	require.NoError(t, store.CreateThreeDSChallenge(challenge))

	body := map[string]interface{}{"confirmed": true, "authMethod": "biometric"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/3ds/3ds-test/confirm", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = cardChiParams(req, map[string]string{"txID": "3ds-test"})

	w := httptest.NewRecorder()
	h.ConfirmThreeDS(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "approved", resp["status"])
}

func TestConfirmThreeDS_Decline(t *testing.T) {
	h, store := setupCardsHandler(t)

	challenge := &models.ThreeDSChallenge{
		TransactionID: "3ds-decline",
		CardID:        "card-123",
		UserID:        consts.TestUser1ID,
		Status:        "pending",
	}
	require.NoError(t, store.CreateThreeDSChallenge(challenge))

	body := map[string]interface{}{"confirmed": false, "authMethod": "pin"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/3ds/3ds-decline/confirm", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = cardChiParams(req, map[string]string{"txID": "3ds-decline"})

	w := httptest.NewRecorder()
	h.ConfirmThreeDS(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "declined", resp["status"])
}

func TestConfirmThreeDS_NotFound(t *testing.T) {
	h, _ := setupCardsHandler(t)

	body := map[string]interface{}{"confirmed": true, "authMethod": "pin"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cards/v1/3ds/nope/confirm", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = cardChiParams(req, map[string]string{"txID": "nope"})

	w := httptest.NewRecorder()
	h.ConfirmThreeDS(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- GetCardApplicationProducts ---

func TestGetCardApplicationProducts(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/cards/v1/applications/app-1/products", nil)
	req = cardChiParams(req, map[string]string{"appID": "app-1"})

	w := httptest.NewRecorder()
	h.GetCardApplicationProducts(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

// --- OrderPlasticCard ---

func TestOrderPlasticCard(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards/card-123/plastic", nil)
	req = cardChiParams(req, map[string]string{"cardID": "card-123"})

	w := httptest.NewRecorder()
	h.OrderPlasticCard(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "card-123", resp["cardId"])
	assert.Equal(t, "PLASTIC", resp["type"])
}

// --- OrderAdditionalCard ---

func TestOrderAdditionalCard(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/accounts/acct-1/cards", nil)
	req = cardChiParams(req, map[string]string{"accountID": "acct-1"})

	w := httptest.NewRecorder()
	h.OrderAdditionalCard(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "acct-1", resp["accountID"])
}

// ── CreateCustomer (delegates to CreateManagedCustomer) ──

func TestCreateCustomer_DelegatesToManaged(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := createManagedCustomerReq("Jane Doe")
	w := httptest.NewRecorder()
	h.CreateCustomer(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// ── CreateCard (stub) ──

func TestCreateCard_ReturnsActiveCard(t *testing.T) {
	h, _ := setupCardsHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/cards/v1/cards", nil)
	w := httptest.NewRecorder()
	h.CreateCard(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["id"])
	assert.Equal(t, consts.CardStatusActive, resp["status"])
}

// ── Test helpers for card-data crypto ──

func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
t.Helper()
key, err := rsa.GenerateKey(rand.Reader, 2048)
require.NoError(t, err)
return key
}

func publicKeyToBase64SPKI(t *testing.T, pub *rsa.PublicKey) string {
t.Helper()
der, err := x509.MarshalPKIXPublicKey(pub)
require.NoError(t, err)
return base64.StdEncoding.EncodeToString(der)
}

func generateTestPublicKeyB64(t *testing.T) string {
t.Helper()
return publicKeyToBase64SPKI(t, &generateTestRSAKey(t).PublicKey)
}
