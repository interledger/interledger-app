package ops

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	email "gitlab.com/fynbos/backend/email"
	kycclient "gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/notify"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/pacioli"
	temporal "go.temporal.io/sdk/client"
)

func TestVerifyWebhook(t *testing.T) {
	if os.Getenv("GATEHUB_WEBHOOK_SECRET") == "" {
		t.SkipNow()
	}
	key, err := hex.DecodeString(os.Getenv("GATEHUB_WEBHOOK_SECRET"))
	if err != nil {
		require.NoError(t, err)
	}
	payload := `{"uuid":"44534833-c733-4c84-aac8-4daf758c0723","timestamp":"2024-04-22T12:12:31.220Z","event_type":"id.verification.accepted","user_uuid":"19227839-caa1-458f-a5ec-a3f03aa3e0e5","environment":"sandbox","data":{"gateway":"GateHub Crypto","verified":{"status":1,"short":"accepted"}}}`

	r, err := http.NewRequest(http.MethodPost, "https://en8pxah8zi194.x.pipedream.net", bytes.NewBufferString(payload))
	require.NoError(t, err)
	r.Header.Set("x-gh-webhook-timestamp", "2024-04-22T12:12:31.220Z")
	r.Header.Set("x-gh-webhook-signature", "785fd686e035d642a04c81f9c65bef1fd1bf69f9280ecf73b6ee41c55910fcd7")

	body, err := Verify(context.Background(), r, key)
	require.NoError(t, err)

	assert.Equal(t, payload, string(body))
}

type actionRequiredWebhookBackends struct {
	db *sqlx.DB
	kc kycclient.Client
}

func (b actionRequiredWebhookBackends) DB() *sqlx.DB                          { return b.db }
func (b actionRequiredWebhookBackends) LinkedAccounts() linkedaccounts.Client { return nil }
func (b actionRequiredWebhookBackends) Users() user.Client                    { return nil }
func (b actionRequiredWebhookBackends) Temporal() temporal.Client             { return nil }
func (b actionRequiredWebhookBackends) Wallets() wallets.Client               { return nil }
func (b actionRequiredWebhookBackends) Pacioli() pacioli.Client               { return nil }
func (b actionRequiredWebhookBackends) Email() email.Client                   { return nil }
func (b actionRequiredWebhookBackends) KYC() kycclient.Client                 { return b.kc }
func (b actionRequiredWebhookBackends) Transactions() transactions.Client     { return nil }
func (b actionRequiredWebhookBackends) Payments() payments.Client             { return nil }
func (b actionRequiredWebhookBackends) Notify() notify.Client                 { return nil }

func TestHandleActionRequiredWebhook_StatusDocumentsRequired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testDB := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	kc := kyc_mock.NewMockClient(ctrl)
	b := actionRequiredWebhookBackends{db: testDB, kc: kc}

	const walletID = "ac4fa1e0-2742-49fc-9a0f-fbb6d5859a7b"
	const externalUserID = "gatehub-user-123"
	_, err := testDB.ExecContext(ctx, "INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2)", externalUserID, walletID)
	require.NoError(t, err)

	tests := []struct {
		name      string
		eventType string
		reason    string
	}{
		{
			name:      "action required",
			eventType: "id.verification.action_required",
			reason:    "Additional documents required",
		},
		{
			name:      "document expired",
			eventType: "id.document_notice.expired",
			reason:    "Document expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := json.Marshal(ActionRequiredWebhook{
				UUID:      "2e607fd0-14d8-4e5e-955d-1cfd8b0a21ad",
				EventType: tt.eventType,
				UserID:    externalUserID,
				Data: ActionRequiredWebhookData{
					Gateway: "Paywiser",
					Reason:  tt.reason,
				},
			})
			require.NoError(t, err)

			kc.EXPECT().SetKYCStatus(ctx, walletID, kycclient.StatusDocumentsRequired).Return(nil)

			rr := httptest.NewRecorder()
			HandleActionRequiredWebhook(ctx, b, payload, rr)

			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}

func TestHandleActionRequiredWebhook_InvalidPayload(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	HandleActionRequiredWebhook(context.Background(), actionRequiredWebhookBackends{}, json.RawMessage(`{"event_type":`), rr)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleActionRequiredWebhook_WalletNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testDB := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	b := actionRequiredWebhookBackends{db: testDB, kc: kyc_mock.NewMockClient(ctrl)}

	payload, err := json.Marshal(ActionRequiredWebhook{
		EventType: "id.verification.action_required",
		UserID:    "missing-user",
		Data: ActionRequiredWebhookData{
			Gateway: "Paywiser",
			Reason:  "Missing docs",
		},
	})
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	HandleActionRequiredWebhook(ctx, b, payload, rr)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleActionRequiredWebhook_IgnoresOtherGateways(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testDB := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	b := actionRequiredWebhookBackends{db: testDB, kc: kyc_mock.NewMockClient(ctrl)}

	payload, err := json.Marshal(ActionRequiredWebhook{
		EventType: "id.verification.action_required",
		UserID:    "ignored-user",
		Data: ActionRequiredWebhookData{
			Gateway: "Other Gateway",
			Reason:  "Ignored",
		},
	})
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	HandleActionRequiredWebhook(ctx, b, payload, rr)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandleActionRequiredWebhook_SetKYCStatusError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testDB := db.MigrateTestDB(t, ctx)
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	kc := kyc_mock.NewMockClient(ctrl)
	b := actionRequiredWebhookBackends{db: testDB, kc: kc}

	const walletID = "ac4fa1e0-2742-49fc-9a0f-fbb6d5859a7b"
	const externalUserID = "gatehub-user-123"
	_, err := testDB.ExecContext(ctx, "INSERT INTO gatehub_users (external_id, wallet_id) VALUES ($1, $2)", externalUserID, walletID)
	require.NoError(t, err)

	payload, err := json.Marshal(ActionRequiredWebhook{
		EventType: "id.verification.action_required",
		UserID:    externalUserID,
		Data: ActionRequiredWebhookData{
			Gateway: "Paywiser",
			Reason:  "Additional documents required",
		},
	})
	require.NoError(t, err)

	kc.EXPECT().SetKYCStatus(ctx, walletID, kycclient.StatusDocumentsRequired).Return(errors.New("write failed"))

	rr := httptest.NewRecorder()
	HandleActionRequiredWebhook(ctx, b, payload, rr)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
