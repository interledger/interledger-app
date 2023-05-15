package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc/ops"
	"gitlab.com/fynbos/backend/kyc/persona"
	persona_mock "gitlab.com/fynbos/backend/kyc/persona/mock"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
)

func TestGetPersonaInquiry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	uc := user_mock.NewMock()
	w, err := uc.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: uuid.NewString(),
		Name:   "test",
	})
	require.NoError(t, err)

	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, uc)

	ctrl := gomock.NewController(t)
	pc := persona_mock.NewMockClient(ctrl)

	// There is no existing inquiry or KYC data
	inqID := uuid.NewString()
	pc.EXPECT().CreateInquiry(ctx, gomock.Any(), gomock.Any()).Return(&persona.InquiryData{
		Type:       "inquiry",
		ID:         inqID,
		Attributes: persona.InquiryAttributes{Status: "pending"},
	}, nil)

	inq, err := ops.GetPersonaInquiry(ctx, b, pc, w.ID, "")
	require.NoError(t, err)

	assert.Equal(t, inq.SessionToken, "")
	assert.Equal(t, inq.ID, inqID)

	// Now lets get an update
	pc.EXPECT().ResumeInquiry(ctx, inqID, gomock.Any()).Return(&persona.InquiryData{ID: inqID, Meta: persona.InquiryMeta{SessionToken: "token"}}, nil)

	inq, err = ops.GetPersonaInquiry(ctx, b, pc, w.ID, "")
	require.NoError(t, err)

	assert.Equal(t, inq.SessionToken, "token")
	assert.Equal(t, inq.ID, inqID)
}
