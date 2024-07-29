package external_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"gitlab.com/fynbos/backend/providers/chimoney/external"
	"gitlab.com/fynbos/env"
	"gopkg.in/stretchr/testify.v1/require"
)

func TestVerifyPaymentLink(t *testing.T) {
	env.SetEnv(t, "local")
	if os.Getenv("CHIMONEY_TOKEN") == "" {
		t.SkipNow()
	}

	c := external.New(nil)
	p, err := c.VerifyPayment(context.Background(), external.VerifyPaymentReq{
		IssueID:   "203757a3-6980-4af1-bf66-bbeaa8de62fc_10.00_1722253969171",
		ChiWallet: "203757a3-6980-4af1-bf66-bbeaa8de62fc",
	})
	require.NoError(t, err)

	j, err := json.MarshalIndent(p, "", " ")
	require.NoError(t, err)

	fmt.Printf("payment: %s\n", string(j))
}
