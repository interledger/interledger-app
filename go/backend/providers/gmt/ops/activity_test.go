package ops

import (
	"context"
	"testing"

	"gitlab.com/fynbos/backend/user"
	users_mock "gitlab.com/fynbos/backend/user/client/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"go.temporal.io/sdk/testsuite"
)

func TestActivity_UpdateSendRecvUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := NewTestBackends(t, db.MigrateTestDB(t, ctx))
	uc := users_mock.NewMock()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.UpdateSendRecvUser)

	sWallet, err := uc.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test_sender",
	})
	require.NoError(t, err)

	rWallet, err := uc.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test_receiver",
	})
	require.NoError(t, err)

	sid, err := getSenderID(ctx, b, sWallet.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), sid)

	rid, err := getReceiverID(ctx, b, rWallet.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), rid)

	cr := ComplianceResp{
		SenderID:         23,
		SenderWalletID:   sWallet.ID,
		ReceiverID:       43,
		ReceiverWalletID: rWallet.ID,
	}
	_, err = env.ExecuteActivity(a.UpdateSendRecvUser, cr)
	require.NoError(t, err)

	sid, err = getSenderID(ctx, b, sWallet.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(23), sid)

	sid, err = getSenderID(ctx, b, rWallet.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), sid)

	rid, err = getReceiverID(ctx, b, rWallet.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(43), rid)

	rid, err = getReceiverID(ctx, b, sWallet.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), rid)
}

func TestActivity_SaveReceipt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := NewTestBackends(t, db.MigrateTestDB(t, ctx))

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.SaveReceipt)

	tr := TransactionResp{
		ID:         "some_id_for_us",
		ReceiptRef: "some_id_for_gmt",
		Status:     "created",
		Licence:    "USA_will_let_me",
		RTR:        "You have not right to a refund, NO!!",
		ErrorMsg:   "",
		Contact:    "Call us on XXXX-XXX-XXX for any queries",
	}
	_, err := env.ExecuteActivity(a.SaveReceipt, tr)
	require.NoError(t, err)

	var res struct {
		ID         string `db:"external_id"`
		ReceiptRef string `db:"receipt"`
		Licence    string `db:"licence"`
		RTR        string `db:"right_to_return"`
		ErrMsg     string `db:"error_msg"`
		Contact    string `db:"contact"`
	}
	err = b.DB().GetContext(ctx, &res, "select external_id, receipt, licence, right_to_return, error_msg, contact from gmt_receipts")
	require.NoError(t, err)

	assert.Equal(t, tr.ID, res.ID)
	assert.Equal(t, tr.ReceiptRef, res.ReceiptRef)
	assert.Equal(t, tr.Licence, res.Licence)
	assert.Equal(t, tr.RTR, res.RTR)
	assert.Equal(t, tr.ErrorMsg, res.ErrMsg)
	assert.Equal(t, tr.Contact, res.Contact)
}

func TestActivity_CreateWorkflowRef(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b := NewTestBackends(t, db.MigrateTestDB(t, ctx))

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	a := NewActivity(b)
	env.RegisterActivity(a.CreateWorkflowRef)

	args := CreateWorkflowRefArgs{
		ExternalID:    "123456789",
		WorkflowID:    uuid.NewString(),
		WorkflowRunID: uuid.NewString(),
		ActivityName:  "Test_Ref",
	}
	res, err := env.ExecuteActivity(a.CreateWorkflowRef, args)
	require.NoError(t, err)

	var refID string
	err = res.Get(&refID)
	require.NoError(t, err)
	assert.NotEmpty(t, refID)

	var ref WorkflowRef
	err = a.b.DB().GetContext(ctx, &ref, "SELECT external_id, workflow_id, workflow_run_id  FROM  gmt_workflow_refs WHERE id=$1", refID)
	require.NoError(t, err)

	assert.Equal(t, args.ExternalID, ref.ExternalID)
	assert.Equal(t, args.WorkflowID, ref.WorkflowID)
	assert.Equal(t, args.WorkflowRunID, ref.WorkflowRunID)
}
