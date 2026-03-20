package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"gitlab.com/fynbos/mock/mockpti/internal/config"
	"gitlab.com/fynbos/mock/mockpti/internal/models"
	"gitlab.com/fynbos/mock/mockpti/internal/storage"
)

func newTestHandler() *Handler {
	store := storage.NewMemoryStorage()
	cfg := &config.Config{
		ClientID: "test-client-id",
	}
	return NewHandler(store, cfg)
}

func newTestRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Post("/users", h.CreateUser)
		r.Get("/users/{id}", h.GetUser)
		r.Patch("/users", h.PatchUser)
		r.Put("/users", h.PutUser)
		r.Post("/users/assessments", h.StartUserAssessment)
		r.Get("/users/{id}/assessments", h.GetUserAssessment)
		r.Post("/users/{id}/wallets", h.CreateWallet)
		r.Get("/users/{id}/wallets", h.ListWallets)
		r.Get("/users/{id}/wallets/{walletId}", h.GetWallet)
		r.Post("/users/{id}/payment-information", h.CreatePaymentInformation)
		r.Get("/users/{id}/payment-information/{piId}", h.GetPaymentInformation)
		r.Post("/auth/jwt", h.CreateJWT)
	})
	return r
}

func ptiHeaders(req *http.Request) {
	req.Header.Set("x-pti-client-id", "test-client-id")
	req.Header.Set("Content-Type", "application/json")
}

func TestCreateUser_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.CreateUserRequest{
		ID:   "user-123",
		Type: "PERSON",
		Name: models.Name{First: "Alice", Last: "Smith"},
		Emails: []models.Email{
			{Address: "alice@example.com", Default: true},
		},
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID != "user-123" {
		t.Errorf("expected ID user-123, got %s", resp.ID)
	}
}

func TestCreateUser_GeneratesID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.CreateUserRequest{
		Type: "PERSON",
		Name: models.Name{First: "Bob"},
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp models.IDResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ID == "" {
		t.Error("expected generated ID, got empty string")
	}
}

func TestCreateUser_MissingClientID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.CreateUserRequest{Type: "PERSON"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	// No x-pti-client-id header
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestGetUser_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	// Create user first
	createBody := models.CreateUserRequest{
		ID:   "user-get-1",
		Type: "PERSON",
		Name: models.Name{First: "Charlie"},
	}
	createPayload, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createPayload))
	ptiHeaders(createReq)
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// Get user
	req := httptest.NewRequest(http.MethodGet, "/users/user-get-1", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var user models.User
	_ = json.NewDecoder(rr.Body).Decode(&user)

	if user.ID != "user-get-1" {
		t.Errorf("expected ID user-get-1, got %s", user.ID)
	}
	if user.Name.First != "Charlie" {
		t.Errorf("expected name Charlie, got %s", user.Name.First)
	}
	if user.Status != "active" {
		t.Errorf("expected status active, got %s", user.Status)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/users/nonexistent", nil)
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestPatchUser_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	// Create user first
	createBody := models.CreateUserRequest{
		ID:   "user-patch-1",
		Type: "PERSON",
		Name: models.Name{First: "Alice", Last: "Smith"},
	}
	createPayload, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createPayload))
	ptiHeaders(createReq)
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// Patch user name
	patchBody := models.PatchUserRequest{
		ID:   "user-patch-1",
		Name: &models.Name{First: "Bob", Last: "Jones"},
	}
	patchPayload, _ := json.Marshal(patchBody)
	req := httptest.NewRequest(http.MethodPatch, "/users", bytes.NewReader(patchPayload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify the update
	getReq := httptest.NewRequest(http.MethodGet, "/users/user-patch-1", nil)
	ptiHeaders(getReq)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	var user models.User
	_ = json.NewDecoder(getRR.Body).Decode(&user)

	if user.Name.First != "Bob" {
		t.Errorf("expected patched name Bob, got %s", user.Name.First)
	}
}

func TestPatchUser_NotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.PatchUserRequest{ID: "nonexistent"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/users", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestPatchUser_MissingID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.PatchUserRequest{Name: &models.Name{First: "No ID"}}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/users", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestPutUser_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	// Create user first
	createBody := models.CreateUserRequest{
		ID:   "user-put-1",
		Type: "PERSON",
		Name: models.Name{First: "Original"},
	}
	createPayload, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(createPayload))
	ptiHeaders(createReq)
	router.ServeHTTP(httptest.NewRecorder(), createReq)

	// Put (replace) user
	putBody := models.CreateUserRequest{
		ID:   "user-put-1",
		Type: "BUSINESS",
		Name: models.Name{First: "Replaced"},
	}
	putPayload, _ := json.Marshal(putBody)
	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewReader(putPayload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify replacement
	getReq := httptest.NewRequest(http.MethodGet, "/users/user-put-1", nil)
	ptiHeaders(getReq)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	var user models.User
	_ = json.NewDecoder(getRR.Body).Decode(&user)

	if user.Name.First != "Replaced" {
		t.Errorf("expected name Replaced, got %s", user.Name.First)
	}
	if user.Type != "BUSINESS" {
		t.Errorf("expected type BUSINESS, got %s", user.Type)
	}
	if user.Status != "active" {
		t.Errorf("expected preserved status active, got %s", user.Status)
	}
}

func TestPutUser_NotFound(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.CreateUserRequest{ID: "nonexistent", Type: "PERSON"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestPutUser_MissingID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.CreateUserRequest{Type: "PERSON"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
