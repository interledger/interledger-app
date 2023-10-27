package ops_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc/ops"
	"gitlab.com/fynbos/backend/kyc/persona"
	persona_mock "gitlab.com/fynbos/backend/kyc/persona/mock"
	signup_mock "gitlab.com/fynbos/backend/signup/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	wallet_mock "gitlab.com/fynbos/backend/wallets/client/mock"

	user_mock "gitlab.com/fynbos/backend/user/client/mock"
)

func TestGetPersonaInquiry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	uc := user_mock.NewMock()
	walletID := uuid.NewString()
	userID := uuid.NewString()
	uc.WalletUser[walletID] = userID
	wc := wallet_mock.NewMockClient(ctrl)
	walletCountry := country.US
	expectedPersonaInquiryTemplate := string(persona.GetTemplateIDForCountry(ctx, walletCountry))
	wc.EXPECT().Get(ctx, gomock.Any()).Return(&wallets.Wallet{
		ID:      walletID,
		Country: walletCountry,
	}, nil)

	sc := signup_mock.NewMockClient(ctrl)

	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, uc, sc, nil, nil, wc)
	pc := persona_mock.NewMockClient(ctrl)

	// There is no existing inquiry or KYC data
	t.Run("creates a new persona inquiry", func(st *testing.T) {
		inqID := uuid.NewString()
		pc.EXPECT().CreateInquiry(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, args persona.IndividualAttributes, idempotencyKey string) (*persona.InquiryData, error) {
			assert.Equal(t, walletCountry.String(), args.CountryCode)
			assert.Equal(t, expectedPersonaInquiryTemplate, args.InquiryTemplateID)

			return &persona.InquiryData{
				Type:       "inquiry",
				ID:         inqID,
				Attributes: persona.InquiryAttributes{Status: "pending"},
			}, nil
		})

		inq, err := ops.GetPersonaInquiry(ctx, b, pc, walletID, "")
		require.NoError(t, err)

		assert.Equal(st, inq.ID, inqID)

		inq, err = ops.GetPersonaInquiry(ctx, b, pc, walletID, "")
		require.NoError(t, err)

		assert.Equal(t, inq.ID, inqID)
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
