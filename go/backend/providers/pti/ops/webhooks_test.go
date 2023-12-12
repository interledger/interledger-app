package ops_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/providers/pti/ops"
	"gopkg.in/stretchr/testify.v1/require"
)

func TestWebhook(t *testing.T) {
	ctx := context.Background()
	k, err := ops.ParsePTIPublicKey()
	require.NoError(t, err)
	assert.Equal(t, "861debeb-98ad-4f9a-a144-351e18093ea9", k.KeyID())

	b := NewBackends(t)
	a := ops.NewActivity(b)
	externalID, walletID := uuid.NewString(), uuid.NewString()
	usr, err := a.SavePtiUser(ctx, externalID, walletID)
	require.NoError(t, err)
	assert.Empty(t, usr.Status)
	assert.Empty(t, usr.AssessmentStatus)

	userStatusWebhook := ops.UserWebhook{
		UserId: externalID,
		Status: "BLOCKED",
	}
	rawWebhook, err := json.Marshal(userStatusWebhook)
	require.NoError(t, err)
	err = ops.HandleUserUpdate(context.Background(), b, rawWebhook)
	require.NoError(t, err)

	usr, err = a.GetPtiUser(ctx, walletID)
	require.NoError(t, err)
	assert.Equal(t, "BLOCKED", usr.Status)

	assessmentStatusWebhook := ops.AssessmentWebhook{
		UserId:     externalID,
		Assessment: "REFUSED",
	}
	rawWebhook, err = json.Marshal(assessmentStatusWebhook)
	require.NoError(t, err)
	err = ops.HandleAssessmentUpdate(context.Background(), b, rawWebhook)
	require.NoError(t, err)

	usr, err = a.GetPtiUser(ctx, walletID)
	require.NoError(t, err)
	assert.Equal(t, "REFUSED", usr.AssessmentStatus)
}
