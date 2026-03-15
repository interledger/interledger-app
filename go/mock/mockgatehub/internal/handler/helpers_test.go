package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/fynbos/mock/mockgatehub/internal/models"
	"gitlab.com/fynbos/mock/mockgatehub/internal/storage"
	"gitlab.com/fynbos/mock/mockgatehub/internal/webhook"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func helperHandler() *Handler {
	store := storage.NewMemoryStorage()
	wm := webhook.NewManager("", "test-secret", nil, nil, "")
	return NewHandler(store, wm)
}

func TestSendJSON(t *testing.T) {
	h := helperHandler()
	w := httptest.NewRecorder()
	h.sendJSON(w, http.StatusOK, map[string]string{"key": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "value", resp["key"])
}

func TestSendError(t *testing.T) {
	h := helperHandler()
	w := httptest.NewRecorder()
	h.sendError(w, http.StatusBadRequest, "something went wrong")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Bad Request", resp.Error)
	assert.Equal(t, "something went wrong", resp.Message)
}

func TestSetCORSHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	setCORSHeaders(w)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestSendJSONWithCORS(t *testing.T) {
	h := helperHandler()
	w := httptest.NewRecorder()
	h.sendJSONWithCORS(w, http.StatusOK, map[string]string{"ok": "true"})

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSendErrorWithCORS(t *testing.T) {
	h := helperHandler()
	w := httptest.NewRecorder()
	h.sendErrorWithCORS(w, http.StatusForbidden, "nope")

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDecodeJSON_Success(t *testing.T) {
	h := helperHandler()
	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	var v models.CreateManagedUserRequest
	err := h.decodeJSON(req, &v)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", v.Email)
}

func TestDecodeJSON_InvalidJSON(t *testing.T) {
	h := helperHandler()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))

	var v models.CreateManagedUserRequest
	err := h.decodeJSON(req, &v)
	assert.Error(t, err)
}
