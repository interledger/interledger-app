package webhook_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_mock "gitlab.com/fynbos/backend/providers/machnet/client/mock"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/providers/machnet/webhook"
)

func TestWebhook(t *testing.T) {
	t.Parallel()
	b := webhook.NewTestBackends(t)
	wh := webhook.New(b, "secret")

	userCardAddedEvent := external.Event{
		ID:             uuid.NewString(),
		EventName:      external.UserCardAdded,
		ResourceID:     uuid.NewString(),
		UserID:         uuid.NewString(),
		SubscriptionID: uuid.NewString(),
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Payload:        []byte("{}"),
	}
	payload, err := json.Marshal(userCardAddedEvent)
	require.NoError(t, err)
	webhookSignature := sign(t, payload, "succeed")

	b.machnet.EXPECT().ValidateWebhook(gomock.Any(), payload, webhookSignature).Return(nil).Times(1)
	b.machnet.EXPECT().HandleEvent(gomock.Any(), userCardAddedEvent).Return(nil).Times(1)
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(payload))
	req.Header.Set(webhook.SignatureHeader, webhookSignature)
	response := httptest.NewRecorder()

	wh(response, req)
	assert.Equal(t, http.StatusOK, response.Code)

	b.machnet.EXPECT().ValidateWebhook(gomock.Any(), payload, "fail").Return(machnet.ErrInvalidSignature).Times(1)
	badRequest := httptest.NewRequest("POST", "/", bytes.NewBuffer(payload))
	badRequest.Header.Set(webhook.SignatureHeader, "fail")
	response = httptest.NewRecorder()

	wh(response, badRequest)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func sign(t *testing.T, payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write(payload)
	if err != nil {
		t.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

type backends struct {
	machnet *machnet_mock.MockClient
}

func (b backends) Machnet() machnet.Client {
	return b.machnet
}
