package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/fynbos/mock/mockpti/internal/models"
)

func TestCreateJWT_Success(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.TokenRequest{
		URL:    "https://example.com/callback",
		Method: "POST",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/jwt", bytes.NewReader(payload))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.TokenResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.AccessToken == "" {
		t.Error("expected non-empty accessToken")
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("expected tokenType Bearer, got %s", resp.TokenType)
	}
	if resp.ExpiresAt <= 0 {
		t.Error("expected positive expiresAt")
	}
}

func TestCreateJWT_MissingClientID(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	body := models.TokenRequest{URL: "https://example.com", Method: "POST"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/jwt", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	// No x-pti-client-id
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestCreateJWT_InvalidBody(t *testing.T) {
	h := newTestHandler()
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/auth/jwt", bytes.NewReader([]byte("not json")))
	ptiHeaders(req)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
