package ops_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/gatehub/ops"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
)

var userID = uuid.NewString()
var walletID = uuid.NewString()
var gatehubUserID = uuid.NewString()
var totpURL = fmt.Sprintf("otpauth://totp/local.interledger.app:%s?algorithm=SHA1&digits=6&issuer=local.interledger.app&period=30&secret=EGO3DEBFSF6Q3RKNRENIQ7XT7JO76MFA", userID)
var now = time.Now()

var seed = fmt.Sprintf(`
INSERT INTO wallets (id, name) VALUES ('%s', 'testingwallet') ON CONFLICT DO NOTHING;
INSERT INTO gatehub_users (external_id, wallet_id) values ('%s', '%s') ON CONFLICT DO NOTHING;
`, walletID, gatehubUserID, walletID)

func TestHandleSCAVerification(t *testing.T) {
	ctx := context.Background()
	db := db.MigrateTestDB(t, ctx)
	users := user_mock.NewMock()
	b := Backends{
		db:    db,
		users: users,
	}

	_, err := db.ExecContext(ctx, seed)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		before func()
		after  func()

		req           ops.SCARequest
		gatehubUserID string
		expected      bool
	}{
		{
			name:          "returns false when the sca code is not present",
			req:           ops.SCARequest{Action: ops.SCAActionVerify},
			gatehubUserID: "",
			expected:      false,
		},
		{
			name:          "returns false if it cannot retrieve the wallet id for the given gatehub user id",
			req:           ops.SCARequest{Action: ops.SCAActionVerify, Code: ptr("111111")},
			gatehubUserID: "random-user-id",
			expected:      false,
		},
		{
			name: "returns false if it cannot retrieve the user id for the given wallet id",
			req:  ops.SCARequest{Action: ops.SCAActionVerify, Code: ptr("111111")},
			before: func() {
				users.MapUserWallet(ctx, userID, "another-wallet-id")
			},
			after: func() {
				users.Cleanup()
			},
			gatehubUserID: gatehubUserID,
			expected:      false,
		},
		{
			name: "returns false if it cannot retrieve the TOTP url for the given user id",
			req:  ops.SCARequest{Action: ops.SCAActionVerify, Code: ptr("111111")},
			before: func() {
				users.MapUserWallet(ctx, userID, walletID)
			},
			after: func() {
				users.Cleanup()
			},
			gatehubUserID: gatehubUserID,
			expected:      false,
		},
		{
			name: "returns false if it cannot construct the otp key from the given totp url",
			before: func() {
				users.MapUserWallet(ctx, userID, walletID)
				users.MapUserTotpURL(ctx, userID, "https://example.com")
			},
			after: func() {
				users.Cleanup()
			},
			req:           ops.SCARequest{Action: ops.SCAActionVerify, Code: ptr("111111")},
			gatehubUserID: gatehubUserID,
			expected:      false,
		},
		{
			name: "returns false for an invalid totp",
			req:  ops.SCARequest{Action: ops.SCAActionVerify, Code: ptr("111111")},
			before: func() {
				users.MapUserWallet(ctx, userID, walletID)
				users.MapUserTotpURL(ctx, userID, totpURL)
			},
			after: func() {
				users.Cleanup()
			},
			gatehubUserID: gatehubUserID,
			expected:      false,
		},
		{
			name: "returns true for a valid totp code",
			req:  ops.SCARequest{Action: ops.SCAActionVerify, Code: ptr(codeFromTOTPURL(t, totpURL, now))},
			before: func() {
				users.MapUserWallet(ctx, userID, walletID)
				users.MapUserTotpURL(ctx, userID, totpURL)
			},
			after: func() {
				users.Cleanup()
			},
			gatehubUserID: gatehubUserID,
			expected:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.before != nil {
				tc.before()
			}
			res := ops.HandleSCAVerification(ctx, b, tc.req, tc.gatehubUserID, now)
			assert.Equal(t, tc.expected, res)
			if tc.after != nil {
				tc.after()
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

func codeFromTOTPURL(t *testing.T, totpURL string, now time.Time) string {
	t.Helper()

	key, err := otp.NewKeyFromURL(totpURL)
	if err != nil {
		t.Fatalf("failed to parse totp url: %v", err)
	}

	code, err := totp.GenerateCode(key.Secret(), now)
	if err != nil {
		t.Fatalf("failed to generate totp code: %v", err)
	}

	return code
}
