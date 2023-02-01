package ops_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/fynbos/backend/authorisation"
	"gitlab.com/fynbos/backend/authorisation/ops"
	"gitlab.com/fynbos/backend/db"
)

func TestCreateGrant(t *testing.T) {

	ctx := context.Background()
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx))

	err := ops.CreateClient(ctx, b, "https://fynbos.me/alice")
	require.NoError(t, err)

	g, err := ops.CreateGrant(ctx, b, authorisation.GrantRequest{
		AccessToken: []authorisation.AccessTokenReq{{
			Access: []authorisation.Access{{
				Type:      "incoming-payment",
				Actions:   []string{"create", "read"},
				Locations: []string{"https://fynbos.me/bobby"},
				Datatypes: []string{""},
			}},
			Label: "TestAccess1",
		}},
		Client: authorisation.ClientReq{
			Display: authorisation.Display{Name: "Not Used", URI: "https://fynbos.me/alice"},
			Key: authorisation.Key{
				Proof: "TODO",
			},
		},
	})
	require.NoError(t, err)

	fmt.Println(g)
}
