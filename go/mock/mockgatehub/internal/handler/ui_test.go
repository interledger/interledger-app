package handler

import (
	"net/http"
	"net/http/httptest"
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
