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
	signup_mock "gitlab.com/fynbos/backend/signup/client/mock"

	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
)

func TestGetPersonaInquiry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	uc := user_mock.NewMock()

	userID := uuid.NewString()
	sc := signup_mock.NewMockClient(ctrl)
	w, err := uc.CreateNewWallet(ctx, user.CreateWalletArgs{
		UserID: userID,
		Name:   "test",
	})
	require.NoError(t, err)

	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, uc, sc)
	pc := persona_mock.NewMockClient(ctrl)

	// There is no existing inquiry or KYC data
	t.Run("creates a new persona inquiry", func(st *testing.T) {
		inqID := uuid.NewString()
		pc.EXPECT().CreateInquiry(ctx, gomock.Any(), gomock.Any()).Return(&persona.InquiryData{
			Type:       "inquiry",
			ID:         inqID,
			Attributes: persona.InquiryAttributes{Status: "pending"},
		}, nil)

		inq, err := ops.GetPersonaInquiry(ctx, b, pc, w.ID, "")
		require.NoError(t, err)

		assert.Equal(st, inq.SessionToken, "")
		assert.Equal(st, inq.ID, inqID)

		// Now lets get an update
		pc.EXPECT().ResumeInquiry(ctx, inqID, gomock.Any()).Return(&persona.InquiryData{ID: inqID, Meta: persona.InquiryMeta{SessionToken: "token"}}, nil)

		inq, err = ops.GetPersonaInquiry(ctx, b, pc, w.ID, "")
		require.NoError(st, err)

		assert.Equal(st, inq.SessionToken, "token")
		assert.Equal(st, inq.ID, inqID)
	})

	t.Run("returns existing one if it is in needs_review state", func(st *testing.T) {
		inqID, walletID := uuid.NewString(), uuid.NewString()
		b.DB().MustExec("INSERT INTO kyc_persona_inquiries (external_id, state, wallet_id) VALUES ($1, $2, $3)", inqID, persona.InquiryNeedsReview, walletID)

		inq, err := ops.GetPersonaInquiry(ctx, b, pc, walletID, "")
		require.NoError(t, err)
		assert.Equal(st, inqID, inq.ID)
		assert.Equal(st, persona.InquiryNeedsReview, inq.Status)
	})
}
