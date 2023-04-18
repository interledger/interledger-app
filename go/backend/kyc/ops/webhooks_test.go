package ops_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/fynbos/backend/kyc"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc/ops"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
)

func TestNewHandlePersonaWebhook(t *testing.T) {

	const jsonFmt = `{ "data": {
  "type": "event",
  "id": "evt_APAvuMVuwRQHqSrLSw1ExpJi",
  "attributes": {
    "name": "%s",
    "created-at": "2023-04-12T13:22:32.716Z",
    "redacted-at": null,
    "payload": {
      "data": {
        "type": "inquiry",
        "id": "inq_fzkoTpXvyFMea7HL8117sDS7",
        "attributes": {
          "status": "expired",
          "reference-id": "%s",
          "note": null
        }
      }
    }
  }
}}`
	ctx := context.Background()

	uc := user_mock.NewMock()

	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, uc)

	cases := []struct {
		name      string
		event     string
		sigHeader string
		walletID  string
		status    kyc.Status
	}{
		{
			name:      "approved",
			event:     "inquiry.approved",
			sigHeader: "t=123,v1=tjR9O4w+xHD8lYXTLxqGBXIzSMHYNM2DKXScm/jeVpI=",
			status:    kyc.StatusApproved,
			walletID:  "5312bb61-9d2d-43ee-b8de-3cb233e84bf5",
		},
		{
			name:      "review",
			event:     "inquiry.marked-for-review",
			sigHeader: "t=123,v1=eY78Dxo4loiLMan4/RQW/kz9vAj8fsQ3a780wx1qipQ=",
			status:    kyc.StatusInReview,
			walletID:  "ef10f4ad-7adb-4c81-add8-1d1bf63d603b",
		},
		{
			name:      "denied",
			event:     "inquiry.declined",
			sigHeader: "t=123,v1=wRkUujVXciAF0+dVrFyEgxHMnvBY810L4BxrWlS9W2g=",
			status:    kyc.StatusDenied,
			walletID:  "13235e11-456f-47d4-bc10-aa1d41ef16b2",
		},
		{
			name:      "expired",
			event:     "inquiry.expired",
			sigHeader: "t=123,v1=HZDWdeuyl2WSCgWa55OaUU2VmkZHV/QB6pAxr87lWL4=",
			status:    kyc.StatusUnknown,
			walletID:  "6c48c2af-382d-447b-82d4-f6f8739ff948",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(jsonFmt, tc.event, tc.walletID)
			req, err := http.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Persona-Signature", tc.sigHeader)

			rr := httptest.NewRecorder()
			handler := ops.NewHandlePersonaWebhook(b)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			kycStatus, err := ops.GetKYCStatus(ctx, b, tc.walletID)
			require.NoError(t, err)

			assert.Equal(t, tc.status, kycStatus)
		})
	}
}
