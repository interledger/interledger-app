package ops_test

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/chimoney/ops"
	"gotest.tools/assert"
)

func TestVerifyWebhook(t *testing.T) {
	if os.Getenv("CHIMONEY_WEBHOOK_SECRET") == "" {
		t.SkipNow()
	}
	secret := ops.ParseWebhookSecret(os.Getenv("CHIMONEY_WEBHOOK_SECRET"))
	if secret == nil {
		t.Fatal("Failed to parse CHIMONEY_WEBHOOK_SECRET")
	}

	payload := `{"data":{"account_number":"0690000040","amount":5500,"bank_code":"044","bank_name":"","complete_message":"","created_at":"2020-01-20T16:09:34.000Z","currency":"NGN","debit_currency":"NGN","eventType":"payout.bank.completed","fee":45,"full_name":"Test Name","id":26251,"is_approved":1,"meta":{"chiRef":"01eab82e-4044-452f-80a2-ebb3727fb2b4","country":"NG","currency":"NGN","type":"bank","valueInUSD":5},"narration":"developer transfer xx007","reference":"01eab82e-4044-452f-80a2-ebb3727fb2b4_1678089714855","requires_approval":0,"status":"SUCCESSFUL"},"eventType":"payout.bank.completed","status":"success"}`

	r, err := http.NewRequest(http.MethodPost, "https://enou3fn0u1me.x.pipedream.net/", bytes.NewBufferString(payload))
	require.NoError(t, err)
	r.Header.Set("svix-id", "msg_2iVgfd8bsmhhjLhRLVDTZWEBbtB")
	r.Header.Set("svix-timestamp", "1719581262")
	r.Header.Set("svix-signature", "v1,CQDy5axyjaQrsBAsyFfh8M6OIS6FMNIyEF87JKaLl/0=")

	body, err := ops.Verify(context.Background(), r, secret)
	require.NoError(t, err)

	assert.Equal(t, payload, string(body))
}
