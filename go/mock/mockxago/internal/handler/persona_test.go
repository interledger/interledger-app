package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestPersonaGetInquiryIframe_UsesInquiryID(t *testing.T) {
	h := setupTestHandler(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change dir to repo root: %v", err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	req := httptest.NewRequest(http.MethodGet, "/v1/inquiries/inq_123/iframe?user_id=wallet-abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", "inq_123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetInquiryIframe(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `name="user_id" value="inq_123"`)
	assert.NotContains(t, body, `value="wallet-abc"`)
}

func TestPersonaGetInquiryIframe_DefaultsToInquiryID(t *testing.T) {
	h := setupTestHandler(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join(cwd, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to change dir to repo root: %v", err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()

	req := httptest.NewRequest(http.MethodGet, "/v1/inquiries/inq_456/iframe", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("inquiryId", "inq_456")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.PersonaGetInquiryIframe(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `name="user_id" value="inq_456"`)
}
