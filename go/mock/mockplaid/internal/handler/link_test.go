package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/fynbos/mock/mockplaid/internal/config"
	"gitlab.com/fynbos/mock/mockplaid/internal/storage"
)

func newTestHandler() (*Handler, storage.Storage) {
	store := storage.NewMemoryStorage()
	return NewHandler(store, &config.Config{}), store
}

func TestCreateLinkToken(t *testing.T) {
	h, store := newTestHandler()

	body := `{"user":{"client_user_id":"u_123"},"products":["auth"]}`
	req := httptest.NewRequest(http.MethodPost, "/link/token/create", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.CreateLinkToken(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		LinkToken  string `json:"link_token"`
		Expiration string `json:"expiration"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.HasPrefix(resp.LinkToken, "link-sandbox-") {
		t.Fatalf("unexpected link_token: %q", resp.LinkToken)
	}
	if resp.Expiration == "" {
		t.Fatal("expiration empty")
	}

	// Session persisted under the minted token.
	s, err := store.GetLinkSession(context.Background(), resp.LinkToken)
	if err != nil {
		t.Fatalf("GetLinkSession: %v", err)
	}
	if s.UserID != "u_123" {
		t.Fatalf("UserID mismatch: got %q", s.UserID)
	}
}
