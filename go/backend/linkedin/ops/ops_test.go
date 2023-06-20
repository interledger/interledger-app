package ops_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedin/ops"
	"testing"
)

func TestCreateAuthURL(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(func(tb *ops.TestBackends) {
		tb.Db = db.MigrateTestDB(t, ctx)
	})

	url, err := ops.CreateAuthURL(ctx, b, &ops.CreateAuthURLArgs{
		ClientID:     uuid.NewString(),
		WalletID:     uuid.NewString(),
		Scopes:       []string{"test.scope"},
		RedirectURL:  "https://fynbos.app/linkedin/callback",
		AuthEndpoint: "https://www.linkedin.com/oauth/v2/authorization",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	fmt.Println(url)
}
