package ops_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/kyc"
	kyc_mock "gitlab.com/fynbos/backend/kyc/client/mock"
	"gitlab.com/fynbos/backend/providers/astra/external"
	external_mock "gitlab.com/fynbos/backend/providers/astra/external/mock"
	"gitlab.com/fynbos/backend/providers/astra/ops"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
)

func TestCreateIntent(t *testing.T) {
	t.Skip("TODO: Fix this test, currently failing")
	ctx := context.Background()

	ctrl := gomock.NewController(t)

	ex := external_mock.NewMockClient(ctrl)
	ky := kyc_mock.NewMockClient(ctrl)
	um := user_mock.NewMock()

	b := ops.NewTestBackends(t, func(tb *ops.TestBackends) {
		tb.DBC = db.MigrateTestDB(t, ctx)
		tb.Extr = ex
		tb.Ky = ky
		tb.Uc = um
	})

	walletID := uuid.NewString()
	intentID := uuid.NewString()

	um.MapUserWallet(ctx, uuid.NewString(), walletID)

	ky.EXPECT().GetPersonaIDNumbers(ctx, walletID).Return(&kyc.PersonaIDNumbers{
		SocialSecurity: "123-0974-0982",
	}, nil).AnyTimes()
	ky.EXPECT().GetIndividualDetails(ctx, walletID).Return(&kyc.IndividualDetails{
		WalletID:     walletID,
		FirstName:    "Bobby",
		LastName:     "Tables",
		CountryCode:  "US",
		PlaceOfBirth: "US",
		Nationality:  "US",
		Gender:       kyc.GenderMale,
		DateOfBirth:  time.Date(1998, time.January, 3, 0, 0, 0, 0, time.UTC),
		Address: &kyc.Address{
			Line1:       "Death Star",
			Line2:       "Degoba System",
			Building:    "",
			Apartment:   "Caption",
			City:        "Mos Eysly",
			State:       "US-WC",
			ZipCode:     "209",
			CountryCode: "US",
		},
		IPAddress: "217.0.0.1",
	}, nil).AnyTimes()

	ex.EXPECT().CreateIntent(ctx, external.CreateIntentReq{
		Email:          "info@interledger.app",
		Phone:          "+27836321959",
		FirstName:      "Bobby",
		LastName:       "Tables",
		Address1:       "Death Star",
		Address2:       "Degoba System",
		City:           "Mos Eysly",
		State:          "WC",
		PostalCode:     "209",
		DateOfBirth:    "1998-01-03",
		SocialSecurity: "12309740982",
		IPAddress:      "217.0.0.1",
	}).Return(intentID, nil).AnyTimes()

	err := ops.CreateIntent(ctx, b, walletID)
	require.NoError(t, err)

	var status string
	err = b.DB().GetContext(ctx, &status, "SELECT status FROM astra_user_intents WHERE wallet_id=$1 AND intent_id=$2", walletID, intentID)
	require.NoError(t, err)

	assert.Equal(t, "unknown", status)
}
