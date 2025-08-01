package cmd_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/cli/cmd"
)

func TestGetPaymentPointer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		op := cmd.OutgoingPayment{
			ID:               "testop",
			PaymentPointer:   "https://local.ilp.link/jimmy",
			ToPaymentPointer: "https://local.ilp.link/janey",
			SendAmount: cmd.Amount{
				Amount:   10,
				Currency: "USD",
			},
		}

		err := json.NewEncoder(w).Encode(op)
		require.NoError(t, err)
	}))
	t.Cleanup(func() {
		srv.Close()
	})

	outgoingPaymentID := srv.URL
	cmd := cmd.NewPaymentPointerCmd(NewTestBackends(t))
	cmd.SetArgs([]string{outgoingPaymentID})

	err := cmd.Execute()
	require.NoError(t, err)
}
