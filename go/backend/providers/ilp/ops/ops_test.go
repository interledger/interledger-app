package ops_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/ilp"
	"gitlab.com/fynbos/backend/providers/ilp/ops"
)

type streamHandler struct {
	t *testing.T
}

func (h streamHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	require.NoError(h.t, err)

	var payload struct {
		PaymentTag string `json:"paymentTag"`
		Asset      struct {
			Code  string
			Scale uint8
		}
	}
	err = json.Unmarshal(body, &payload)
	require.NoError(h.t, err)
	require.Equal(h.t, "test", payload.PaymentTag)
	require.Equal(h.t, "USD", payload.Asset.Code)
	require.Equal(h.t, uint8(2), payload.Asset.Scale)

	ret, err := json.Marshal(ilp.StreamCredentials{
		SharedSecret: "lalsasfasf",
		IlpAddress:   "t.fynbos.123",
	})
	require.NoError(h.t, err)

	_, err = resp.Write(ret)
	require.NoError(h.t, err)
}

func TestCreateStreamCredentials(t *testing.T) {
	streamServer := httptest.NewServer(streamHandler{t})
	t.Cleanup(func() {
		streamServer.Close()
	})
	b := backends{
		streamServerUrl: streamServer.URL,
	}

	creds, err := ops.CreateStreamCredentials(context.Background(), b, ilp.CreateStreamCredentialsArgs{
		PaymentTag: "test",
		Currency:   currency.ParseCurrency("USD"),
	})
	require.NoError(t, err)

	assert.Equal(t, "t.fynbos.123", creds.IlpAddress)
	assert.Equal(t, "lalsasfasf", creds.SharedSecret)
}

func TestClearIncomingPackets(t *testing.T) {
	b := backends{
		db: db.MigrateTestDB(t, context.Background()),
	}

	err := ops.ClearIncomingPackets(context.Background(), b, []ilp.IncomingPacket{
		{
			PaymentTag: uuid.NewString(),
			Amount: currency.Amount{
				Value:    10,
				Currency: currency.ParseCurrency("USD"),
			},
			Peer: "coil",
		},
		{
			PaymentTag: uuid.NewString(),
			Amount: currency.Amount{
				Value:    10,
				Currency: currency.ParseCurrency("USD"),
			},
			Peer: "coil",
		},
	})
	require.NoError(t, err)

	var count int
	err = b.db.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM ilp_packets;")
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

type backends struct {
	db              *sqlx.DB
	streamServerUrl string
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func (b backends) StreamServerBaseURL() string {
	return b.streamServerUrl
}
