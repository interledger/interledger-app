package ops_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
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
	t.Run("uses correct template for wallet country", func(st *testing.T) {
		assert.Equal(st, "itmpl_p9xFdzFCVrRSv7zbPUtbHLfSRFZQ", string(persona.GetTemplateIDForCountry(ctx, country.ZA)))
		assert.Equal(st, "itmpl_JYHP6J5MtyaSd9UzcZRB2KBiPKCo", string(persona.GetTemplateIDForCountry(ctx, country.US)))
		assert.Equal(st, "itmpl_wAUZFdNhtuoQUoqf5uQmisEmX4qh", string(persona.GetTemplateIDForCountry(ctx, country.GB)))
	})
}

func TestGetApprovedPersonaInquiryURL(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	uc := user_mock.NewMock()
	walletID := uuid.NewString()
	userID := uuid.NewString()
	uc.WalletUser[walletID] = userID
	wc := wallet_mock.NewMockClient(ctrl)
	sc := signup_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, uc, sc, nil, nil, wc)

	inquiryID := uuid.NewString()
	b.DB().MustExec("INSERT INTO kyc_persona_inquiries (external_id, wallet_id, state) VALUES ($1,$2,$3);", inquiryID, walletID, persona.InquiryFailed)

	// Must fail as it is not approved
	_, err := ops.GetApprovedPersonaInquiryURL(context.Background(), b, uuid.NewString())
	assert.ErrorIs(t, err, kyc.ErrNoKYCInfo)

	// update state to check that only approved ones are returned
	b.DB().MustExec("UPDATE kyc_persona_inquiries SET state=$1 WHERE external_id=$2;", persona.InquiryApproved, inquiryID)

	inquiryURL, err := ops.GetApprovedPersonaInquiryURL(context.Background(), b, walletID)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("https://app.withpersona.com/dashboard/inquiries/%s", inquiryID), inquiryURL)

	_, err = ops.GetApprovedPersonaInquiryURL(context.Background(), b, uuid.NewString())
	assert.ErrorIs(t, err, kyc.ErrNoKYCInfo)
}

func TestGetZAIDNumber(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	uc := user_mock.NewMock()
	wc := wallet_mock.NewMockClient(ctrl)
	sc := signup_mock.NewMockClient(ctrl)
	b := ops.NewTestBackends(t, db.MigrateTestDB(t, ctx), nil, uc, sc, nil, nil, wc)
	pc := persona_mock.NewMockClient(ctrl)

	zaIDPattern := regexp.MustCompile(`^\d{13}$`)

	t.Run("fake mode returns a 13-digit ZA ID without calling Persona", func(t *testing.T) {
		// pc has no EXPECT calls — any call would fail the test
		for range 10 {
			idNum, err := ops.GetZAIDNumber(ctx, b, pc, uuid.NewString(), true)
			require.NoError(t, err)
			assert.Regexp(t, zaIDPattern, idNum, "expected 13-digit ZA ID number")
		}
	})

	t.Run("returns ErrNoKYCInfo when no persona account record exists", func(t *testing.T) {
		_, err := ops.GetZAIDNumber(ctx, b, pc, uuid.NewString(), false)
		assert.ErrorIs(t, err, kyc.ErrNoKYCInfo)
	})

	t.Run("returns ZA ID fetched from Persona account", func(t *testing.T) {
		walletID, accID := uuid.NewString(), uuid.NewString()
		b.DB().MustExec("INSERT INTO kyc_persona_accounts (external_id, wallet_id, updated_at) VALUES ($1,$2,now())", accID, walletID)

		pc.EXPECT().GetAccount(ctx, accID).Return(&persona.AccountData{
			Attributes: persona.IndividualAttributes{
				IdentificationNumbers: map[string][]persona.IdentificationNumber{
					"pp": {{IssuingCountry: "ZA", IdentificationNumber: "8406270000087"}},
				},
			},
		}, nil)

		idNum, err := ops.GetZAIDNumber(ctx, b, pc, walletID, false)
		require.NoError(t, err)
		assert.Equal(t, "8406270000087", idNum)
	})

	t.Run("returns ErrNoKYCInfo when Persona account has no ZA ID", func(t *testing.T) {
		walletID, accID := uuid.NewString(), uuid.NewString()
		b.DB().MustExec("INSERT INTO kyc_persona_accounts (external_id, wallet_id, updated_at) VALUES ($1,$2,now())", accID, walletID)

		pc.EXPECT().GetAccount(ctx, accID).Return(&persona.AccountData{
			Attributes: persona.IndividualAttributes{
				IdentificationNumbers: map[string][]persona.IdentificationNumber{
					"pp": {{IssuingCountry: "US", IdentificationNumber: "123456789"}},
				},
			},
		}, nil)

		_, err := ops.GetZAIDNumber(ctx, b, pc, walletID, false)
		assert.ErrorIs(t, err, kyc.ErrNoKYCInfo)
	})
}
