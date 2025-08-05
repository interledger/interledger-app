package ops_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	email_client "gitlab.com/fynbos/backend/email/client/mock"
	"gitlab.com/fynbos/backend/kyc/ops"
	"gitlab.com/fynbos/backend/kyc/persona"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	wallet_mock "gitlab.com/fynbos/backend/wallets/client/mock"
)

func TestNewHandlePersonaWebhook(t *testing.T) {
	const inquiryJsonFmt = `{ "data": {
  "type": "event",
  "id": "evt_APAvuMVuwRQHqSrLSw1ExpJi",
  "attributes": {
    "name": "%s",
    "created-at": "2023-04-12T13:22:32.716Z",
    "redacted-at": null,
    "payload": {
      "data": {
        "type": "inquiry",
        "id": "%s",
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
	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})
	em := email_client.NewMockClient(ctrl)
	wc := wallet_mock.NewMockClient(ctrl)
	wc.EXPECT().Get(ctx, gomock.Any()).Return(&wallets.Wallet{}, nil).AnyTimes()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, uc, nil, nil, em, wc)

	inquiryCases := []struct {
		name          string
		event         string
		sigHeader     string
		walletID      string
		inquiryStatus persona.InquiryStatus
		inquiryID     string
	}{
		{
			name:          "approved",
			event:         "inquiry.approved",
			sigHeader:     "t=123,v1=b6347d3b8c3ec470fc9585d32f1a8605723348c1d834cd8329749c9bf8de5692",
			inquiryStatus: persona.InquiryApproved,
			walletID:      "5312bb61-9d2d-43ee-b8de-3cb233e84bf5",
			inquiryID:     "inq_fzkoTpXvyFMea7HL8117sDS7",
		},
		{
			name:          "review",
			event:         "inquiry.marked-for-review",
			sigHeader:     "t=123,v1=7f7d72c5cdcee5bae41a80179676cb67cec177d8f791d47b711e8eaa96d59c62",
			inquiryStatus: persona.InquiryNeedsReview,
			walletID:      "ef10f4ad-7adb-4c81-add8-1d1bf63d603b",
			inquiryID:     "inq_fzkoTpXvyFMea7HL8117sDS6",
		},
		{
			name:          "denied",
			event:         "inquiry.declined",
			sigHeader:     "t=123,v1=86a5856ae365963722a6d908b8b33c5a81a3a276e1209f3e5ab6b4ac8ba75bfb",
			inquiryStatus: persona.InquiryDeclined,
			walletID:      "13235e11-456f-47d4-bc10-aa1d41ef16b2",
			inquiryID:     "inq_fzkoTpXvyFMea7HL8117sDS5",
		},

		// the next events are for the same inquiry. Only the first one must do an update of our cached inquiry state
		{
			name:          "expired",
			event:         "inquiry.expired",
			sigHeader:     "t=123,v1=312b8e3a547f36d7d5603ff6874e7479ac6053462a655ba3d5252522f2a016db",
			inquiryStatus: persona.InquiryExpired,
			walletID:      "6c48c2af-382d-447b-82d4-f6f8739ff948",
			inquiryID:     "inq_fzkoTpXvyFMea7HL8117sDS4",
		},
		{
			name:          "failed",
			event:         "inquiry.failed",
			sigHeader:     "t=123,v1=fc60f22e6aafccc5fac7d4bb27d8e6f8be4a09d053b703f274729b114408dbf7",
			inquiryStatus: persona.InquiryFailed,
			walletID:      "6c48c2af-382d-447b-82d4-f6f8739ff948",
			inquiryID:     "inq_fzkoTpXvyFMea7HL8117sDS4",
		},
		{
			name:          "created",
			event:         "inquiry.created",
			sigHeader:     "t=123,v1=a0798484624f314c3b3c971129ad383d590f757d42306c5d8ec89267cd703e53",
			inquiryStatus: persona.InquiryCreated,
			walletID:      "6c48c2af-382d-447b-82d4-f6f8739ff948",
			inquiryID:     "inq_fzkoTpXvyFMea7HL8117sDS4",
		},
		{
			name:          "started",
			event:         "inquiry.started",
			sigHeader:     "t=123,v1=2ba55d3f496c35e80a5582fdfc4c02f36c7dc3d37dcd79f3256e90ea1f1db63d",
			inquiryStatus: persona.InquiryPending,
			walletID:      "6c48c2af-382d-447b-82d4-f6f8739ff948",
			inquiryID:     "inq_fzkoTpXvyFMea7HL8117sDS4",
		},
	}

	for _, tc := range inquiryCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = b.DB().MustExec("INSERT INTO kyc_persona_inquiries (external_id, state, wallet_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;", tc.inquiryID, persona.InquiryCreated, tc.walletID)

			body := fmt.Sprintf(inquiryJsonFmt, tc.event, tc.inquiryID, tc.walletID)
			req, err := http.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Persona-Signature", tc.sigHeader)

			rr := httptest.NewRecorder()
			handler := ops.NewHandlePersonaWebhook(b)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var inquiryStatus string
			err = b.DB().Get(&inquiryStatus, "SELECT state from kyc_persona_inquiries WHERE wallet_id=$1 AND external_id=$2;", tc.walletID, tc.inquiryID)
			require.NoError(t, err)

			if tc.inquiryID == "inq_fzkoTpXvyFMea7HL8117sDS4" { // should only update using the latest webhook
				assert.Equal(t, persona.InquiryExpired, persona.InquiryStatus(inquiryStatus))
			} else {
				assert.Equal(t, tc.inquiryStatus, persona.InquiryStatus(inquiryStatus))
			}
		})
	}

	const accountJsonFmt = `{ "data": {
  "type": "event",
  "id": "evt_U8thNQcQUPBp35DgUDGbRtZK",
  "attributes": {
    "name": "%s",
    "created-at": "2023-04-12T13:22:32.716Z",
    "redacted-at": null,
    "payload": {
      "data": {
        "type": "account",
        "id": "%s",
        "attributes": {
          "reference-id": "%s"
        }
      }
    }
  }
}}`

	accountCases := []struct {
		event            string
		sigHeader        string
		walletID         string
		personaAccountID string
	}{
		{
			event:            "account.created",
			sigHeader:        "t=123,v1=80723ee09601ba6e9d849f1c0d02241f1b50b2300020f748f2b703ac272c1d62",
			walletID:         "6c48c2af-382d-447b-82d4-f6f8739ff948",
			personaAccountID: "act_1LtncnQbaLme9PgR5LQuga7k",
		},
	}
	for _, tc := range accountCases {
		body := fmt.Sprintf(accountJsonFmt, tc.event, tc.personaAccountID, tc.walletID)
		req, err := http.NewRequest(http.MethodPost, "/webhook", strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Persona-Signature", tc.sigHeader)

		rr := httptest.NewRecorder()
		handler := ops.NewHandlePersonaWebhook(b)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var walletID string
		err = b.DB().Get(&walletID, "SELECT wallet_id FROM kyc_persona_accounts WHERE external_id=$1;", tc.personaAccountID)
		require.NoError(t, err)
		assert.Equal(t, tc.walletID, walletID)
	}
}
