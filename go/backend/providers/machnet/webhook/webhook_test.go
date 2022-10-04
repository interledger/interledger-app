package webhook_test

import (
	"bytes"
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
	ctrl := gomock.NewController(t)
	b := backends{machnet: machnet_mock.NewMockClient(ctrl)}
	wh := webhook.New(b)

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

	b.machnet.EXPECT().HandleEvent(gomock.Any(), userCardAddedEvent).Return(nil).Times(1)
	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(payload))
	response := httptest.NewRecorder()

	wh(response, req)

	assert.Equal(t, http.StatusOK, response.Code)
}

type backends struct {
	machnet *machnet_mock.MockClient
}

func (b backends) Machnet() machnet.Client {
	return b.machnet
}
