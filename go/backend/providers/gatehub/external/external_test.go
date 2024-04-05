package external_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/gatehub/external"
	"gitlab.com/fynbos/env"
)

func TestClient(t *testing.T) {
	env.SetEnv(t, "local")
	if os.Getenv("GATEHUB_APP_ID") == "" || os.Getenv("GATEHUB_SECRET") == "" {
		t.SkipNow()
	}
	c := external.NewClient(os.Getenv("GATEHUB_APP_ID"), os.Getenv("GATEHUB_SECRET"), nil)

	externalUserID := "c5b40ec1-b450-451b-b8c0-1b45f94b9db0"

	_, err := c.IssueToken(context.Background(), externalUserID, external.Onboarding)
	require.NoError(t, err)

	userWallets, err := c.GetUserWallets(context.Background(), externalUserID)
	require.NoError(t, err)
	fmt.Printf("userWallets %+v\n", userWallets)
}
