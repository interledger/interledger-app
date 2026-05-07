package apperrors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/appcontext"
	"gitlab.com/fynbos/backend/errcodes"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
)

type appErrorBody struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
	ReqID     string `json:"req_id,omitempty"`
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) appErrorBody {
	t.Helper()
	var body appErrorBody
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	return body
}

func newRequest(t *testing.T) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	return r
}

func TestWriteAppError(t *testing.T) {
	t.Parallel()

	t.Run("sets content-type, status and body", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := newRequest(t)

		WriteAppError(w, r, http.StatusNotFound, errcodes.ErrCodeNotFound, "not found")

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		body := decodeBody(t, w)
		assert.Equal(t, errcodes.ErrCodeNotFound, body.ErrorCode)
		assert.Equal(t, "not found", body.Message)
		assert.Empty(t, body.ReqID)
	})

	t.Run("includes req_id from context", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := newRequest(t)
		r = r.WithContext(appcontext.WithRequestID(r.Context(), "req-abc-123"))

		WriteAppError(w, r, http.StatusBadRequest, errcodes.ErrCodeBadRequest, "bad request")

		body := decodeBody(t, w)
		assert.Equal(t, "req-abc-123", body.ReqID)
	})

	t.Run("omits req_id when context has none", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := newRequest(t)

		WriteAppError(w, r, http.StatusOK, "CODE", "msg")

		body := decodeBody(t, w)
		assert.Empty(t, body.ReqID)
	})
}

func TestToHTTPError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{"user not found", user.ErrNoUserFound, http.StatusUnauthorized, errcodes.ErrCodeUserNoUserFound},
		{"no wallet found", wallets.ErrNoWalletFound, http.StatusNotFound, errcodes.ErrCodeWalletsNoWalletFound},
		{"duplicate wallet", wallets.ErrDuplicateWallet, http.StatusConflict, errcodes.ErrCodeWalletsDuplicateWallet},
		{"wallet conflict", wallets.ErrWalletConflict, http.StatusConflict, errcodes.ErrCodeWalletsWalletConflict},
		{"linked account not found", linkedaccounts.ErrNotFound, http.StatusNotFound, errcodes.ErrCodeLinkedAccNotFound},
		{"kyc resubmission required", kyc.ErrKYCResubmissionRequired, http.StatusForbidden, errcodes.ErrCodeKYCResubmissionRequired},
		{"gatehub not found", gatehub.ErrNotFound, http.StatusNotFound, errcodes.ErrCodeNotFound},
		{"gatehub internal", gatehub.ErrInternal, http.StatusInternalServerError, errcodes.ErrCodeInternal},
		{"gatehub timed out", gatehub.ErrTimedOut, http.StatusGatewayTimeout, errcodes.ErrCodeGatewayTimeout},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			r := newRequest(t)

			ToHTTPError(w, r, tc.err)

			assert.Equal(t, tc.expectedStatus, w.Code)
			body := decodeBody(t, w)
			assert.Equal(t, tc.expectedCode, body.ErrorCode)
		})
	}

	t.Run("wrapped known error is matched via errors.Is", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := newRequest(t)

		ToHTTPError(w, r, fmt.Errorf("context: %w", user.ErrNoUserFound))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		body := decodeBody(t, w)
		assert.Equal(t, errcodes.ErrCodeUserNoUserFound, body.ErrorCode)
	})

	t.Run("unknown error returns 500 internal", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := newRequest(t)

		ToHTTPError(w, r, errors.New("something unexpected"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		body := decodeBody(t, w)
		assert.Equal(t, errcodes.ErrCodeInternal, body.ErrorCode)
	})
}
