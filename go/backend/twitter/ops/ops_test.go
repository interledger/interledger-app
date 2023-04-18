package ops_test

import (
	"context"
	"fmt"
	"testing"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/twitter/ops"
)

func TestCreateAuthURL(t *testing.T) {
	ctx := context.Background()
	b := ops.NewTestBackends(func(tb *ops.TestBackends) {
		tb.Db = db.MigrateTestDB(t, ctx)
	})

	url, err := ops.CreateAuthURL(ctx, b, &ops.CreateAuthURLArgs{
		ClientID:     "eFpmVDNvU0dNZFg3WFJMTU94cXU6MTpjaQ",
		Scopes:       []string{"tweet.read"},
		RedirectURL:  "https://fynbos.app/twitter/callback",
		AuthEndpoint: "https://twitter.com/i/oauth2/authorize",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	fmt.Println(url)
}
