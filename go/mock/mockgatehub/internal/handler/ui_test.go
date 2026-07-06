package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/consts"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/models"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/storage"
	"github.com/interledger/interledger-app/go/mock/mockgatehub/internal/webhook"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUIHandler(store storage.Storage) *Handler {
	wm := webhook.NewManager("", "test-secret", nil, nil, "")
	return NewHandler(store, wm)
}

// ── Dashboard ─────────────────────────────────────────────────────────────────

func TestUIDashboard_Empty(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	rr := httptest.NewRecorder()
	h.UIDashboard(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "MockGatehub Admin")
}

func TestUIDashboard_WithUsers(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	rr := httptest.NewRecorder()
	h.UIDashboard(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), consts.TestUser1Email)
	assert.Contains(t, rr.Body.String(), consts.TestUser2Email)
}

func TestUIDashboard_ContentType(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	rr := httptest.NewRecorder()
	h.UIDashboard(rr, req)
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")
}

// ── User Detail ───────────────────────────────────────────────────────────────

func TestUIUserDetail_NotFound(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	req := httptest.NewRequest(http.MethodGet, "/ui/users/nonexistent", nil)
	req = withChiParams(req, map[string]string{"userID": "nonexistent"})
	rr := httptest.NewRecorder()
	h.UIUserDetail(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUIUserDetail_Found(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/users/"+consts.TestUser1ID, nil)
	req = withChiParams(req, map[string]string{"userID": consts.TestUser1ID})
	rr := httptest.NewRecorder()
	h.UIUserDetail(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), consts.TestUser1Email)
}

func TestUIUserDetail_ShowsBalances(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/users/"+consts.TestUser1ID, nil)
	req = withChiParams(req, map[string]string{"userID": consts.TestUser1ID})
	rr := httptest.NewRecorder()
	h.UIUserDetail(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "USD")
}

func TestUIUserDetail_ShowsTransactions(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	require.NoError(t, store.CreateTransaction(&models.Transaction{
		ID:       "tx-ui-detail-1",
		UserID:   consts.TestUser1ID,
		Amount:   "42.00",
		Currency: "USD",
	}))
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/users/"+consts.TestUser1ID, nil)
	req = withChiParams(req, map[string]string{"userID": consts.TestUser1ID})
	rr := httptest.NewRecorder()
	h.UIUserDetail(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "tx-ui-detail-1")
}

func TestUIUserDetail_ShowsCardTransactionsSeparately(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	_, card := seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusActive)

	numericCardID := 123456
	amount := "18.50"
	currency := "EUR"
	status := "COMPLETED"
	ctx := &models.CardTransaction{
		TransactionID:       "ctx-ui-detail-1",
		CardID:              &numericCardID,
		TransactionAmount:   &amount,
		TransactionCurrency: &currency,
		TxStatus:            &status,
		CreatedAt:           "2026-05-12T12:00:00Z",
		GHResponseCode:      "00",
	}
	require.NoError(t, store.CreateCardTransaction(ctx))
	require.NoError(t, store.AddCardTransactionIndex(card.ID, ctx.TransactionID))

	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/users/"+consts.TestUser1ID, nil)
	req = withChiParams(req, map[string]string{"userID": consts.TestUser1ID})
	rr := httptest.NewRecorder()
	h.UIUserDetail(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Transactions")
	assert.Contains(t, rr.Body.String(), "Card Transactions")
	assert.Contains(t, rr.Body.String(), "ctx-ui-detail-1")
	assert.Contains(t, rr.Body.String(), "123456")
}

func TestUIUserDetail_ContentType(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/users/"+consts.TestUser1ID, nil)
	req = withChiParams(req, map[string]string{"userID": consts.TestUser1ID})
	rr := httptest.NewRecorder()
	h.UIUserDetail(rr, req)
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")
}

// ── KYC Form ──────────────────────────────────────────────────────────────────

func TestUIKYCForm_RendersUsers(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/actions/kyc", nil)
	rr := httptest.NewRecorder()
	h.UIKYCForm(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), consts.TestUser1Email)
}

func TestUIKYCForm_PreSelectsUser(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/actions/kyc?userID="+consts.TestUser1ID, nil)
	rr := httptest.NewRecorder()
	h.UIKYCForm(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `selected`)
}

func TestUIKYCForm_ShowsFlash(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	req := httptest.NewRequest(http.MethodGet, "/ui/actions/kyc?flash=KYC+event+sent&ok=1", nil)
	rr := httptest.NewRecorder()
	h.UIKYCForm(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "KYC event sent")
}

// ── KYC Action ────────────────────────────────────────────────────────────────

func TestUIKYCAction_MissingUser(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/kyc",
		strings.NewReader("gateway=paywiser-eu-sandbox&status=accepted"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIKYCAction(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "flash=User+is+required")
}

func TestUIKYCAction_UnknownUser(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/kyc",
		strings.NewReader("userID=nonexistent&gateway=paywiser&status=accepted"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIKYCAction(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "flash=User+not+found")
}

func TestUIKYCAction_AcceptedUpdatesState(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&gateway=paywiser-eu-sandbox&status=accepted"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/kyc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIKYCAction(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "ok=1")
	user, err := store.GetUser(consts.TestUser1ID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateAccepted, user.KYCState)
}

func TestUIKYCAction_RejectedUpdatesState(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&gateway=paywiser&status=rejected"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/kyc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIKYCAction(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	user, err := store.GetUser(consts.TestUser1ID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateRejected, user.KYCState)
}

func TestUIKYCAction_PendingMapsToActionRequired(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&gateway=paywiser&status=pending"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/kyc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIKYCAction(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	user, err := store.GetUser(consts.TestUser1ID)
	require.NoError(t, err)
	assert.Equal(t, consts.KYCStateActionRequired, user.KYCState)
}

func TestUIKYCAction_DefaultGatewayWhenEmpty(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&status=accepted"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/kyc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIKYCAction(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "ok=1")
}

// ── Card TX Cards (AJAX helper) ───────────────────────────────────────────────

func seedCustomerAndCard(t *testing.T, store storage.Storage, userID, cardStatus string) (*models.Customer, *models.Card) {
	t.Helper()
	custID := "cust-" + userID
	customer := &models.Customer{ID: &custID, SourceID: userID, Type: "managed"}
	require.NoError(t, store.CreateCustomer(customer))
	card := &models.Card{
		ID:         "card-" + userID,
		CustomerID: custID,
		MaskedPan:  "512345******0001",
		Status:     cardStatus,
	}
	require.NoError(t, store.CreateCard(card))
	return customer, card
}

func TestUICardTxCards_NoCustomer(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/actions/card-transaction/cards?userID="+consts.TestUser1ID, nil)
	rr := httptest.NewRecorder()
	h.UICardTxCards(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	var cards []interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &cards))
	assert.Empty(t, cards)
}

func TestUICardTxCards_WithCards(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusActive)
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/actions/card-transaction/cards?userID="+consts.TestUser1ID, nil)
	rr := httptest.NewRecorder()
	h.UICardTxCards(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	var cards []map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &cards))
	assert.Len(t, cards, 1)
}

func TestUICardTxCards_FiltersInactive(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusBlocked)
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/actions/card-transaction/cards?userID="+consts.TestUser1ID, nil)
	rr := httptest.NewRecorder()
	h.UICardTxCards(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	var cards []interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &cards))
	assert.Empty(t, cards)
}

// ── Card TX Form ──────────────────────────────────────────────────────────────

func TestUICardTxForm_Renders(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ui/actions/card-transaction", nil)
	rr := httptest.NewRecorder()
	h.UICardTxForm(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), consts.TestUser1Email)
	assert.Contains(t, rr.Body.String(), "Webhook Title")
	assert.Contains(t, rr.Body.String(), "Card Transaction Object (JSON)")
}

// ── Card TX Action ────────────────────────────────────────────────────────────

func TestUICardTxAction_HappyPath(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	_, card := seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusActive)
	h := newUIHandler(store)

	values := url.Values{}
	values.Set("userID", consts.TestUser1ID)
	values.Set("cardID", card.ID)
	values.Set("title", "Card Purchase")
	values.Set("body", "EUR 50.00 at Test Shop")
	values.Set("cardTxObject", `{"id":999,"transactionId":"user-provided-tx-id","createdAt":"2001-01-01T00:00:00.000Z","cardId":987654,"transactionAmount":"50.00","transactionCurrency":"USD","billingAmount":"50.00","billingCurrency":"USD","type":0,"merchantName":"Test Shop","customField":"keep-me"}`)

	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxAction(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "/ui/users/"+consts.TestUser1ID)

	ids, err := store.GetCardTransactionIDs(card.ID)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	created, err := store.GetCardTransaction(ids[0])
	require.NoError(t, err)
	require.NotNil(t, created.CardID)
	assert.Equal(t, 987654, *created.CardID)
	assert.Equal(t, ids[0], created.TransactionID)
	assert.Equal(t, uiCardTxCreatedAt, created.CreatedAt)

	raw, err := store.GetRawCardTransaction(ids[0])
	require.NoError(t, err)
	var rawMap map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &rawMap))
	assert.Equal(t, "keep-me", rawMap["customField"])
	assert.Equal(t, uiCardTxCreatedAt, rawMap["createdAt"])
	assert.NotEqual(t, "user-provided-tx-id", rawMap["transactionId"])
	require.Contains(t, rawMap, "id")
	assert.Equal(t, float64(1), rawMap["id"])

	bal, err := store.GetBalance(consts.TestUser1ID, "USD")
	require.NoError(t, err)
	assert.Equal(t, 10000.0, bal)
}

func TestUICardTxPreview_ShowsNormalizedCallbackAndTransaction(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	_, card := seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusActive)
	h := newUIHandler(store)

	values := url.Values{}
	values.Set("userID", consts.TestUser1ID)
	values.Set("cardID", card.ID)
	values.Set("title", "Card Purchase")
	values.Set("body", "EUR 50.00 at Test Shop")
	values.Set("cardTxObject", `{"id":888,"transactionId":"tx-from-user","createdAt":"2000-01-01T00:00:00.000Z","transactionAmount":"50.00","transactionCurrency":"USD","customField":"keep-me"}`)

	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction/preview", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxPreview(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var preview map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &preview))
	assert.Equal(t, float64(1), preview["nextTransactionId"])

	webhook, ok := preview["webhook"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, consts.WebhookEventCardTransaction, webhook["event_type"])
	assert.Equal(t, consts.TestUser1ID, webhook["user_uuid"])

	transaction, ok := preview["transaction"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, uiCardTxCreatedAt, transaction["createdAt"])
	assert.NotEqual(t, "tx-from-user", transaction["transactionId"])
	assert.Equal(t, float64(1), transaction["id"])
	assert.Equal(t, "keep-me", transaction["customField"])
}

func TestUICardTxAction_InvalidCardTransactionObject(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	_, card := seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusActive)
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&cardID=" + card.ID + "&title=Card+Purchase&body=foo&cardTxObject={bad-json"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxAction(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUICardTxAction_InactiveCard(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	_, card := seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusBlocked)
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&cardID=" + card.ID + "&title=Card+Purchase&body=foo&cardTxObject={\"transactionAmount\":\"10.00\",\"transactionCurrency\":\"USD\"}"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxAction(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUICardTxAction_UnknownCard(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&cardID=nonexistent&title=Card+Purchase&body=foo&cardTxObject={\"transactionAmount\":\"10.00\",\"transactionCurrency\":\"USD\"}"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxAction(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── Card TX Set Status ────────────────────────────────────────────────────────

func TestUICardTxSetStatus_Complete(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	_, card := seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusActive)

	status := "PENDING"
	tx := &models.CardTransaction{TransactionID: "ctx-status-1", TxStatus: &status}
	require.NoError(t, store.CreateCardTransaction(tx))
	require.NoError(t, store.AddCardTransactionIndex(card.ID, tx.TransactionID))

	h := newUIHandler(store)
	body := "txID=ctx-status-1&status=COMPLETED&userID=" + consts.TestUser1ID
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxSetStatus(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "/ui/users/"+consts.TestUser1ID)
	updated, err := store.GetCardTransaction("ctx-status-1")
	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", *updated.TxStatus)
}

func TestUICardTxSetStatus_Fail(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	_, card := seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusActive)

	status := "PENDING"
	tx := &models.CardTransaction{TransactionID: "ctx-status-2", TxStatus: &status}
	require.NoError(t, store.CreateCardTransaction(tx))
	require.NoError(t, store.AddCardTransactionIndex(card.ID, tx.TransactionID))

	h := newUIHandler(store)
	body := "txID=ctx-status-2&status=FAILED&userID=" + consts.TestUser1ID
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxSetStatus(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
}

func TestUICardTxSetStatus_MissingFields(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction/status", strings.NewReader("txID=x"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxSetStatus(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUICardTxSetStatus_InvalidStatus(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	body := "txID=x&status=UNKNOWN&userID=" + consts.TestUser1ID
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxSetStatus(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// ── Withdrawal Set Event ──────────────────────────────────────────────────────

func seedPendingWithdrawal(t *testing.T, store storage.Storage, userID, txID, currency string, amount float64) {
	t.Helper()
	require.NoError(t, store.AddBalance(userID, currency, amount))
	require.NoError(t, store.CreateTransaction(&models.Transaction{
		ID:          txID,
		UserID:      userID,
		Amount:      "50.00",
		TotalAmount: "50.00",
		Fee:         "0.00",
		Currency:    currency,
		DepositType: consts.DepositTypeWithdrawal,
		Status:      consts.TransactionStatusPending,
	}))
}

func TestUIWithdrawalSetEvent_Complete(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	seedPendingWithdrawal(t, store, consts.TestUser1ID, "wd-complete-1", "EUR", 100.0)

	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&txID=wd-complete-1&event=core.withdrawal.completed"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/withdrawal/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIWithdrawalSetEvent(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "/ui/users/"+consts.TestUser1ID)
	tx, err := store.GetTransaction("wd-complete-1")
	require.NoError(t, err)
	assert.Equal(t, consts.TransactionStatusCompleted, tx.Status)
}

func TestUIWithdrawalSetEvent_Reject(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	seedPendingWithdrawal(t, store, consts.TestUser1ID, "wd-reject-1", "EUR", 100.0)

	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&txID=wd-reject-1&event=more-bridge.withdrawal.rejected"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/withdrawal/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIWithdrawalSetEvent(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "/ui/users/"+consts.TestUser1ID)
	tx, err := store.GetTransaction("wd-reject-1")
	require.NoError(t, err)
	assert.Equal(t, consts.TransactionStatusFailed, tx.Status)
}

func TestUIWithdrawalSetEvent_MissingFields(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/withdrawal/status", strings.NewReader("userID=x"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIWithdrawalSetEvent(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUIWithdrawalSetEvent_InvalidEvent(t *testing.T) {
	h := newUIHandler(storage.NewMemoryStorage())
	body := "userID=x&txID=y&event=bogus.event"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/withdrawal/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIWithdrawalSetEvent(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUIWithdrawalSetEvent_TxNotFound(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&txID=nonexistent&event=core.withdrawal.completed"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/withdrawal/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIWithdrawalSetEvent(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "/ui/users/"+consts.TestUser1ID)
}

func TestUIWithdrawalSetEvent_NotPending(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	require.NoError(t, store.CreateTransaction(&models.Transaction{
		ID:          "wd-done-1",
		UserID:      consts.TestUser1ID,
		DepositType: consts.DepositTypeWithdrawal,
		Status:      consts.TransactionStatusCompleted,
	}))
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&txID=wd-done-1&event=core.withdrawal.completed"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/withdrawal/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UIWithdrawalSetEvent(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "/ui/users/"+consts.TestUser1ID)
}
