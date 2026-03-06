package ops_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/gatehub/ops"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
)

var userID = uuid.NewString()
var walletID = uuid.NewString()
var gatehubUserID = uuid.NewString()

var seed = fmt.Sprintf(`
INSERT INTO wallets (id, name) VALUES ('%s', 'testingwallet') ON CONFLICT DO NOTHING;
INSERT INTO user_wallets (user_id, wallet_id) values ('%s', '%s') ON CONFLICT DO NOTHING;
INSERT INTO gatehub_users (external_id, wallet_id) values ('%s', '%s') ON CONFLICT DO NOTHING;
`, walletID, userID, walletID, gatehubUserID, walletID)

func TestHandleSCAVerification(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	b := Backends{
		db:    db,
		users: user_mock.NewMock(),
	}
	_, err := db.ExecContext(ctx, seed)
	if err != nil {
		t.Fatal(err)
	}

	// func HandleSCAVerification(ctx context.Context, b Backends, req SCARequest, gatehubUserID string) bool {
	cases := []struct {
		name          string
		req           ops.SCARequest
		gatehubUserID string
		expected      bool
	}{
		{
			name:          "returns false when the sca code is not present",
			req:           ops.SCARequest{Action: ops.SCAActionVerify},
			gatehubUserID: gatehubUserID,
			expected:      false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := ops.HandleSCAVerification(ctx, b, tc.req, tc.gatehubUserID)
			assert.Equal(t, tc.expected, res)
		})
	}
}
