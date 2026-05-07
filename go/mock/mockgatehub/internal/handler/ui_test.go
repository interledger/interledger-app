package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/fynbos/mock/mockgatehub/internal/consts"
	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/webhook"

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
}

// ── Card TX Action ────────────────────────────────────────────────────────────

func TestUICardTxAction_HappyPath(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	_, card := seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusActive)
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&cardID=" + card.ID + "&amount=50.00&currency=USD&txType=0&merchantName=Test+Shop"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxAction(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "/ui/users/"+consts.TestUser1ID)
	bal, err := store.GetBalance(consts.TestUser1ID, "USD")
	require.NoError(t, err)
	assert.Equal(t, 9950.0, bal)
}

func TestUICardTxAction_InsufficientFunds(t *testing.T) {
	store := storage.NewMemoryStorage()
	require.NoError(t, storage.SeedTestUsers(store))
	_, card := seedCustomerAndCard(t, store, consts.TestUser1ID, consts.CardStatusActive)
	h := newUIHandler(store)
	body := "userID=" + consts.TestUser1ID + "&cardID=" + card.ID + "&amount=999999.00&currency=USD&txType=0"
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
	body := "userID=" + consts.TestUser1ID + "&cardID=" + card.ID + "&amount=10.00&currency=USD&txType=0"
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
	body := "userID=" + consts.TestUser1ID + "&cardID=nonexistent&amount=10.00&currency=USD&txType=0"
	req := httptest.NewRequest(http.MethodPost, "/ui/actions/card-transaction", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UICardTxAction(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
