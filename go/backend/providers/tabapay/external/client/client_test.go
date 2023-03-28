package client_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	"gitlab.com/fynbos/backend/providers/tabapay/external/client"
	"gitlab.com/fynbos/env"
)

func TestApiWithoutProxy(t *testing.T) {
	env.SetEnv(t, "local")
	if os.Getenv("TABAPAY_CLIENT_ID") == "" && os.Getenv("TABAPAY_BEARER_TOKEN") == "" {
		t.Skip("no credentials")
	}

	client, err := client.New(client.NewClientArgs{
		ClientID:    os.Getenv("TABAPAY_CLIENT_ID"),
		BearerToken: os.Getenv("TABAPAY_BEARER_TOKEN"),
	})
	require.NoError(t, err)

	ctx := context.Background()
	queryCard, err := client.QueryCard(ctx, external.QueryCardArgs{
		Card: &external.Card{
			AccountNumber: "8405840764999994", // VISA
		},
		Owner: &external.Owner{},
	})
	require.NoError(t, err)

	fmt.Printf("%+v\n", queryCard)
}
