package ops_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/providers/gmt/ops"
)

func TestActivity_UpdateSendRecvUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := ops.NewTestBackends(db.MigrateTestDB(t, ctx))

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := ops.NewActivity(b)
	env.RegisterActivity(a.UpdateSendRecvUser)

	sWallet := newWallet(t, b.DB())
	rWallet := newWallet(t, b.DB())
	cr := ops.ComplianceResp{
		SenderID:         23,
		SenderWalletID:   sWallet,
		ReceiverID:       43,
		ReceiverWalletID: rWallet,
	}
	_, err := env.ExecuteActivity(a.UpdateSendRecvUser, cr)
	require.NoError(t, err)
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
