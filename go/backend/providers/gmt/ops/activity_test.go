package ops

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"gitlab.com/fynbos/backend/db"
)

func TestActivity_UpdateSendRecvUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := NewTestBackends(db.MigrateTestDB(t, ctx))

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.UpdateSendRecvUser)

	sWallet := newWallet(t, b.DB())
	rWallet := newWallet(t, b.DB())
	cr := ComplianceResp{
		SenderID:         23,
		SenderWalletID:   sWallet,
		ReceiverID:       43,
		ReceiverWalletID: rWallet,
	}
	_, err := env.ExecuteActivity(a.UpdateSendRecvUser, cr)
	require.NoError(t, err)

	sid, err := getSenderID(ctx, b, sWallet)
	require.NoError(t, err)
	assert.Equal(t, int64(23), sid)

	sid, err = getSenderID(ctx, b, rWallet)
	require.NoError(t, err)
	assert.Equal(t, int64(0), sid)

	rid, err := getReceiverID(ctx, b, rWallet)
	require.NoError(t, err)
	assert.Equal(t, int64(43), rid)

	rid, err = getReceiverID(ctx, b, sWallet)
	require.NoError(t, err)
	assert.Equal(t, int64(0), rid)
}

func newWallet(t *testing.T, db *sqlx.DB) string {
	walletID := uuid.NewString()
	_, err := db.Exec(
		"INSERT INTO wallets (id, name) VALUES ($1, $2);",
		walletID,
		"test",
	)
	require.NoError(t, err)

	return walletID
}
