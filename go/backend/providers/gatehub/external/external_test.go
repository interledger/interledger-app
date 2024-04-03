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

	token, err := c.IssueToken(context.Background(), "", external.Onboarding)
	require.NoError(t, err)

	fmt.Printf("token: %+v", token)
}
